package repository

import (
	"context"
	"sync"

	"mendo/internal/domain"
)

// InMemoryEventStore は学習用のインメモリイベントストア。
// 本番では PostgreSQL, EventStoreDB, DynamoDB 等を使う。
//
// 1つのテーブル（map）に全集約のイベントをスタックする。
// aggregate_id でフィルタして取得する。
type InMemoryEventStore struct {
	mu     sync.RWMutex
	events map[string][]domain.Event // key = aggregate_id
}

func NewInMemoryEventStore() *InMemoryEventStore {
	return &InMemoryEventStore{
		events: make(map[string][]domain.Event),
	}
}

// Save はイベントをストアに追記する。Insert only。
func (s *InMemoryEventStore) Save(_ context.Context, events []domain.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, event := range events {
		aggID := event.GetAggregateID()
		s.events[aggID] = append(s.events[aggID], event)
	}
	return nil
}

// Load は指定した集約のイベント列を時系列順に取得する。
func (s *InMemoryEventStore) Load(_ context.Context, aggregateID string) ([]domain.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events, ok := s.events[aggregateID]
	if !ok {
		return nil, nil
	}

	// コピーを返す（呼び出し元が変更しても影響しない）
	result := make([]domain.Event, len(events))
	copy(result, events)
	return result, nil
}
