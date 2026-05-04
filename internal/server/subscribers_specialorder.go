package server

import (
	"context"
	"fmt"

	"mendo/internal/domain/contract"
	"mendo/internal/di"
	"mendo/internal/domain"
	"mendo/internal/domain/specialorder"
	"mendo/internal/infrastructure/eventbus"
)

func registerSpecialOrderSubscribers(bus *eventbus.WatermillEventBus, app *di.App) {
	// CookingDispatched → Kitchen BC に調理タスク追加（プロセスマネージャーから厨房へ）
	bus.Subscribe(specialorder.EventTypeCookingDispatched, func(ctx context.Context, event domain.Event) error {
		dispatched, ok := event.(specialorder.CookingDispatched)
		if !ok {
			return fmt.Errorf("unexpected event type: %T", event)
		}
		publicEvent := contract.OrderConfirmedPublic{OrderID: dispatched.OrderID}
		fmt.Printf("[ProcessManager] CookingDispatched → Kitchen BC に調理指示 (orderID: %s)\n", dispatched.OrderID)
		return app.StartCookingUC.HandleOrderConfirmedPublic(ctx, publicEvent)
	})

	// SpecialOrderRejected → お客さんに別メニュー提案通知
	bus.Subscribe(specialorder.EventTypeSpecialOrderRejected, func(ctx context.Context, event domain.Event) error {
		rejected, ok := event.(specialorder.SpecialOrderRejected)
		if !ok {
			return fmt.Errorf("unexpected event type: %T", event)
		}
		fmt.Printf("[ProcessManager] 特別注文却下 → お客さんに通知 (reason: %s, suggested: %s)\n", rejected.Reason, rejected.SuggestedMenu)
		return nil
	})
}
