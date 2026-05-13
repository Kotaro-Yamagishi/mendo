package order

import (
	"context"
	"log/slog"

	"mendo/internal/apperrors"
	"mendo/internal/domain"
	"mendo/internal/domain/order"
)

// ConfirmOrderESUsecase はイベントソーシング版の注文確定ユースケース。
//
// 従来版との違い:
//   - リポジトリではなくイベントストアを使う
//   - 集約のロード = イベント列を取得して ReconstructFromEvents
//   - 集約の保存 = 未コミットイベントをイベントストアに Save
type ConfirmOrderESUsecase struct {
	eventStore domain.EventStore
	outbox     domain.Outbox
	publisher  domain.EventPublisher
}

func NewConfirmOrderESUsecase(es domain.EventStore, ob domain.Outbox, pub domain.EventPublisher) *ConfirmOrderESUsecase {
	return &ConfirmOrderESUsecase{eventStore: es, outbox: ob, publisher: pub}
}

func (uc *ConfirmOrderESUsecase) Execute(ctx context.Context, orderID string) error {
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

	// 3. コマンド実行（業務ルールチェック + 新しいイベントが追加される）
	if err := o.Confirm(); err != nil {
		return err // domain の AppError をそのまま返す
	}

	// 4. 未コミットイベントをイベントストアに保存（INSERT only）
	if err := uc.eventStore.Save(ctx, o.UncommittedEvents()); err != nil {
		return err
	}

	// 5. Outbox にイベントを保存
	if err := uc.outbox.Store(ctx, o.UncommittedEvents()); err != nil {
		return err
	}

	// 6. EventBus に Publish（Projection 更新や後続ユースケースの起動）
	if err := uc.publisher.Publish(ctx, o.UncommittedEvents()...); err != nil {
		return err
	}

	// 7. イベントをクリア
	o.ClearEvents()

	slog.InfoContext(ctx, "order confirmed", slog.String("order_id", orderID))
	return nil
}
