package eventbus

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"mendo/internal/domain"
)

type EventHandler func(ctx context.Context, event domain.Event) error

type WatermillEventBus struct {
	handlers   map[string][]EventHandler
	dlq        domain.DeadLetterQueue
	maxRetries int
}

func NewWatermillEventBus(dlq domain.DeadLetterQueue, maxRetries int) *WatermillEventBus {
	return &WatermillEventBus{
		handlers:   make(map[string][]EventHandler),
		dlq:        dlq,
		maxRetries: maxRetries,
	}
}

func (b *WatermillEventBus) Publish(ctx context.Context, events ...domain.Event) error {
	for _, event := range events {
		fmt.Printf("[EventBus] Published: %s\n", event.GetEventType())
		eventType := event.GetEventType()
		for i, handler := range b.handlers[eventType] {
			handlerName := fmt.Sprintf("%s_handler_%d", eventType, i)
			var lastErr error
			for attempt := 1; attempt <= b.maxRetries; attempt++ {
				if err := handler(ctx, event); err != nil {
					lastErr = err
					fmt.Printf("[EventBus] Retry %d/%d for %s: %v\n", attempt, b.maxRetries, handlerName, err)
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
					return fmt.Errorf("DLQ store failed: %w", dlqErr)
				}
				fmt.Printf("[EventBus] → DLQ: %s (handler: %s, error: %s)\n", eventType, handlerName, lastErr.Error())
			}
		}
	}
	return nil
}

func (b *WatermillEventBus) Subscribe(eventType string, handler EventHandler) {
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}
