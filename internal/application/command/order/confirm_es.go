package order

import (
	"context"
	"fmt"

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
		return fmt.Errorf("failed to load events: %w", err)
	}

	// 2. イベント列から集約の現在の状態を復元
	o := order.ReconstructFromEvents(events)

	// 3. コマンド実行（業務ルールチェック + 新しいイベントが追加される）
	if err := o.Confirm(); err != nil {
		return fmt.Errorf("failed to confirm order: %w", err)
	}

	// 4. 未コミットイベントをイベントストアに保存（INSERT only）
	if err := uc.eventStore.Save(ctx, o.UncommittedEvents()); err != nil {
		return fmt.Errorf("failed to save events: %w", err)
	}

	// 5. Outbox にイベントを保存
	if err := uc.outbox.Store(ctx, o.UncommittedEvents()); err != nil {
		return fmt.Errorf("failed to store events in outbox: %w", err)
	}

	// 6. EventBus に Publish（Projection 更新や後続ユースケースの起動）
	if err := uc.publisher.Publish(ctx, o.UncommittedEvents()...); err != nil {
		return fmt.Errorf("failed to publish events: %w", err)
	}

	// 7. イベントをクリア
	o.ClearEvents()

	return nil
}
