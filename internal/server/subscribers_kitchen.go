package server

import (
	"context"
	"fmt"

	"mendo/internal/apperrors"
	"mendo/internal/di"
	"mendo/internal/domain"
	"mendo/internal/domain/kitchen"
	"mendo/internal/infrastructure/eventbus"
)

// registerKitchenSubscribers は Kitchen 集約のイベント購読者を登録する。
func registerKitchenSubscribers(bus *eventbus.WatermillEventBus, app *di.App) {
	// CookingCompleted → 提供通知 + ボード更新
	bus.Subscribe(kitchen.EventTypeCookingCompleted, func(ctx context.Context, event domain.Event) error {
		app.OrderBoard.ApplyKitchenEvent(event)
		completed, ok := event.(kitchen.CookingCompleted)
		if !ok {
			return apperrors.Infrastructure("予期しないイベント型", fmt.Errorf("got %T", event))
		}
		fmt.Printf("[EventBus] CookingCompleted → 提供通知 + ボード更新 (orderID: %s)\n", completed.OrderID)
		return nil
	})

	// CookingRejected → 補償アクション: Order をキャンセル
	bus.Subscribe(kitchen.EventTypeCookingRejected, func(ctx context.Context, event domain.Event) error {
		rejected, ok := event.(kitchen.CookingRejected)
		if !ok {
			return apperrors.Infrastructure("予期しないイベント型", fmt.Errorf("got %T", event))
		}
		fmt.Printf("[Saga] CookingRejected → 補償アクション: Order をキャンセル (orderID: %s, reason: %s)\n", rejected.OrderID, rejected.Reason)
		return app.CancelOrderUC.Execute(ctx, rejected.OrderID, "厨房フル稼働のため自動キャンセル")
	})
}
