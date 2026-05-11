package outbox_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"mendo/internal/domain"
	"mendo/internal/domain/order"
	"mendo/internal/infrastructure/outbox"
)

// safeOutbox は Fetch が実際のデータを返す Outbox Stub。
// goroutine から並行アクセスされるため mutex で保護する。
type safeOutbox struct {
	mu              sync.Mutex
	events          []domain.Event
	fetchErr        error
	markDeliveredCh [][]string
	markErr         error
}

func (s *safeOutbox) Store(_ context.Context, _ []domain.Event) error { return nil }

func (s *safeOutbox) Fetch(_ context.Context, _ int) ([]domain.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.fetchErr != nil {
		return nil, s.fetchErr
	}
	return s.events, nil
}

func (s *safeOutbox) MarkDelivered(_ context.Context, ids []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.markErr != nil {
		return s.markErr
	}
	s.markDeliveredCh = append(s.markDeliveredCh, ids)
	return nil
}

func (s *safeOutbox) markCalledCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.markDeliveredCh)
}

func (s *safeOutbox) firstMarkIDs() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.markDeliveredCh) == 0 {
		return nil
	}
	return s.markDeliveredCh[0]
}

// safePublisher は EventPublisher の Spy。goroutine から並行アクセスされるため mutex で保護する。
type safePublisher struct {
	mu         sync.Mutex
	published  []domain.Event
	publishErr error
}

func (s *safePublisher) Publish(_ context.Context, events ...domain.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.publishErr != nil {
		return s.publishErr
	}
	s.published = append(s.published, events...)
	return nil
}

func (s *safePublisher) publishedCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.published)
}

// stubDomainEvent は domain.Event の最小実装。
type stubDomainEvent struct {
	eventType   string
	aggregateID string
}

func (e stubDomainEvent) GetEventType() string     { return e.eventType }
func (e stubDomainEvent) GetAggregateID() string   { return e.aggregateID }
func (e stubDomainEvent) GetCorrelationID() string { return "" }

// runRelay は RelayService を起動して1サイクル以上実行し、cancel で停止するヘルパー。
// interval=10ms で起動し、25ms 待つことで確実に1回は ticker が発火する。
func runRelay(t *testing.T, ob domain.Outbox, pub domain.EventPublisher) {
	t.Helper()
	svc := outbox.NewRelayService(ob, pub, 10*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	svc.Start(ctx)
	time.Sleep(25 * time.Millisecond) // 少なくとも1回の ticker 発火を確認
	cancel()
	time.Sleep(5 * time.Millisecond) // goroutine 終了を待つ
}

// TestRelay_Normal は Fetch → Publish → MarkDelivered の正常系を確認する。
func TestRelay_Normal(t *testing.T) {
	events := []domain.Event{
		stubDomainEvent{eventType: order.EventTypeOrderCreated, aggregateID: "agg-1"},
		stubDomainEvent{eventType: order.EventTypeOrderCreated, aggregateID: "agg-2"},
	}
	ob := &safeOutbox{events: events}
	pub := &safePublisher{}

	runRelay(t, ob, pub)

	if pub.publishedCount() < 2 {
		t.Errorf("Published %d events, want at least 2", pub.publishedCount())
	}
	if ob.markCalledCount() == 0 {
		t.Error("MarkDelivered was not called")
	}
	ids := ob.firstMarkIDs()
	if len(ids) != 2 {
		t.Errorf("MarkDelivered received %d IDs, want 2", len(ids))
	}
}

// TestRelay_EmptyFetch は Fetch が0件の場合に Publish と MarkDelivered が呼ばれないことを確認する。
func TestRelay_EmptyFetch(t *testing.T) {
	ob := &safeOutbox{events: nil}
	pub := &safePublisher{}

	runRelay(t, ob, pub)

	if pub.publishedCount() != 0 {
		t.Errorf("Published %d events, want 0", pub.publishedCount())
	}
	if ob.markCalledCount() != 0 {
		t.Errorf("MarkDelivered called %d times, want 0", ob.markCalledCount())
	}
}

// TestRelay_PublishFail は Publish 失敗時に MarkDelivered が呼ばれないことを確認する。
func TestRelay_PublishFail(t *testing.T) {
	events := []domain.Event{
		stubDomainEvent{eventType: order.EventTypeOrderCreated, aggregateID: "agg-1"},
	}
	ob := &safeOutbox{events: events}
	pub := &safePublisher{publishErr: errors.New("publish failed")}

	runRelay(t, ob, pub)

	if ob.markCalledCount() != 0 {
		t.Errorf("MarkDelivered called %d times, want 0", ob.markCalledCount())
	}
}

// TestRelay_FetchFail は Fetch 失敗時に Publish と MarkDelivered が呼ばれないことを確認する。
// relay はエラーを fmt.Printf するのみで panic しない。
func TestRelay_FetchFail(t *testing.T) {
	ob := &safeOutbox{fetchErr: errors.New("db connection lost")}
	pub := &safePublisher{}

	// panic しないことを確認する
	runRelay(t, ob, pub)

	if pub.publishedCount() != 0 {
		t.Errorf("Published %d events, want 0", pub.publishedCount())
	}
}
