package logging

import (
	"log/slog"
	"net/http"
	"time"
)

// CorrelationMiddleware は CorrelationID を生成/取得して context に格納する。
// リクエスト開始/終了のログも出力する。
func CorrelationMiddleware(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CorrelationID: ヘッダから取得 or 新規生成
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = NewCorrelationID()
		}

		// context に格納
		ctx := WithCorrelationID(r.Context(), correlationID)
		r = r.WithContext(ctx)

		// レスポンスヘッダにも付与（デバッグ・トレース用）
		w.Header().Set("X-Correlation-ID", correlationID)

		// リクエスト開始ログ
		start := time.Now()
		logger.InfoContext(ctx, "request started",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)

		// レスポンスのステータスコードを記録するラッパー
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)

		// リクエスト完了ログ
		duration := time.Since(start)
		logger.InfoContext(ctx, "request completed",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", sw.status),
			slog.Int64("duration_ms", duration.Milliseconds()),
		)
	})
}

// statusWriter は http.ResponseWriter をラップしてステータスコードを記録する。
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
