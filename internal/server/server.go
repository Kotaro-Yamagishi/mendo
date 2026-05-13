package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"mendo/internal/closing"
	"mendo/internal/di"
	"mendo/internal/infrastructure/eventbus"
	"mendo/internal/logging"
	"mendo/internal/staff"
)

// Run は HTTP サーバーを起動する。
func Run(app *di.App, bus *eventbus.WatermillEventBus) {
	logger := app.Logger
	// --- Staff（アクティブレコード。第10章）---
	staffStore := staff.NewStore()
	staffHandler := staff.NewHandler(staffStore)

	// --- 閉店処理（トランザクションスクリプト。第10章）---
	closingHandler := closing.NewHandler()

	// インポートワーカーを起動
	app.ImportWorker.Start(context.Background())

	// ルーティング
	mux := http.NewServeMux()
	RegisterRoutes(mux, app, staffHandler, closingHandler, logger)

	// イベント購読
	RegisterSubscribers(bus, app)

	// CorrelationMiddleware でラップ
	handler := logging.CorrelationMiddleware(mux, logger)

	// サーバー設定
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	logger.Info("server starting", slog.String("addr", ":8080"))
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("server stopped", "error", err)
	}
}
