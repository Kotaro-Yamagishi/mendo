package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"mendo/internal/apperrors"
	"mendo/internal/logging"
)

// AppHandlerFunc は error を返せる handler。
type AppHandlerFunc func(w http.ResponseWriter, r *http.Request) error

// ErrorMiddleware は AppHandlerFunc を http.HandlerFunc に変換する。
// エラー変換、panic recovery、ログ出力を一括で行う。
func ErrorMiddleware(h AppHandlerFunc, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.ErrorContext(r.Context(), "panic recovered",
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

		correlationID := logging.GetCorrelationID(r.Context())
		if correlationID == "" {
			correlationID = logging.NewCorrelationID()
		}

		var appErr *apperrors.AppError
		if !errors.As(err, &appErr) {
			// AppError でない = 想定外
			logger.ErrorContext(r.Context(), "unexpected error",
				"error", err,
				"method", r.Method,
				"path", r.URL.Path,
			)
			writeErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", correlationID)
			return
		}

		appErr.CorrelationID = correlationID
		status := appErr.Category.HTTPStatus()

		if !appErr.Category.IsClientError() {
			// サーバーエラー: メッセージを隠す
			logger.ErrorContext(r.Context(), "system error",
				"code", appErr.Code,
				"message", appErr.Message,
				"cause", appErr.Cause,
				"details", appErr.Details,
				"method", r.Method,
				"path", r.URL.Path,
			)
			writeErrorJSON(w, status, appErr.Code, "internal server error", correlationID)
			return
		}

		// クライアントエラー: メッセージを返す
		logger.WarnContext(r.Context(), "client error",
			"code", appErr.Code,
			"message", appErr.Message,
			"details", appErr.Details,
		)
		writeErrorJSONWithDetails(w, status, appErr, correlationID)
	}
}

func writeErrorJSON(w http.ResponseWriter, status int, code, message, correlationID string) {
	errBody := map[string]any{
		"code":    code,
		"message": message,
	}
	if correlationID != "" {
		errBody["correlation_id"] = correlationID
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"error": errBody}) //nolint:errcheck
}

func writeErrorJSONWithDetails(w http.ResponseWriter, status int, appErr *apperrors.AppError, correlationID string) {
	errBody := map[string]any{
		"code":    appErr.Code,
		"message": appErr.Message,
	}
	if correlationID != "" {
		errBody["correlation_id"] = correlationID
	}
	// Validation の fields を追加
	if fields, ok := appErr.Details["fields"]; ok {
		errBody["fields"] = fields
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"error": errBody}) //nolint:errcheck
}
