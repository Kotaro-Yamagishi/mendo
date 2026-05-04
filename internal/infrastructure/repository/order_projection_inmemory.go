package repository

import (
	"context"
	"fmt"
	"sync"

	"mendo/internal/domain"
	"mendo/internal/domain/order"
)

// InMemoryOrderStateStore は Order の Projection をインメモリで管理する。
// 本番では PostgresOrderStateStore 等に差し替える。
//
// データの流れ:
//
//	events テーブル（INSERT only。真実の源）
//	  ↓ イベントが Publish される
//	subscriber が受け取る
//	  ↓ HandleEvent() を呼ぶ
//	order_projections（UPDATE される。読み取り用キャッシュ）
//	  ↓ FindByID() / FindAll() で読む
//	handler がレスポンスを返す
type InMemoryOrderStateStore struct {
	mu    sync.RWMutex
	store map[string]*order.OrderStateRow
}

func NewInMemoryOrderStateStore() *InMemoryOrderStateStore {
	return &InMemoryOrderStateStore{
		store: make(map[string]*order.OrderStateRow),
	}
}

// HandleEvent はドメインイベントを受け取って Projection を更新する。
// order.ProjectionWriter を満たす。
func (s *InMemoryOrderStateStore) HandleEvent(_ context.Context, event domain.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch e := event.(type) {
	case order.OrderCreated:
		s.store[e.GetAggregateID()] = &order.OrderStateRow{
			OrderID:   e.GetAggregateID(),
			SeatNo:    e.SeatNo,
			Status:    "pending",
			ItemCount: 0,
		}
		fmt.Printf("[Projection] INSERT: order_id=%s, seat_no=%d, status=pending\n", e.GetAggregateID(), e.SeatNo)

	case order.ItemAdded:
		if row, ok := s.store[e.GetAggregateID()]; ok {
			row.ItemCount++
			fmt.Printf("[Projection] UPDATE: order_id=%s, item_count=%d\n", e.GetAggregateID(), row.ItemCount)
		}

	case order.OrderConfirmed:
		if row, ok := s.store[e.GetAggregateID()]; ok {
			row.Status = "confirmed"
			fmt.Printf("[Projection] UPDATE: order_id=%s, status=confirmed\n", e.GetAggregateID())
		}

	case order.OrderCancelled:
		if row, ok := s.store[e.GetAggregateID()]; ok {
			row.Status = "canceled"
			fmt.Printf("[Projection] UPDATE: order_id=%s, status=canceled\n", e.GetAggregateID())
		}
	}
	return nil
}

// FindByID は Projection から注文の現在状態を読み取る。
// order.ProjectionReader を満たす。
func (s *InMemoryOrderStateStore) FindByID(_ context.Context, orderID string) (*order.OrderStateRow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row, ok := s.store[orderID]
	if !ok {
		return nil, fmt.Errorf("order not found: %s", orderID)
	}
	return row, nil
}

// FindAll は全注文の現在状態を取得する。
// order.ProjectionReader を満たす。
func (s *InMemoryOrderStateStore) FindAll(_ context.Context) ([]order.OrderStateRow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows := make([]order.OrderStateRow, 0, len(s.store))
	for _, row := range s.store {
		rows = append(rows, *row)
	}
	return rows, nil
}

// CountPending は未完了（pending）の注文数を返す。
// order.Reader を満たす。
func (s *InMemoryOrderStateStore) CountPending(_ context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, row := range s.store {
		if row.Status == "pending" {
			count++
		}
	}
	return count, nil
}
