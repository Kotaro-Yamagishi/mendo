package order

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"

	"mendo/internal/apperrors"
	"mendo/internal/domain"
	"mendo/internal/domain/order"
)

// CancelOrderUsecase は注文キャンセルのユースケース（ES版）。
type CancelOrderUsecase struct {
	eventStore domain.EventStore
	outbox     domain.Outbox
}

func NewCancelOrderUsecase(es domain.EventStore, ob domain.Outbox) *CancelOrderUsecase {
	return &CancelOrderUsecase{eventStore: es, outbox: ob}
}

func (uc *CancelOrderUsecase) Execute(ctx context.Context, orderID, reason string) error {
	ctx, span := otel.Tracer("mendo").Start(ctx, "CancelOrder.Execute")
	defer span.End()

	// 1. イベントストアからイベント列をロード
	events, err := uc.eventStore.Load(ctx, orderID)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return apperrors.NotFound("order", orderID)
	}

	// 2. イベント列から集約の現在の状態を復元
	o := order.ReconstructFromEvents(events)

	// 3. キャンセルコマンド実行（業務ルール: 確定済みのみキャンセル可能）
	if err := o.Cancel(reason); err != nil {
		return err // domain の AppError をそのまま返す
	}

	// 4. 未コミットイベントをイベントストアに保存
	if err := uc.eventStore.Save(ctx, o.UncommittedEvents()); err != nil {
		return err
	}

	// 5. Outbox にイベントを保存（Publish は RelayService が行う）
	if err := uc.outbox.Store(ctx, o.UncommittedEvents()); err != nil {
		return err
	}

	// 6. イベントをクリア
	o.ClearEvents()

	slog.InfoContext(ctx, "order canceled", slog.String("order_id", orderID), slog.String("reason", reason))
	return nil
}
