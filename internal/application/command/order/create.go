package order

import (
	"context"

	"mendo/internal/domain"
	"mendo/internal/domain/menu"
	"mendo/internal/domain/order"
)

// CreateOrderInput は注文作成の入力。
type CreateOrderInput struct {
	SeatNo int                    `json:"seat_no"`
	Items  []CreateOrderItemInput `json:"items"`
}

type CreateOrderItemInput struct {
	MenuID   string   `json:"menu_id"`
	Toppings []string `json:"toppings"`
	Hardness string   `json:"hardness"`
}

// CreateOrderUsecase は注文作成のユースケース（ES版）。
// 食券機で注文が作られた時に実行される。この時点ではまだ確定（Confirm）されていない。
type CreateOrderUsecase struct {
	eventStore domain.EventStore
	outbox     domain.Outbox
	publisher  domain.EventPublisher
}

func NewCreateOrderUsecase(es domain.EventStore, ob domain.Outbox, pub domain.EventPublisher) *CreateOrderUsecase {
	return &CreateOrderUsecase{eventStore: es, outbox: ob, publisher: pub}
}

func (uc *CreateOrderUsecase) Execute(ctx context.Context, input CreateOrderInput) (string, error) {
	// 1. 注文 ID を生成
	orderID := order.NewOrderID().String()

	// 2. 注文集約を生成（OrderCreated イベントが uncommitted に積まれる）
	o := order.Create(orderID, input.SeatNo)

	// 3. 注文明細を追加（ItemAdded イベントが uncommitted に積まれる）
	for _, itemInput := range input.Items {
		if err := o.AddItem(menu.MenuID(itemInput.MenuID), itemInput.Toppings, itemInput.Hardness); err != nil {
			return "", err // domain の AppError をそのまま返す
		}
	}

	// 4. 未コミットイベントをイベントストアに保存
	if err := uc.eventStore.Save(ctx, o.UncommittedEvents()); err != nil {
		return "", err
	}

	// 5. Outbox にイベントを保存
	if err := uc.outbox.Store(ctx, o.UncommittedEvents()); err != nil {
		return "", err
	}

	// 6. EventBus に Publish（Projection 更新や後続ユースケースの起動）
	if err := uc.publisher.Publish(ctx, o.UncommittedEvents()...); err != nil {
		return "", err
	}

	// 7. イベントをクリア
	o.ClearEvents()

	return orderID, nil
}
