package order

import (
	"mendo/internal/domain"
	"mendo/internal/apperrors"
	"mendo/internal/domain/menu"
)

// Order はイベントソーシング版の注文集約。
//
// 従来版（domain/order/）との違い:
//   - 状態を DB に直接保存しない
//   - 状態変化を全てイベントとして記録
//   - イベント列を Apply して現在の状態を復元する
type Order struct {
	id     string
	seatNo int
	items  []Item
	status string
	version int

	// uncommittedEvents はまだ EventStore に保存されていないイベント。
	// Save 後に Clear される。
	uncommittedEvents []domain.Event
}

type Item struct {
	MenuID   menu.MenuID
	Toppings []string
	Hardness string
}

// --- イベントから状態を復元する ---

// ReconstructFromEvents はイベントストアからロードしたイベント列を
// 順に Apply して集約の現在の状態を復元する。
func ReconstructFromEvents(events []domain.Event) *Order {
	o := &Order{}
	for _, event := range events {
		o.apply(event)
		o.version++
	}
	return o
}

// apply はイベントを集約の状態に反映する（内部用）。
// ここには業務ルールを書かない。状態の反映のみ。
func (o *Order) apply(event domain.Event) {
	switch e := event.(type) {
	case OrderCreated:
		o.id = e.AggregateID
		o.seatNo = e.SeatNo
		o.status = "pending"
	case ItemAdded:
		o.items = append(o.items, Item{
			MenuID:   e.MenuID,
			Toppings: e.Toppings,
			Hardness: e.Hardness,
		})
	case OrderConfirmed:
		o.status = "confirmed"
	case OrderCancelled:
		o.status = "canceled"
	}
}

// --- コマンド（業務ルール + イベント発行）---

// Create は注文を作成する。
func Create(orderID string, seatNo int) *Order {
	o := &Order{}
	event := NewOrderCreated(orderID, seatNo, "")
	o.apply(event)
	o.uncommittedEvents = append(o.uncommittedEvents, event)
	return o
}

// AddItem は注文に商品を追加する。
func (o *Order) AddItem(menuID menu.MenuID, toppings []string, hardness string) error {
	// 業務ルール: トッピングは3つまで
	if len(toppings) > 3 {
		return apperrors.Domain(ErrCodeTooManyToppings, "トッピングは3つまでです")
	}

	event := NewItemAdded(o.id, menuID, toppings, hardness, "")
	o.apply(event)
	o.uncommittedEvents = append(o.uncommittedEvents, event)
	return nil
}

// Confirm は注文を確定する。
func (o *Order) Confirm() error {
	// 業務ルール
	if o.status != "pending" {
		return apperrors.Domain(ErrCodeNotPending, "確定待ち以外は確定できません")
	}
	if len(o.items) == 0 {
		return apperrors.Domain(ErrCodeEmptyItems, "注文が空です")
	}

	confirmedItems := make([]ConfirmedItem, len(o.items))
	for i, item := range o.items {
		confirmedItems[i] = ConfirmedItem{
			MenuID:   string(item.MenuID),
			Toppings: item.Toppings,
			Hardness: item.Hardness,
		}
	}
	event := NewOrderConfirmed(o.id, confirmedItems, o.seatNo, "")
	o.apply(event)
	o.uncommittedEvents = append(o.uncommittedEvents, event)
	return nil
}

// Cancel は注文をキャンセルする。
func (o *Order) Cancel(reason string) error {
	if o.status != StatusConfirmed {
		return apperrors.Domain(ErrCodeNotConfirmed, "確定済みのみキャンセルできます")
	}

	event := NewOrderCancelled(o.id, reason, "")
	o.apply(event)
	o.uncommittedEvents = append(o.uncommittedEvents, event)
	return nil
}

// --- アクセサ ---

func (o *Order) ID() string                       { return o.id }
func (o *Order) Status() string                    { return o.status }
func (o *Order) Version() int                      { return o.version }
func (o *Order) UncommittedEvents() []domain.Event { return o.uncommittedEvents }
func (o *Order) ClearEvents()                      { o.uncommittedEvents = nil }
