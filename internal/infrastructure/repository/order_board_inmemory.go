package repository

import (
	"time"

	"mendo/internal/domain"
	"mendo/internal/domain/kitchen"
	"mendo/internal/domain/order"
)

// OrderBoardProjection は「注文状況ボード」のリードモデル。
// Order と Kitchen の両方のイベントを横断して1つのビューを作る。
//
// 複数の集約のイベントを使うため、特定の集約パッケージには置けない。
// → infrastructure/repository/ に配置する。
//
// 用途: 店舗のモニターに「注文一覧 + 調理状況」を表示する画面用。
type OrderBoardProjection struct {
	Entries []OrderBoardEntry
}

type OrderBoardEntry struct {
	OrderID       string
	SeatNo        int
	OrderStatus   string     // "pending", "confirmed", "canceled"
	CookingStatus string     // "waiting", "cooking", "completed", ""
	OrderedAt     time.Time
	CookingDoneAt *time.Time // 調理完了時刻（nil = まだ）
}

func NewOrderBoardProjection() *OrderBoardProjection {
	return &OrderBoardProjection{}
}

// ApplyOrderEvent は Order 集約のイベントを処理する。
func (p *OrderBoardProjection) ApplyOrderEvent(event domain.Event) {
	switch e := event.(type) {
	case order.OrderCreated:
		p.Entries = append(p.Entries, OrderBoardEntry{
			OrderID:       e.GetAggregateID(),
			SeatNo:        e.SeatNo,
			OrderStatus:   "pending",
			CookingStatus: "",
			OrderedAt:     e.OccurredAt,
		})
	case order.OrderConfirmed:
		entry := p.findEntry(e.GetAggregateID())
		if entry != nil {
			entry.OrderStatus = "confirmed"
			entry.CookingStatus = "waiting" // 確定されたら調理待ち
		}
	case order.OrderCancelled:
		entry := p.findEntry(e.GetAggregateID())
		if entry != nil {
			entry.OrderStatus = "canceled"
			entry.CookingStatus = ""
		}
	}
}

// ApplyKitchenEvent は Kitchen 集約のイベントを処理する。
func (p *OrderBoardProjection) ApplyKitchenEvent(event domain.Event) {
	if e, ok := event.(kitchen.CookingCompleted); ok {
		entry := p.findEntry(string(e.OrderID))
		if entry != nil {
			entry.CookingStatus = "completed"
			now := e.OccurredAt
			entry.CookingDoneAt = &now
		}
	}
}

func (p *OrderBoardProjection) findEntry(orderID string) *OrderBoardEntry {
	for i := range p.Entries {
		if p.Entries[i].OrderID == orderID {
			return &p.Entries[i]
		}
	}
	return nil
}
