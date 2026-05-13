package outbox

import (
	"context"
	"log/slog"
	"time"

	"mendo/internal/apperrors"
	"mendo/internal/domain"
)

// RelayService は送信箱のイベントを定期的に EventBus に配信する。
// 本番では別プロセスやゴルーチンで実行する。
type RelayService struct {
	outbox    domain.Outbox
	publisher domain.EventPublisher
	interval  time.Duration
	logger    *slog.Logger
}

func NewRelayService(ob domain.Outbox, pub domain.EventPublisher, interval time.Duration, logger *slog.Logger) *RelayService {
	return &RelayService{outbox: ob, publisher: pub, interval: interval, logger: logger}
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
					r.logger.Error("outbox relay error", "error", err)
				}
			}
		}
	}()
	r.logger.Info("outbox relay started", slog.String("interval", r.interval.String()))
}

func (r *RelayService) relay(ctx context.Context) error {
	events, err := r.outbox.Fetch(ctx, 100)
	if err != nil {
		return apperrors.Infrastructure("Outbox イベントの取得に失敗", err)
	}
	if len(events) == 0 {
		return nil
	}

	if err := r.publisher.Publish(ctx, events...); err != nil {
		return apperrors.Infrastructure("Outbox イベントの配信に失敗", err)
	}

	eventIDs := make([]string, 0, len(events))
	for _, event := range events {
		eventIDs = append(eventIDs, event.GetAggregateID())
	}
	if err := r.outbox.MarkDelivered(ctx, eventIDs); err != nil {
		return apperrors.Infrastructure("Outbox 配信済みマークに失敗", err)
	}

	r.logger.Info("outbox relay delivered", slog.Int("count", len(events)))
	return nil
}
