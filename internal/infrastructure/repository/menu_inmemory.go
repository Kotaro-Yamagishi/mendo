package repository

import (
	"context"
	"sync"

	"mendo/internal/apperrors"
	"mendo/internal/domain/menu"
)

type InMemoryMenuRepository struct {
	mu    sync.RWMutex
	menus map[string]*menu.Menu
}

func NewInMemoryMenuRepository() *InMemoryMenuRepository {
	return &InMemoryMenuRepository{menus: make(map[string]*menu.Menu)}
}

func (r *InMemoryMenuRepository) FindByID(_ context.Context, id menu.MenuID) (*menu.Menu, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.menus[string(id)]
	if !ok {
		return nil, apperrors.NotFound("menu", string(id))
	}
	return m, nil
}

func (r *InMemoryMenuRepository) FindAll(_ context.Context) ([]*menu.Menu, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*menu.Menu, 0, len(r.menus))
	for _, m := range r.menus {
		result = append(result, m)
	}
	return result, nil
}

func (r *InMemoryMenuRepository) Save(_ context.Context, m *menu.Menu) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.menus[m.ID().String()] = m
	return nil
}
