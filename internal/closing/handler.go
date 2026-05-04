package closing

import (
	"encoding/json"
	"log"
	"net/http"
)

// Handler は閉店処理の HTTP ハンドラ。
// トランザクションスクリプトなので usecase 層もない。handler が直接関数を呼ぶ。
type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /admin/close-shop", h.HandleCloseShop)
}

func (h *Handler) HandleCloseShop(w http.ResponseWriter, r *http.Request) {
	// 本番では DB から未完了注文 ID を取得する
	// 学習用なので空リスト
	canceledCount, err := CloseShop([]string{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         "closed",
		"canceled_count": canceledCount,
	}); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
