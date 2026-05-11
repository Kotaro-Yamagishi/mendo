package testutil

import (
	"context"

	"mendo/internal/domain"
)

// StubEventStore はイベントストアの Stub + Spy。
// Load で返す値を設定でき（Stub）、Save で受け取った値を記録する（Spy）。
type StubEventStore struct {
	Events  []domain.Event // Load で返すイベント列
	LoadErr error          // Load で返すエラー
	SaveErr error          // Save で返すエラー
	Saved   []domain.Event // Save で受け取ったイベント（Spy）
}

func (s *StubEventStore) Load(_ context.Context, _ string) ([]domain.Event, error) {
	if s.LoadErr != nil {
		return nil, s.LoadErr
	}
	return s.Events, nil
}

func (s *StubEventStore) Save(_ context.Context, events []domain.Event) error {
	if s.SaveErr != nil {
		return s.SaveErr
	}
	s.Saved = append(s.Saved, events...)
	return nil
}
