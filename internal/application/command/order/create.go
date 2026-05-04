package order

import (
	"context"
	"fmt"

	"mendo/internal/domain"
	"mendo/internal/domain/menu"
	"mendo/internal/domain/order"
)

// CreateOrderInput は注文作成の入力。
type CreateOrderInput struct {
	SeatNo int
	Items  []CreateOrderItemInput
}

type CreateOrderItemInput struct {
	MenuID   string
	Toppings []string
	Hardness string
}

// CreateOrderUsecase は注文作成のユースケース（ES版）。
// 食券機で注文が作られた時に実行される。この時点ではまだ確定（Confirm）されていない。
type CreateOrderUsecase struct {
	eventStore domain.EventStore
	outbox     domain.Outbox
}

func NewCreateOrderUsecase(es domain.EventStore, ob domain.Outbox) *CreateOrderUsecase {
	return &CreateOrderUsecase{eventStore: es, outbox: ob}
}

func (uc *CreateOrderUsecase) Execute(ctx context.Context, input CreateOrderInput) (string, error) {
	// 1. 注文 ID を生成
	orderID := order.NewOrderID().String()

	// 2. 注文集約を生成（OrderCreated イベントが uncommitted に積まれる）
	o := order.Create(orderID, input.SeatNo)

	// 3. 注文明細を追加（ItemAdded イベントが uncommitted に積まれる）
	for _, itemInput := range input.Items {
		if err := o.AddItem(menu.MenuID(itemInput.MenuID), itemInput.Toppings, itemInput.Hardness); err != nil {
			return "", fmt.Errorf("failed to add item: %w", err)
		}
	}

	// 4. 未コミットイベントをイベントストアに保存
	if err := uc.eventStore.Save(ctx, o.UncommittedEvents()); err != nil {
		return "", fmt.Errorf("failed to save events: %w", err)
	}

	// 5. Outbox にイベントを保存（Publish は RelayService が行う）
	if err := uc.outbox.Store(ctx, o.UncommittedEvents()); err != nil {
		return "", fmt.Errorf("failed to store events in outbox: %w", err)
	}

	// 6. イベントをクリア
	o.ClearEvents()

	return orderID, nil
}
