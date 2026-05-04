package order

import (
	"context"

	"mendo/internal/domain"
)

// --- 書き込み側（EventStore 経由） ---

// Reader は Order の読み取り操作を定義する。
type Reader interface {
	CountPending(ctx context.Context) (int, error)
}

// --- 読み取り側（Projection） ---

// OrderStateRow は Projection テーブルの1行に相当するリードモデル。
// events テーブルから計算した「今の状態」のキャッシュ。
type OrderStateRow struct {
	OrderID   string
	SeatNo    int
	Status    string
	ItemCount int
}

// ProjectionReader は Projection（リードモデル）の読み取り IF。
type ProjectionReader interface {
	FindByID(ctx context.Context, orderID string) (*OrderStateRow, error)
	FindAll(ctx context.Context) ([]OrderStateRow, error)
}

// ProjectionWriter は Projection（リードモデル）の書き込み IF。
// ドメインイベントを受け取って Projection を更新する。
type ProjectionWriter interface {
	HandleEvent(ctx context.Context, event domain.Event) error
}
