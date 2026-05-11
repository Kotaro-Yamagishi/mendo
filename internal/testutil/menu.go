package testutil

import (
	"context"
	"errors"
	"sync"

	"mendo/internal/domain/menu"
)

// StubMenuReader は menu.Reader の Stub。
type StubMenuReader struct {
	Menu    *menu.Menu
	Menus   []*menu.Menu
	FindErr error
}

func (s *StubMenuReader) FindByID(_ context.Context, _ menu.MenuID) (*menu.Menu, error) {
	if s.FindErr != nil {
		return nil, s.FindErr
	}
	if s.Menu == nil {
		return nil, errors.New("not found")
	}
	return s.Menu, nil
}

func (s *StubMenuReader) FindAll(_ context.Context) ([]*menu.Menu, error) {
	return s.Menus, nil
}

// SpyMenuWriter は menu.Writer の Spy。
// processChunk は goroutine で並行実行されるため mutex で保護する。
type SpyMenuWriter struct {
	mu        sync.Mutex
	SavedMenu *menu.Menu
	SaveErr   error
}

func (s *SpyMenuWriter) Save(_ context.Context, m *menu.Menu) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.SaveErr != nil {
		return s.SaveErr
	}
	s.SavedMenu = m
	return nil
}
