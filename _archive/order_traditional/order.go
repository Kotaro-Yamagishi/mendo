package order

import (
	"context"
	"errors"
	"mendo/internal/domain"
	"mendo/internal/domain/menu"
)

// Order は注文集約のルート。外部からの唯一のアクセス入口。
type Order struct {
	id           OrderID
	items        []OrderItem
	seatNo       SeatNumber
	status       Status
	domainEvents []domain.Event
}

func NewOrder(seatNo SeatNumber) *Order {
	return &Order{
		id:     NewOrderID(),
		seatNo: seatNo,
		status: StatusPending,
	}
}

// AddItem は注文に商品を追加する。
func (o *Order) AddItem(item OrderItem) {
	o.items = append(o.items, item)
}

// Confirm は注文を確定する。業務ルールをチェックしてからイベントを発行する。
func (o *Order) Confirm(ctx context.Context, menuReader menu.Reader) error {
	if o.status != StatusPending {
		return errors.New("確定待ち以外の注文は確定できません")
	}

	if len(o.items) == 0 {
		return errors.New("注文が空です")
	}

	// 業務ルール: 品切れメニューは注文不可
	for _, item := range o.items {
		m, err := menuReader.FindByID(ctx, item.MenuID())
		if err != nil {
			return err
		}
		if !m.IsAvailable() {
			return errors.New(m.Name().String() + " は品切れです")
		}
	}

	// 業務ルール: トッピングは3つまで
	for _, item := range o.items {
		if len(item.Toppings()) > 3 {
			return errors.New("トッピングは3つまでです")
		}
	}

	// 状態変更
	o.status = StatusConfirmed

	// ドメインイベント発行
	o.domainEvents = append(o.domainEvents, NewOrderConfirmed(o.id, o.items, o.seatNo, ""))

	return nil
}

// Cancel は注文をキャンセルする。
func (o *Order) Cancel(reason string) error {
	// 業務ルール: 確定済みの注文のみキャンセル可能
	if o.status != StatusConfirmed {
		return errors.New("確定済みの注文のみキャンセルできます")
	}

	o.status = StatusCancelled
	o.domainEvents = append(o.domainEvents, NewOrderCancelled(o.id, reason, ""))
	return nil
}

// --- アクセサ ---

func (o *Order) ID() OrderID               { return o.id }
func (o *Order) Status() Status             { return o.status }
func (o *Order) Items() []OrderItem         { return o.items }
func (o *Order) DomainEvents() []domain.Event { return o.domainEvents }
