package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"mendo/internal/closing"
	"mendo/internal/di"
	"mendo/internal/infrastructure/eventbus"
	"mendo/internal/staff"
)

// Run は HTTP サーバーを起動する。
func Run(app *di.App, bus *eventbus.WatermillEventBus) {
	// --- Staff（アクティブレコード。第10章）---
	staffStore := staff.NewStore()
	staffHandler := staff.NewHandler(staffStore)

	// --- 閉店処理（トランザクションスクリプト。第10章）---
	closingHandler := closing.NewHandler()

	// インポートワーカーを起動
	app.ImportWorker.Start(context.Background())

	// ルーティング
	mux := http.NewServeMux()
	RegisterRoutes(mux, app, staffHandler, closingHandler)

	// イベント購読
	RegisterSubscribers(bus, app)

	// サーバー設定
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	fmt.Println("mendo server starting on :8080")
	log.Fatal(srv.ListenAndServe())
}
