//go:build integration

package e2e_test

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"

	kitchencommand "mendo/internal/application/command/kitchen"
	ordercommand "mendo/internal/application/command/order"
	socommand "mendo/internal/application/command/specialorder"
	orderquery "mendo/internal/application/query/order"
	"mendo/internal/domain"
	"mendo/internal/domain/kitchen"
	"mendo/internal/domain/order"
	"mendo/internal/infrastructure/eventbus"
	dsmysql "mendo/internal/infrastructure/datasource/mysql"
	infraMysql "mendo/internal/infrastructure/mysql"
	"mendo/internal/infrastructure/repository"
	"mendo/internal/interface/handler"
)

var testDB *sql.DB
var testServer *httptest.Server

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		dsn = "root:test@tcp(localhost:13306)/mendo_test?charset=utf8mb4&parseTime=true&loc=UTC"
	}

	var err error
	testDB, err = connectWithRetry(dsn, 10, 2*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to test DB: %v\n", err)
		fmt.Fprintf(os.Stderr, "hint: docker run -d --name mendo-e2e -p 13306:3306 -e MYSQL_ROOT_PASSWORD=test -e MYSQL_DATABASE=mendo_test mysql:8.0\n")
		os.Exit(1)
	}

	if err := runMigrations(testDB); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	if err := runSeed(testDB); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run seed: %v\n", err)
		os.Exit(1)
	}

	testServer = setupServer()

	code := m.Run()
	testServer.Close()
	testDB.Close()
	os.Exit(code)
}

// setupServer はテスト用の HTTP サーバーを手動配線で構築する。
// Wire（DI）は使わず、MySQL datasource → repository → usecase → handler の順に組み立てる。
func setupServer() *httptest.Server {
	// --- インフラ層: datasource ---
	eventStoreDS := dsmysql.NewMySQLEventStoreDataSource(testDB)
	outboxDS := dsmysql.NewMySQLOutboxDataSource(testDB)
	kitchenDS := dsmysql.NewMySQLKitchenDataSource(testDB)
	specialOrderDS := dsmysql.NewMySQLSpecialOrderDataSource(testDB)

	// --- インフラ層: repository ---
	eventStoreRepo := repository.NewEventStoreRepository(eventStoreDS)
	outboxRepo := repository.NewOutboxRepository(outboxDS)
	kitchenRepo := repository.NewKitchenRepository(kitchenDS)
	specialOrderRepo := repository.NewSpecialOrderRepository(specialOrderDS)

	// --- インフラ層: TransactionManager（MySQL トランザクション）---
	txManager := infraMysql.NewMySQLTransactionManager(testDB)

	// --- インフラ層: InMemory EventBus（イベント配信はインメモリ）---
	dlq := repository.NewInMemoryDLQ()
	bus := eventbus.NewWatermillEventBus(dlq, 3, slog.Default())

	// --- インフラ層: InMemory OrderStateStore（Projection）---
	orderStateStore := repository.NewInMemoryOrderStateStore()

	// --- 厨房 ID ---
	kitchenID := kitchen.KitchenID("kitchen-1")

	// --- アプリケーション層: オーダー系ユースケース ---
	createOrderUC := ordercommand.NewCreateOrderUsecase(eventStoreRepo, outboxRepo, bus)
	confirmOrderUC := ordercommand.NewConfirmOrderESUsecase(eventStoreRepo, outboxRepo, bus)
	cancelOrderUC := ordercommand.NewCancelOrderUsecase(eventStoreRepo, outboxRepo)
	completeCookingUC := kitchencommand.NewCompleteCookingUsecase(kitchenRepo, kitchenRepo, outboxRepo, txManager, kitchenID)
	startCookingUC := kitchencommand.NewStartCookingUsecase(kitchenRepo, kitchenRepo, bus, kitchenID)
	listOrdersUC := orderquery.NewListOrdersUsecase(orderStateStore)

	// --- アプリケーション層: 特別注文系ユースケース ---
	createSpecialOrderUC := socommand.NewCreateSpecialOrderUsecase(specialOrderRepo, bus)
	approveSpecialOrderUC := socommand.NewApproveSpecialOrderUsecase(specialOrderRepo, specialOrderRepo, bus)
	rejectSpecialOrderUC := socommand.NewRejectSpecialOrderUsecase(specialOrderRepo, specialOrderRepo, bus)
	resubmitSpecialOrderUC := socommand.NewResubmitSpecialOrderUsecase(specialOrderRepo, specialOrderRepo, bus)

	// --- イベント購読: OrderCreated → Projection 更新 ---
	bus.Subscribe(order.EventTypeOrderCreated, func(ctx context.Context, event domain.Event) error {
		return orderStateStore.HandleEvent(ctx, event)
	})
	// --- イベント購読: ItemAdded → Projection 更新 ---
	bus.Subscribe(order.EventTypeItemAdded, func(ctx context.Context, event domain.Event) error {
		return orderStateStore.HandleEvent(ctx, event)
	})
	// --- イベント購読: OrderConfirmed → Projection 更新 + 調理開始 ---
	bus.Subscribe(order.EventTypeOrderConfirmed, func(ctx context.Context, event domain.Event) error {
		if err := orderStateStore.HandleEvent(ctx, event); err != nil {
			return err
		}
		confirmed, ok := event.(order.OrderConfirmed)
		if !ok {
			return fmt.Errorf("unexpected event type: %T", event)
		}
		publicEvent := order.ToPublicConfirmed(confirmed)
		return startCookingUC.HandleOrderConfirmedPublic(ctx, publicEvent)
	})

	// --- インターフェース層: ハンドラ ---
	orderHandler := handler.NewOrderHandler(createOrderUC, confirmOrderUC, cancelOrderUC, nil, listOrdersUC)
	kitchenHandler := handler.NewKitchenHandler(completeCookingUC)
	specialOrderHandler := handler.NewSpecialOrderHandler(createSpecialOrderUC, approveSpecialOrderUC, rejectSpecialOrderUC, resubmitSpecialOrderUC)

	// --- ルーティング ---
	logger := slog.Default()
	wrap := func(h handler.AppHandlerFunc) http.HandlerFunc {
		return handler.ErrorMiddleware(h, logger)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /orders", wrap(orderHandler.HandleCreate))
	mux.HandleFunc("POST /orders/{id}/confirm", wrap(orderHandler.HandleConfirm))
	mux.HandleFunc("POST /orders/{id}/cancel", wrap(orderHandler.HandleCancel))
	mux.HandleFunc("POST /kitchen/complete", wrap(kitchenHandler.HandleCompleteCooking))
	mux.HandleFunc("POST /special-orders", wrap(specialOrderHandler.HandleCreate))
	mux.HandleFunc("POST /special-orders/{id}/approve", wrap(specialOrderHandler.HandleApprove))
	mux.HandleFunc("POST /special-orders/{id}/reject", wrap(specialOrderHandler.HandleReject))
	mux.HandleFunc("POST /special-orders/{id}/resubmit", wrap(specialOrderHandler.HandleResubmit))

	return httptest.NewServer(mux)
}

func connectWithRetry(dsn string, maxRetries int, interval time.Duration) (*sql.DB, error) {
	var db *sql.DB
	var err error
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			time.Sleep(interval)
			continue
		}
		if err = db.Ping(); err == nil {
			return db, nil
		}
		db.Close()
		time.Sleep(interval)
	}
	return nil, fmt.Errorf("failed to connect after %d retries: %w", maxRetries, err)
}

func projectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	// filename = /path/to/mendo/internal/e2e/testmain_test.go
	// 2階層上がると mendo/ ルートになる
	return filepath.Join(filepath.Dir(filename), "..", "..")
}

func runMigrations(db *sql.DB) error {
	migrationsDir := filepath.Join(projectRoot(), "migrations")
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)
	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", f, err)
		}
		statements := strings.Split(string(content), ";")
		for _, stmt := range statements {
			// 各行からコメントを除去
			lines := strings.Split(stmt, "\n")
			var cleaned []string
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed == "" || strings.HasPrefix(trimmed, "--") {
					continue
				}
				cleaned = append(cleaned, line)
			}
			cleanedStmt := strings.TrimSpace(strings.Join(cleaned, "\n"))
			if cleanedStmt == "" {
				continue
			}
			if _, err := db.Exec(cleanedStmt); err != nil {
				return fmt.Errorf("exec migration %s: %w\nStatement: %s", f, err, cleanedStmt)
			}
		}
	}
	return nil
}

// runSeed は migrations/seed.sql を読み込んで実行する。
func runSeed(db *sql.DB) error {
	seedFile := filepath.Join(projectRoot(), "migrations", "seed.sql")
	content, err := os.ReadFile(seedFile)
	if err != nil {
		return fmt.Errorf("read seed file: %w", err)
	}
	statements := strings.Split(string(content), ";")
	for _, stmt := range statements {
		lines := strings.Split(stmt, "\n")
		var cleaned []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "--") {
				continue
			}
			cleaned = append(cleaned, line)
		}
		cleanedStmt := strings.TrimSpace(strings.Join(cleaned, "\n"))
		if cleanedStmt == "" {
			continue
		}
		if _, err := db.Exec(cleanedStmt); err != nil {
			return fmt.Errorf("exec seed: %w\nStatement: %s", err, cleanedStmt)
		}
	}
	return nil
}
