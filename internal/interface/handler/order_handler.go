package handler

import (
	"net/http"

	ordercommand "mendo/internal/application/command/order"
	orderquery "mendo/internal/application/query/order"
)

// OrderHandler は注文関連の HTTP ハンドラ。
// ユースケースを呼ぶだけ。業務ロジックは持たない。
type OrderHandler struct {
	createUC     *ordercommand.CreateOrderUsecase
	confirmUC    *ordercommand.ConfirmOrderESUsecase
	cancelUC     *ordercommand.CancelOrderUsecase
	waitTimeUC   *orderquery.EstimateWaitTimeUsecase
	listOrdersUC *orderquery.ListOrdersUsecase
}

func NewOrderHandler(
	createUC *ordercommand.CreateOrderUsecase,
	confirmUC *ordercommand.ConfirmOrderESUsecase,
	cancelUC *ordercommand.CancelOrderUsecase,
	waitTimeUC *orderquery.EstimateWaitTimeUsecase,
	listOrdersUC *orderquery.ListOrdersUsecase,
) *OrderHandler {
	return &OrderHandler{
		createUC:     createUC,
		confirmUC:    confirmUC,
		cancelUC:     cancelUC,
		waitTimeUC:   waitTimeUC,
		listOrdersUC: listOrdersUC,
	}
}

// HandleCreate は POST /orders のハンドラ。食券機から注文を作成する。
func (h *OrderHandler) HandleCreate(w http.ResponseWriter, r *http.Request) error {
	var input ordercommand.CreateOrderInput
	if err := readJSON(r, &input); err != nil {
		return err
	}

	orderID, err := h.createUC.Execute(r.Context(), input)
	if err != nil {
		return err
	}

	WriteSuccess(w, http.StatusCreated, map[string]string{"order_id": orderID})
	return nil
}

// HandleConfirm は POST /orders/{id}/confirm のハンドラ。注文を確定する。
func (h *OrderHandler) HandleConfirm(w http.ResponseWriter, r *http.Request) error {
	orderID := r.PathValue("id") // Go 1.22+ の PathValue

	if err := h.confirmUC.Execute(r.Context(), orderID); err != nil {
		return err
	}

	WriteSuccess(w, http.StatusOK, map[string]string{"status": "confirmed"})
	return nil
}

// HandleCancel は POST /orders/{id}/cancel のハンドラ。注文をキャンセルする。
func (h *OrderHandler) HandleCancel(w http.ResponseWriter, r *http.Request) error {
	orderID := r.PathValue("id")

	var body struct {
		Reason string `json:"reason"`
	}
	if err := readJSON(r, &body); err != nil {
		return err
	}

	if err := h.cancelUC.Execute(r.Context(), orderID, body.Reason); err != nil {
		return err
	}

	WriteSuccess(w, http.StatusOK, map[string]string{"status": "canceled"})
	return nil
}

// HandleWaitTime は GET /wait-time のハンドラ。待ち時間を取得する。
func (h *OrderHandler) HandleWaitTime(w http.ResponseWriter, r *http.Request) error {
	duration, err := h.waitTimeUC.Execute(r.Context())
	if err != nil {
		return err
	}

	WriteSuccess(w, http.StatusOK, map[string]string{
		"estimated_wait": duration.String(),
	})
	return nil
}

// HandleList は GET /orders のハンドラ。
// Projection テーブル（リードモデル）から注文一覧を取得する。
// events テーブルは触らない。イベント再生もしない。速い。
func (h *OrderHandler) HandleList(w http.ResponseWriter, r *http.Request) error {
	rows, err := h.listOrdersUC.Execute(r.Context())
	if err != nil {
		return err
	}
	WriteSuccess(w, http.StatusOK, rows)
	return nil
}

// HandleGetByID は GET /orders/{id} のハンドラ。
// Projection テーブルから1件取得。events テーブルは触らない。
func (h *OrderHandler) HandleGetByID(w http.ResponseWriter, r *http.Request) error {
	orderID := r.PathValue("id")
	row, err := h.listOrdersUC.FindByID(r.Context(), orderID)
	if err != nil {
		return err
	}
	WriteSuccess(w, http.StatusOK, row)
	return nil
}
