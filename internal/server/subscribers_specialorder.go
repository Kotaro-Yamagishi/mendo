package server

import (
	"context"
	"fmt"
	"log/slog"

	"mendo/internal/apperrors"
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
			return apperrors.Infrastructure("予期しないイベント型", fmt.Errorf("got %T", event))
		}
		publicEvent := contract.OrderConfirmedPublic{OrderID: dispatched.OrderID}
		slog.InfoContext(ctx, "process manager",
			slog.String("event_type", specialorder.EventTypeCookingDispatched),
			slog.String("action", "dispatch to kitchen"),
			slog.String("order_id", dispatched.OrderID),
		)
		return app.StartCookingUC.HandleOrderConfirmedPublic(ctx, publicEvent)
	})

	// SpecialOrderRejected → お客さんに別メニュー提案通知
	bus.Subscribe(specialorder.EventTypeSpecialOrderRejected, func(ctx context.Context, event domain.Event) error {
		rejected, ok := event.(specialorder.SpecialOrderRejected)
		if !ok {
			return apperrors.Infrastructure("予期しないイベント型", fmt.Errorf("got %T", event))
		}
		slog.InfoContext(ctx, "process manager",
			slog.String("event_type", specialorder.EventTypeSpecialOrderRejected),
			slog.String("action", "notify customer"),
			slog.String("reason", rejected.Reason),
			slog.String("suggested_menu", rejected.SuggestedMenu),
		)
		return nil
	})
}
