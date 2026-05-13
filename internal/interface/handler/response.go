package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"mendo/internal/apperrors"
)

type successResponse struct {
	Data interface{} `json:"data"`
}

// WriteSuccess は成功レスポンスを JSON で書き込む。外部パッケージ（staff, closing 等）から利用可能。
func WriteSuccess(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(successResponse{Data: data}); err != nil {
		slog.Error("failed to write response", "error", err)
	}
}

func readJSON(r *http.Request, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return apperrors.Validation("INVALID_REQUEST_BODY", "リクエストの形式が不正です").WithCause(err)
	}
	return nil
}
