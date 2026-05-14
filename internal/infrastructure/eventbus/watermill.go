package eventbus

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"

	"mendo/internal/apperrors"
	"mendo/internal/domain"
)

type EventHandler func(ctx context.Context, event domain.Event) error

type WatermillEventBus struct {
	handlers   map[string][]EventHandler
	dlq        domain.DeadLetterQueue
	maxRetries int
	logger     *slog.Logger
}

func NewWatermillEventBus(dlq domain.DeadLetterQueue, maxRetries int, logger *slog.Logger) *WatermillEventBus {
	return &WatermillEventBus{
		handlers:   make(map[string][]EventHandler),
		dlq:        dlq,
		maxRetries: maxRetries,
		logger:     logger,
	}
}

func (b *WatermillEventBus) Publish(ctx context.Context, events ...domain.Event) error {
	for _, event := range events {
		if err := b.publishOne(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func (b *WatermillEventBus) publishOne(ctx context.Context, event domain.Event) error {
	ctx, span := otel.Tracer("mendo").Start(ctx, "EventBus.Publish."+event.GetEventType())
	defer span.End()

	b.logger.InfoContext(ctx, "event published",
		slog.String("event_type", event.GetEventType()),
		slog.String("aggregate_id", event.GetAggregateID()),
	)
	eventType := event.GetEventType()
	for i, handler := range b.handlers[eventType] {
		handlerName := fmt.Sprintf("%s_handler_%d", eventType, i)
		var lastErr error
		for attempt := 1; attempt <= b.maxRetries; attempt++ {
			if err := handler(ctx, event); err != nil {
				lastErr = err
				b.logger.WarnContext(ctx, "event handler retry",
					slog.Int("attempt", attempt),
					slog.Int("max_retries", b.maxRetries),
					slog.String("handler", handlerName),
					slog.String("error", err.Error()),
				)
				continue
			}
			lastErr = nil
			break
		}
		if lastErr != nil {
			letter := &domain.DeadLetter{
				ID:          uuid.New().String(),
				Event:       event,
				Error:       lastErr.Error(),
				FailCount:   b.maxRetries,
				LastFailAt:  time.Now(),
				HandlerName: handlerName,
			}
			if dlqErr := b.dlq.Store(ctx, letter); dlqErr != nil {
				return apperrors.Infrastructure("DLQ への保存に失敗", dlqErr)
			}
			b.logger.ErrorContext(ctx, "event sent to DLQ",
				slog.String("event_type", eventType),
				slog.String("handler", handlerName),
				slog.String("error", lastErr.Error()),
				slog.Int("fail_count", b.maxRetries),
			)
		}
	}
	return nil
}

func (b *WatermillEventBus) Subscribe(eventType string, handler EventHandler) {
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}
