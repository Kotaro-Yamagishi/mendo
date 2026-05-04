package outbox

import (
	"context"
	"fmt"
	"time"

	"mendo/internal/domain"
)

// RelayService は送信箱のイベントを定期的に EventBus に配信する。
// 本番では別プロセスやゴルーチンで実行する。
type RelayService struct {
	outbox    domain.Outbox
	publisher domain.EventPublisher
	interval  time.Duration
}

func NewRelayService(ob domain.Outbox, pub domain.EventPublisher, interval time.Duration) *RelayService {
	return &RelayService{outbox: ob, publisher: pub, interval: interval}
}

// Start はリレーサービスを開始する。ゴルーチンで実行。
func (r *RelayService) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := r.relay(ctx); err != nil {
					fmt.Printf("[OutboxRelay] error: %v\n", err)
				}
			}
		}
	}()
	fmt.Printf("[OutboxRelay] started (interval: %s)\n", r.interval)
}

func (r *RelayService) relay(ctx context.Context) error {
	events, err := r.outbox.Fetch(ctx, 100)
	if err != nil {
		return fmt.Errorf("failed to fetch outbox events: %w", err)
	}
	if len(events) == 0 {
		return nil
	}

	if err := r.publisher.Publish(ctx, events...); err != nil {
		return fmt.Errorf("failed to publish outbox events: %w", err)
	}

	eventIDs := make([]string, 0, len(events))
	for _, event := range events {
		eventIDs = append(eventIDs, event.GetAggregateID())
	}
	if err := r.outbox.MarkDelivered(ctx, eventIDs); err != nil {
		return fmt.Errorf("failed to mark delivered: %w", err)
	}

	fmt.Printf("[OutboxRelay] delivered %d events\n", len(events))
	return nil
}
