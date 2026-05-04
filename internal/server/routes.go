package server

import (
	"net/http"

	"mendo/internal/closing"
	"mendo/internal/di"
	"mendo/internal/staff"
)

// RegisterRoutes は HTTP ルーティングを設定する。
func RegisterRoutes(mux *http.ServeMux, app *di.App, staffHandler *staff.Handler, closingHandler *closing.Handler) { //nolint:funlen // ルーティング登録は一覧性を優先して1関数にまとめる
	// 注文関連
	mux.HandleFunc("POST /orders", app.OrderHandler.HandleCreate)
	mux.HandleFunc("POST /orders/{id}/confirm", app.OrderHandler.HandleConfirm)
	mux.HandleFunc("POST /orders/{id}/cancel", app.OrderHandler.HandleCancel)
	mux.HandleFunc("GET /wait-time", app.OrderHandler.HandleWaitTime)

	// 注文一覧・詳細（Projection から読み取り）
	mux.HandleFunc("GET /orders", app.OrderHandler.HandleList)
	mux.HandleFunc("GET /orders/{id}", app.OrderHandler.HandleGetByID)

	// 厨房関連
	mux.HandleFunc("POST /kitchen/complete", app.KitchenHandler.HandleCompleteCooking)

	// メニュー関連
	mux.HandleFunc("POST /menus/{id}/soldout", app.MenuHandler.HandleSoldOut)

	// DLQ 管理
	mux.HandleFunc("GET /admin/dlq", app.DLQHandler.HandleList)
	mux.HandleFunc("POST /admin/dlq/{id}/retry", app.DLQHandler.HandleRetry)

	// 特別注文関連（プロセスマネージャー）
	mux.HandleFunc("POST /special-orders", app.SpecialOrderHandler.HandleCreate)
	mux.HandleFunc("POST /special-orders/{id}/approve", app.SpecialOrderHandler.HandleApprove)
	mux.HandleFunc("POST /special-orders/{id}/reject", app.SpecialOrderHandler.HandleReject)
	mux.HandleFunc("POST /special-orders/{id}/resubmit", app.SpecialOrderHandler.HandleResubmit)

	// スタッフ管理（アクティブレコード。第10章）
	staffHandler.RegisterRoutes(mux)

	// 閉店処理（トランザクションスクリプト。第10章）
	closingHandler.RegisterRoutes(mux)

	// インポート関連（非同期バッチ）
	mux.HandleFunc("POST /admin/import/menus", app.ImportHandler.HandleImportMenus)
	mux.HandleFunc("GET /admin/import/{id}/status", app.ImportHandler.HandleImportStatus)
}
