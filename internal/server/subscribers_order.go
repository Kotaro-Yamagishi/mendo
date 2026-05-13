package server

import (
	"context"
	"fmt"
	"log/slog"

	"mendo/internal/apperrors"
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
			return err
		}
		app.OrderBoard.ApplyOrderEvent(event)

		// 他の BC / 外部向け: 公開イベントに変換
		created, ok := event.(order.OrderCreated)
		if !ok {
			return apperrors.Infrastructure("予期しないイベント型", fmt.Errorf("got %T", event))
		}
		publicEvent := order.ToPublicCreated(created)
		slog.InfoContext(ctx, "event published",
			slog.String("event_type", order.EventTypeOrderCreated),
			slog.String("action", "convert to public event"),
			slog.String("order_id", publicEvent.OrderID),
			slog.Int("seat_no", publicEvent.SeatNo),
		)
		return nil
	})

	// OrderConfirmed → Projection 更新 + 厨房連携 + ボード更新
	bus.Subscribe(order.EventTypeOrderConfirmed, func(ctx context.Context, event domain.Event) error {
		// 1. 内部イベントとして Projection 更新（同じ BC 内 → 内部イベントのまま使う）
		if err := app.OrderStateStore.HandleEvent(ctx, event); err != nil {
			return err
		}
		app.OrderBoard.ApplyOrderEvent(event)

		confirmed, ok := event.(order.OrderConfirmed)
		if !ok {
			return apperrors.Infrastructure("予期しないイベント型", fmt.Errorf("got %T", event))
		}

		// 2. 他の BC へは公開イベントに変換してから渡す
		publicEvent := order.ToPublicConfirmed(confirmed)
		slog.InfoContext(ctx, "event published",
			slog.String("event_type", order.EventTypeOrderConfirmed),
			slog.String("action", "convert to public event"),
			slog.String("order_id", publicEvent.OrderID),
		)
		return app.StartCookingUC.HandleOrderConfirmedPublic(ctx, publicEvent)
	})

	// OrderCanceled → Projection 更新 + ボード更新
	bus.Subscribe(order.EventTypeOrderCanceled, func(ctx context.Context, event domain.Event) error {
		// 同じ BC 内: 内部イベントのまま使う
		if err := app.OrderStateStore.HandleEvent(ctx, event); err != nil {
			return err
		}
		app.OrderBoard.ApplyOrderEvent(event)

		// 他の BC / 外部向け: 公開イベントに変換
		canceled, ok := event.(order.OrderCancelled)
		if !ok {
			return apperrors.Infrastructure("予期しないイベント型", fmt.Errorf("got %T", event))
		}
		publicEvent := order.ToPublicCanceled(canceled)
		slog.InfoContext(ctx, "event published",
			slog.String("event_type", order.EventTypeOrderCanceled),
			slog.String("action", "convert to public event"),
			slog.String("order_id", publicEvent.OrderID),
			slog.String("reason", publicEvent.Reason),
		)
		return nil
	})

	// ItemAdded → Projection 更新
	bus.Subscribe(order.EventTypeItemAdded, func(ctx context.Context, event domain.Event) error {
		if err := app.OrderStateStore.HandleEvent(ctx, event); err != nil {
			return err
		}
		slog.InfoContext(ctx, "projection updated",
			slog.String("event_type", order.EventTypeItemAdded),
		)
		return nil
	})
}
