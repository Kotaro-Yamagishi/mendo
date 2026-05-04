package domain

import (
	"context"
	"time"
)

// DeadLetter は処理失敗したイベントの記録。
// EventBus のハンドラが設定回数リトライしても失敗した場合に DLQ に保存される。
type DeadLetter struct {
	ID          string
	Event       Event
	Error       string
	FailCount   int
	LastFailAt  time.Time
	HandlerName string
}

// DeadLetterQueue は処理失敗イベントの保管場所。
type DeadLetterQueue interface {
	Store(ctx context.Context, letter *DeadLetter) error
	List(ctx context.Context) ([]DeadLetter, error)
	Remove(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*DeadLetter, error)
}
