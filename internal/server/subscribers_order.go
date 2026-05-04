package server

import (
	"context"
	"fmt"

	"mendo/internal/di"
	"mendo/internal/domain"
	"mendo/internal/domain/order"
	"mendo/internal/infrastructure/eventbus"
)

// registerOrderSubscribers は Order 集約のイベント購読者を登録する。
func registerOrderSubscribers(bus *eventbus.WatermillEventBus, app *di.App) {
	// OrderCreated → Projection 更新 + ボード更新
	bus.Subscribe(order.EventTypeOrderCreated, func(ctx context.Context, event domain.Event) error {
		// 同じ BC 内: 内部イベントのまま使う
		if err := app.OrderStateStore.HandleEvent(ctx, event); err != nil {
			return fmt.Errorf("projection update failed: %w", err)
		}
		app.OrderBoard.ApplyOrderEvent(event)

		// 他の BC / 外部向け: 公開イベントに変換
		created, ok := event.(order.OrderCreated)
		if !ok {
			return fmt.Errorf("unexpected event type: %T", event)
		}
		publicEvent := order.ToPublicCreated(created)
		fmt.Printf("[EventBus] OrderCreated → 公開イベントに変換 (orderID: %s, seatNo: %d)\n", publicEvent.OrderID, publicEvent.SeatNo)
		return nil
	})

	// OrderConfirmed → Projection 更新 + 厨房連携 + ボード更新
	bus.Subscribe(order.EventTypeOrderConfirmed, func(ctx context.Context, event domain.Event) error {
		// 1. 内部イベントとして Projection 更新（同じ BC 内 → 内部イベントのまま使う）
		if err := app.OrderStateStore.HandleEvent(ctx, event); err != nil {
			return fmt.Errorf("projection update failed: %w", err)
		}
		app.OrderBoard.ApplyOrderEvent(event)

		confirmed, ok := event.(order.OrderConfirmed)
		if !ok {
			return fmt.Errorf("unexpected event type: %T", event)
		}

		// 2. 他の BC へは公開イベントに変換してから渡す
		publicEvent := order.ToPublicConfirmed(confirmed)
		fmt.Printf("[EventBus] OrderConfirmed → 公開イベントに変換 → Kitchen BC (orderID: %s)\n", publicEvent.OrderID)
		return app.StartCookingUC.HandleOrderConfirmedPublic(ctx, publicEvent)
	})

	// OrderCanceled → Projection 更新 + ボード更新
	bus.Subscribe(order.EventTypeOrderCanceled, func(ctx context.Context, event domain.Event) error {
		// 同じ BC 内: 内部イベントのまま使う
		if err := app.OrderStateStore.HandleEvent(ctx, event); err != nil {
			return fmt.Errorf("projection update failed: %w", err)
		}
		app.OrderBoard.ApplyOrderEvent(event)

		// 他の BC / 外部向け: 公開イベントに変換
		canceled, ok := event.(order.OrderCancelled)
		if !ok {
			return fmt.Errorf("unexpected event type: %T", event)
		}
		publicEvent := order.ToPublicCanceled(canceled)
		fmt.Printf("[EventBus] OrderCanceled → 公開イベントに変換 (orderID: %s, reason: %s)\n", publicEvent.OrderID, publicEvent.Reason)
		return nil
	})

	// ItemAdded → Projection 更新
	bus.Subscribe(order.EventTypeItemAdded, func(ctx context.Context, event domain.Event) error {
		if err := app.OrderStateStore.HandleEvent(ctx, event); err != nil {
			return fmt.Errorf("projection update failed: %w", err)
		}
		fmt.Println("[EventBus] ItemAdded → Projection 更新")
		return nil
	})
}
