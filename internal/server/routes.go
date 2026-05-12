package server

import (
	"log/slog"
	"net/http"

	"mendo/internal/closing"
	"mendo/internal/di"
	"mendo/internal/interface/handler"
	"mendo/internal/staff"
)

// RegisterRoutes は HTTP ルーティングを設定する。
func RegisterRoutes(mux *http.ServeMux, app *di.App, staffHandler *staff.Handler, closingHandler *closing.Handler) { //nolint:funlen // ルーティング登録は一覧性を優先して1関数にまとめる
	logger := slog.Default()
	wrap := func(h handler.AppHandlerFunc) http.HandlerFunc {
		return handler.ErrorMiddleware(h, logger)
	}

	// 注文関連
	mux.HandleFunc("POST /orders", wrap(app.OrderHandler.HandleCreate))
	mux.HandleFunc("POST /orders/{id}/confirm", wrap(app.OrderHandler.HandleConfirm))
	mux.HandleFunc("POST /orders/{id}/cancel", wrap(app.OrderHandler.HandleCancel))
	mux.HandleFunc("GET /wait-time", wrap(app.OrderHandler.HandleWaitTime))

	// 注文一覧・詳細（Projection から読み取り）
	mux.HandleFunc("GET /orders", wrap(app.OrderHandler.HandleList))
	mux.HandleFunc("GET /orders/{id}", wrap(app.OrderHandler.HandleGetByID))

	// 厨房関連
	mux.HandleFunc("POST /kitchen/complete", wrap(app.KitchenHandler.HandleCompleteCooking))

	// メニュー関連
	mux.HandleFunc("POST /menus/{id}/soldout", wrap(app.MenuHandler.HandleSoldOut))

	// DLQ 管理
	mux.HandleFunc("GET /admin/dlq", wrap(app.DLQHandler.HandleList))
	mux.HandleFunc("POST /admin/dlq/{id}/retry", wrap(app.DLQHandler.HandleRetry))

	// 特別注文関連（プロセスマネージャー）
	mux.HandleFunc("POST /special-orders", wrap(app.SpecialOrderHandler.HandleCreate))
	mux.HandleFunc("POST /special-orders/{id}/approve", wrap(app.SpecialOrderHandler.HandleApprove))
	mux.HandleFunc("POST /special-orders/{id}/reject", wrap(app.SpecialOrderHandler.HandleReject))
	mux.HandleFunc("POST /special-orders/{id}/resubmit", wrap(app.SpecialOrderHandler.HandleResubmit))

	// スタッフ管理（アクティブレコード。第10章）
	mux.HandleFunc("GET /staff", wrap(staffHandler.HandleList))
	mux.HandleFunc("GET /staff/{id}", wrap(staffHandler.HandleGetByID))
	mux.HandleFunc("POST /staff", wrap(staffHandler.HandleCreate))
	mux.HandleFunc("DELETE /staff/{id}", wrap(staffHandler.HandleDelete))

	// 閉店処理（トランザクションスクリプト。第10章）
	mux.HandleFunc("POST /admin/close-shop", wrap(closingHandler.HandleCloseShop))

	// インポート関連（非同期バッチ）
	mux.HandleFunc("POST /admin/import/menus", wrap(app.ImportHandler.HandleImportMenus))
	mux.HandleFunc("GET /admin/import/{id}/status", wrap(app.ImportHandler.HandleImportStatus))
}
