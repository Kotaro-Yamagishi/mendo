package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mendo/internal/closing"
	"mendo/internal/di"
	"mendo/internal/infrastructure/eventbus"
	"mendo/internal/interface/handler"
	"mendo/internal/logging"
	"mendo/internal/staff"
)

// Run は HTTP サーバーを起動し、シグナルを受けて Graceful Shutdown する。
func Run(app *di.App, bus *eventbus.WatermillEventBus) {
	logger := app.Logger

	// --- Staff（アクティブレコード。第10章）---
	staffStore := staff.NewStore()
	staffHandler := staff.NewHandler(staffStore)

	// --- 閉店処理（トランザクションスクリプト。第10章）---
	closingHandler := closing.NewHandler()

	// --- ヘルスチェック ---
	healthHandler := handler.NewHealthHandler()

	// バックグラウンドワーカー用の context（キャンセル可能）
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	// インポートワーカーを起動
	app.ImportWorker.Start(workerCtx)

	// Outbox リレーを起動
	app.OutboxRelay.Start(workerCtx)

	// ルーティング
	mux := http.NewServeMux()
	RegisterRoutes(mux, app, staffHandler, closingHandler, logger, healthHandler)

	// イベント購読
	RegisterSubscribers(bus, app)

	// CorrelationMiddleware でラップ
	httpHandler := logging.CorrelationMiddleware(mux, logger)

	// サーバー設定
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      httpHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// シグナル受信用チャネル
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	// サーバーをゴルーチンで起動
	go func() {
		healthHandler.SetReady()
		logger.Info("server starting", slog.String("addr", ":8080"))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server listen error", "error", err)
		}
	}()

	// シグナルを待つ
	sig := <-sigCh
	logger.Info("shutdown signal received", slog.String("signal", sig.String()))

	// --- Graceful Shutdown 開始 ---
	healthHandler.SetNotReady() // Readiness を落とす（LB がトラフィックを止める）

	// 1. HTTP サーバーを Graceful Shutdown（処理中のリクエストを待つ）
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", "error", err)
	} else {
		logger.Info("server shutdown completed")
	}

	// 2. バックグラウンドワーカーを停止
	workerCancel()
	logger.Info("background workers stopped")

	logger.Info("process exiting")
}
