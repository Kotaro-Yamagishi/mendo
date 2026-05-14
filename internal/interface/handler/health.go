package handler

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

// HealthHandler はヘルスチェック用のハンドラ。
type HealthHandler struct {
	ready atomic.Bool
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// SetReady はサーバーがリクエストを受ける準備ができたことを通知する。
func (h *HealthHandler) SetReady() {
	h.ready.Store(true)
}

// SetNotReady はサーバーがリクエストを受けられない状態にする（Graceful Shutdown 開始時）。
func (h *HealthHandler) SetNotReady() {
	h.ready.Store(false)
}

// HandleLive は GET /health/live のハンドラ。プロセスが生きていれば 200。
func (h *HealthHandler) HandleLive(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"}) //nolint:errcheck
}

// HandleReady は GET /health/ready のハンドラ。
// リクエストを受ける準備ができていれば 200、そうでなければ 503。
func (h *HealthHandler) HandleReady(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.ready.Load() {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ready"}) //nolint:errcheck
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "not ready"}) //nolint:errcheck
	}
}
