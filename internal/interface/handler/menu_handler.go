package handler

import (
	"net/http"

	menucommand "mendo/internal/application/command/menu"
	"mendo/internal/domain/menu"
)

// MenuHandler はメニュー関連の HTTP ハンドラ。
type MenuHandler struct {
	soldOutUC *menucommand.SoldOutMenuUsecase
}

func NewMenuHandler(soldOutUC *menucommand.SoldOutMenuUsecase) *MenuHandler {
	return &MenuHandler{soldOutUC: soldOutUC}
}

// HandleSoldOut は POST /menus/{id}/soldout のハンドラ。メニューを品切れにする。
func (h *MenuHandler) HandleSoldOut(w http.ResponseWriter, r *http.Request) error {
	menuID := menu.MenuID(r.PathValue("id"))

	if err := h.soldOutUC.Execute(r.Context(), menuID); err != nil {
		return err
	}

	WriteSuccess(w, http.StatusOK, map[string]string{"status": "sold_out"})
	return nil
}

