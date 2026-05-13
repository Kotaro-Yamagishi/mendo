package closing

import (
	"net/http"

	"mendo/internal/interface/handler"
)

// Handler は閉店処理の HTTP ハンドラ。
// トランザクションスクリプトなので usecase 層もない。handler が直接関数を呼ぶ。
type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) HandleCloseShop(w http.ResponseWriter, r *http.Request) error {
	// 本番では DB から未完了注文 ID を取得する
	// 学習用なので空リスト
	canceledCount, err := CloseShop(r.Context(), []string{})
	if err != nil {
		return err
	}
	handler.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"status":         "closed",
		"canceled_count": canceledCount,
	})
	return nil
}
