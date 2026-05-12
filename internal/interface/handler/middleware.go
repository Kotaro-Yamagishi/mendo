package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"mendo/internal/apperrors"
)

// AppHandlerFunc は error を返せる handler。
type AppHandlerFunc func(w http.ResponseWriter, r *http.Request) error

// ErrorMiddleware は AppHandlerFunc を http.HandlerFunc に変換する。
// エラー変換、panic recovery、ログ出力を一括で行う。
func ErrorMiddleware(h AppHandlerFunc, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error("panic recovered",
					"panic", rec,
					"method", r.Method,
					"path", r.URL.Path,
				)
				writeErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", "")
			}
		}()

		err := h(w, r)
		if err == nil {
			return
		}

		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}

		var appErr *apperrors.AppError
		if !errors.As(err, &appErr) {
			// AppError でない = 想定外
			logger.Error("unexpected error",
				"error", err,
				"trace_id", traceID,
				"method", r.Method,
				"path", r.URL.Path,
			)
			writeErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", traceID)
			return
		}

		appErr.TraceID = traceID
		status := appErr.Category.HTTPStatus()

		if !appErr.Category.IsClientError() {
			// サーバーエラー: メッセージを隠す
			logger.Error("system error",
				"code", appErr.Code,
				"message", appErr.Message,
				"cause", appErr.Cause,
				"details", appErr.Details,
				"trace_id", traceID,
				"method", r.Method,
				"path", r.URL.Path,
			)
			writeErrorJSON(w, status, appErr.Code, "internal server error", traceID)
			return
		}

		// クライアントエラー: メッセージを返す
		logger.Warn("client error",
			"code", appErr.Code,
			"message", appErr.Message,
			"details", appErr.Details,
			"trace_id", traceID,
		)
		writeErrorJSONWithDetails(w, status, appErr, traceID)
	}
}

func writeErrorJSON(w http.ResponseWriter, status int, code, message, traceID string) {
	errBody := map[string]any{
		"code":    code,
		"message": message,
	}
	if traceID != "" {
		errBody["trace_id"] = traceID
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"error": errBody}) //nolint:errcheck
}

func writeErrorJSONWithDetails(w http.ResponseWriter, status int, appErr *apperrors.AppError, traceID string) {
	errBody := map[string]any{
		"code":    appErr.Code,
		"message": appErr.Message,
	}
	if traceID != "" {
		errBody["trace_id"] = traceID
	}
	// Validation の fields を追加
	if fields, ok := appErr.Details["fields"]; ok {
		errBody["fields"] = fields
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"error": errBody}) //nolint:errcheck
}
