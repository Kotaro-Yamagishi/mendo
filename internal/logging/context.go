package logging

import (
	"context"

	"github.com/google/uuid"
)

type correlationIDKey struct{}

// WithCorrelationID は CorrelationID を context に格納する。
func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationIDKey{}, id)
}

// GetCorrelationID は context から CorrelationID を取得する。見つからなければ空文字を返す。
func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(correlationIDKey{}).(string); ok {
		return id
	}
	return ""
}

// NewCorrelationID は新しい CorrelationID を生成する。
func NewCorrelationID() string {
	return uuid.New().String()
}
