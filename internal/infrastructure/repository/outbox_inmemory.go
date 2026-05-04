package repository

import (
	"context"
	"sync"

	"mendo/internal/domain"
)

// InMemoryOutbox は学習用のインメモリ送信箱。
// 本番では DB テーブル（outbox テーブル）で実装する。
type InMemoryOutbox struct {
	mu      sync.Mutex
	pending []outboxEntry
}

type outboxEntry struct {
	event     domain.Event
	delivered bool
}

func NewInMemoryOutbox() *InMemoryOutbox {
	return &InMemoryOutbox{}
}

// Store はイベントを送信箱に保存する。
// 本番では eventStore.Save() と同じ DB トランザクション内で呼ぶ。
func (o *InMemoryOutbox) Store(_ context.Context, events []domain.Event) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	for _, event := range events {
		o.pending = append(o.pending, outboxEntry{event: event})
	}
	return nil
}

// Fetch は未配信のイベントを取得する。
func (o *InMemoryOutbox) Fetch(_ context.Context, limit int) ([]domain.Event, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	var result []domain.Event
	for _, entry := range o.pending {
		if !entry.delivered {
			result = append(result, entry.event)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

// MarkDelivered は配信済みにマークする。
func (o *InMemoryOutbox) MarkDelivered(_ context.Context, eventIDs []string) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	idSet := make(map[string]bool)
	for _, id := range eventIDs {
		idSet[id] = true
	}
	for i := range o.pending {
		if idSet[o.pending[i].event.GetAggregateID()] {
			o.pending[i].delivered = true
		}
	}
	return nil
}
