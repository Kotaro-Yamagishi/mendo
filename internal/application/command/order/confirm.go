package order

import (
	"context"
	"fmt"

	"mendo/internal/domain"
	"mendo/internal/domain/order"
)

// ConfirmOrderUsecase は注文確定のユースケース（ES版）。
// confirm_es.go の ConfirmOrderESUsecase と同等。
// こちらは ConfirmOrderUsecase として外部から利用される旧名。
type ConfirmOrderUsecase struct {
	eventStore domain.EventStore
	outbox     domain.Outbox
}

func NewConfirmOrderUsecase(es domain.EventStore, ob domain.Outbox) *ConfirmOrderUsecase {
	return &ConfirmOrderUsecase{eventStore: es, outbox: ob}
}

func (uc *ConfirmOrderUsecase) Execute(ctx context.Context, orderID string) error {
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

	// 4. 未コミットイベントをイベントストアに保存
	if err := uc.eventStore.Save(ctx, o.UncommittedEvents()); err != nil {
		return fmt.Errorf("failed to save events: %w", err)
	}

	// 5. Outbox にイベントを保存（Publish は RelayService が行う）
	if err := uc.outbox.Store(ctx, o.UncommittedEvents()); err != nil {
		return fmt.Errorf("failed to store events in outbox: %w", err)
	}

	// 6. イベントをクリア
	o.ClearEvents()

	return nil
}
