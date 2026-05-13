package eventbus_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"mendo/internal/domain"
	"mendo/internal/domain/order"
	"mendo/internal/infrastructure/eventbus"
	"mendo/internal/testutil"
)

// stubEvent は domain.Event の最小 Stub。
type stubEvent struct {
	domain.DomainEvent
}

func (e stubEvent) GetEventType() string    { return e.DomainEvent.EventType }
func (e stubEvent) GetAggregateID() string  { return e.DomainEvent.AggregateID }
func (e stubEvent) GetCorrelationID() string { return e.DomainEvent.CorrelationID }

func newStubEvent(eventType, aggregateID string) stubEvent {
	return stubEvent{
		DomainEvent: domain.NewDomainEvent(aggregateID, "TestAggregate", eventType, "corr-1"),
	}
}

// TestSubscribePublish_HandlerCalled はハンドラが呼ばれることを確認する。
func TestSubscribePublish_HandlerCalled(t *testing.T) {
	dlq := &testutil.StubDeadLetterQueue{}
	bus := eventbus.NewWatermillEventBus(dlq, 3, slog.Default())

	called := 0
	bus.Subscribe(order.EventTypeOrderCreated, func(_ context.Context, _ domain.Event) error {
		called++
		return nil
	})

	ev := newStubEvent(order.EventTypeOrderCreated, "agg-1")
	if err := bus.Publish(context.Background(), ev); err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}

	if called != 1 {
		t.Errorf("handler called %d times, want 1", called)
	}
}

// TestPublish_HandlerFailure_Retried はハンドラ失敗時に maxRetries 回再試行されることを確認する。
func TestPublish_HandlerFailure_Retried(t *testing.T) {
	dlq := &testutil.StubDeadLetterQueue{}
	maxRetries := 3
	bus := eventbus.NewWatermillEventBus(dlq, maxRetries, slog.Default())

	attempts := 0
	bus.Subscribe(order.EventTypeOrderCreated, func(_ context.Context, _ domain.Event) error {
		attempts++
		return errors.New("handler error")
	})

	ev := newStubEvent(order.EventTypeOrderCreated, "agg-1")
	if err := bus.Publish(context.Background(), ev); err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}

	if attempts != maxRetries {
		t.Errorf("handler attempted %d times, want %d", attempts, maxRetries)
	}
}

// TestPublish_AllRetriesFail_StoredInDLQ は全リトライ失敗後に DLQ に保存されることを確認する。
func TestPublish_AllRetriesFail_StoredInDLQ(t *testing.T) {
	dlq := &testutil.StubDeadLetterQueue{}
	maxRetries := 3
	bus := eventbus.NewWatermillEventBus(dlq, maxRetries, slog.Default())

	bus.Subscribe(order.EventTypeOrderConfirmed, func(_ context.Context, _ domain.Event) error {
		return errors.New("permanent failure")
	})

	ev := newStubEvent(order.EventTypeOrderConfirmed, "agg-2")
	if err := bus.Publish(context.Background(), ev); err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}

	if len(dlq.StoredLetters) != 1 {
		t.Fatalf("DLQ stored %d letters, want 1", len(dlq.StoredLetters))
	}
}

// TestPublish_AllRetriesFail_DeadLetterContent は DLQ に保存された DeadLetter の内容を検証する。
func TestPublish_AllRetriesFail_DeadLetterContent(t *testing.T) {
	dlq := &testutil.StubDeadLetterQueue{}
	maxRetries := 2
	bus := eventbus.NewWatermillEventBus(dlq, maxRetries, slog.Default())

	handlerErr := errors.New("handler exploded")
	bus.Subscribe("kitchen.started", func(_ context.Context, _ domain.Event) error {
		return handlerErr
	})

	ev := newStubEvent("kitchen.started", "agg-3")
	if err := bus.Publish(context.Background(), ev); err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}

	if len(dlq.StoredLetters) != 1 {
		t.Fatalf("DLQ stored %d letters, want 1", len(dlq.StoredLetters))
	}

	letter := dlq.StoredLetters[0]

	if letter.FailCount != maxRetries {
		t.Errorf("FailCount = %d, want %d", letter.FailCount, maxRetries)
	}
	if letter.Error != handlerErr.Error() {
		t.Errorf("Error = %q, want %q", letter.Error, handlerErr.Error())
	}
	if letter.HandlerName != "kitchen.started_handler_0" {
		t.Errorf("HandlerName = %q, want %q", letter.HandlerName, "kitchen.started_handler_0")
	}
	if letter.ID == "" {
		t.Error("DeadLetter ID must not be empty")
	}
}
