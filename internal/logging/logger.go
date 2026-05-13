package logging

import (
	"context"
	"log/slog"
	"os"
)

// NewLogger はアプリケーション全体で使うロガーを生成する。
// LOG_FORMAT=json で JSON 出力（本番用）、それ以外は Text 出力（開発用）。
// LOG_LEVEL で出力レベルを制御（DEBUG, INFO, WARN, ERROR）。デフォルト INFO。
func NewLogger() *slog.Logger {
	level := parseLevel(os.Getenv("LOG_LEVEL"))
	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if os.Getenv("LOG_FORMAT") == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(&correlationHandler{inner: handler})
}

func parseLevel(s string) slog.Level {
	switch s {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// correlationHandler は内部の Handler をラップし、CorrelationID を自動付与する。
type correlationHandler struct {
	inner slog.Handler
}

func (h *correlationHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *correlationHandler) Handle(ctx context.Context, r slog.Record) error {
	if id := GetCorrelationID(ctx); id != "" {
		r.AddAttrs(slog.String("correlation_id", id))
	}
	return h.inner.Handle(ctx, r)
}

func (h *correlationHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &correlationHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h *correlationHandler) WithGroup(name string) slog.Handler {
	return &correlationHandler{inner: h.inner.WithGroup(name)}
}
