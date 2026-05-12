package handler

import (
	"net/http"

	kitchencommand "mendo/internal/application/command/kitchen"
	"mendo/internal/domain/order"
)

// KitchenHandler は厨房関連の HTTP ハンドラ。
type KitchenHandler struct {
	completeCookingUC *kitchencommand.CompleteCookingUsecase
}

func NewKitchenHandler(completeCookingUC *kitchencommand.CompleteCookingUsecase) *KitchenHandler {
	return &KitchenHandler{completeCookingUC: completeCookingUC}
}

// HandleCompleteCooking は POST /kitchen/complete のハンドラ。
// 厨房スタッフが調理完了ボタンを押した時に呼ばれる。
func (h *KitchenHandler) HandleCompleteCooking(w http.ResponseWriter, r *http.Request) error {
	var body struct {
		OrderID string `json:"order_id"`
	}
	if err := readJSON(r, &body); err != nil {
		return err
	}

	if err := h.completeCookingUC.Execute(r.Context(), order.OrderID(body.OrderID)); err != nil {
		return err
	}

	WriteSuccess(w, http.StatusOK, map[string]string{"status": "cooking_completed"})
	return nil
}
