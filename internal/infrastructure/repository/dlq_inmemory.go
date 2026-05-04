package repository

import (
	"context"
	"fmt"
	"sync"

	"mendo/internal/domain"
)

type InMemoryDLQ struct {
	mu      sync.RWMutex
	letters map[string]domain.DeadLetter
}

func NewInMemoryDLQ() *InMemoryDLQ {
	return &InMemoryDLQ{letters: make(map[string]domain.DeadLetter)}
}

func (d *InMemoryDLQ) Store(_ context.Context, letter *domain.DeadLetter) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.letters[letter.ID] = *letter
	return nil
}

func (d *InMemoryDLQ) List(_ context.Context) ([]domain.DeadLetter, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	result := make([]domain.DeadLetter, 0, len(d.letters))
	for _, l := range d.letters {
		result = append(result, l)
	}
	return result, nil
}

func (d *InMemoryDLQ) Remove(_ context.Context, id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.letters, id)
	return nil
}

func (d *InMemoryDLQ) FindByID(_ context.Context, id string) (*domain.DeadLetter, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	l, ok := d.letters[id]
	if !ok {
		return nil, fmt.Errorf("dead letter not found: %s", id)
	}
	return &l, nil
}
