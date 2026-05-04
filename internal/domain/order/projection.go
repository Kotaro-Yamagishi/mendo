package order

import (
	"mendo/internal/domain"
	"mendo/internal/domain/menu"
	"time"
)

// --- 状態モデル（管理画面表示用）---

// OrderStateProjection は注文の現在の状態を表すリードモデル。
// イベント列を Apply して生成する。業務ルールは持たない。読み取り専用。
type OrderStateProjection struct {
	OrderID   string
	SeatNo    int
	Items     []Item
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
	Version   int
}

func NewOrderStateProjection(events []domain.Event) *OrderStateProjection {
	p := &OrderStateProjection{}
	for _, event := range events {
		p.Apply(event)
	}
	return p
}

func (p *OrderStateProjection) Apply(event domain.Event) {
	switch e := event.(type) {
	case OrderCreated:
		p.OrderID = e.AggregateID
		p.SeatNo = e.SeatNo
		p.Status = "pending"
		p.CreatedAt = e.OccurredAt
		p.UpdatedAt = e.OccurredAt
	case ItemAdded:
		p.Items = append(p.Items, Item{
			MenuID:   e.MenuID,
			Toppings: e.Toppings,
			Hardness: e.Hardness,
		})
		p.UpdatedAt = e.OccurredAt
	case OrderConfirmed:
		p.Status = "confirmed"
		p.UpdatedAt = e.OccurredAt
	case OrderCancelled:
		p.Status = "canceled"
		p.UpdatedAt = e.OccurredAt
	}
	p.Version++
}

// --- 分析モデル（営業分析用）---

// OrderAnalyticsProjection は注文の分析データを表すリードモデル。
// 同じイベント列から別の観点でデータを生成する。
type OrderAnalyticsProjection struct {
	OrderID       string
	ItemCount     int
	ToppingCount  int
	Status        string
	MenuIDs       []menu.MenuID // どのメニューが注文されたか
	TimeToConfirm time.Duration // 作成から確定までの時間
	createdAt     time.Time
}

func NewOrderAnalyticsProjection(events []domain.Event) *OrderAnalyticsProjection {
	p := &OrderAnalyticsProjection{}
	for _, event := range events {
		p.Apply(event)
	}
	return p
}

func (p *OrderAnalyticsProjection) Apply(event domain.Event) {
	switch e := event.(type) {
	case OrderCreated:
		p.OrderID = e.AggregateID
		p.Status = "pending"
		p.createdAt = e.OccurredAt
	case ItemAdded:
		p.ItemCount++
		p.ToppingCount += len(e.Toppings)
		p.MenuIDs = append(p.MenuIDs, e.MenuID)
	case OrderConfirmed:
		p.Status = "confirmed"
		p.TimeToConfirm = e.OccurredAt.Sub(p.createdAt)
	case OrderCancelled:
		p.Status = "canceled"
	}
}
