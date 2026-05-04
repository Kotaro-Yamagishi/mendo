package repository

import (
	"context"
	"fmt"
	"sync"

	"mendo/internal/domain/specialorder"
)

type InMemorySpecialOrderRepository struct {
	mu     sync.RWMutex
	orders map[string]*specialorder.SpecialOrder
}

func NewInMemorySpecialOrderRepository() *InMemorySpecialOrderRepository {
	return &InMemorySpecialOrderRepository{orders: make(map[string]*specialorder.SpecialOrder)}
}

func (r *InMemorySpecialOrderRepository) FindByID(_ context.Context, id specialorder.SpecialOrderID) (*specialorder.SpecialOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	so, ok := r.orders[string(id)]
	if !ok {
		return nil, fmt.Errorf("special order not found: %s", id)
	}
	return so, nil
}

func (r *InMemorySpecialOrderRepository) Save(_ context.Context, so *specialorder.SpecialOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[string(so.ID())] = so
	return nil
}
