package repository

import (
	"context"
	"fmt"
	"sync"

	"mendo/internal/domain/kitchen"
)

type InMemoryKitchenRepository struct {
	mu       sync.RWMutex
	kitchens map[string]*kitchen.Kitchen
}

func NewInMemoryKitchenRepository() *InMemoryKitchenRepository {
	return &InMemoryKitchenRepository{kitchens: make(map[string]*kitchen.Kitchen)}
}

func (r *InMemoryKitchenRepository) FindByID(_ context.Context, id kitchen.KitchenID) (*kitchen.Kitchen, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	k, ok := r.kitchens[string(id)]
	if !ok {
		return nil, fmt.Errorf("kitchen not found: %s", id)
	}
	return k, nil
}

func (r *InMemoryKitchenRepository) Save(_ context.Context, k *kitchen.Kitchen) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.kitchens[k.ID().String()] = k
	return nil
}
