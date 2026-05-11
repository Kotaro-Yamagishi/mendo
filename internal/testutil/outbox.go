package testutil

import (
	"context"

	"mendo/internal/domain"
)

// SpyOutbox は Outbox の Spy。Store で受け取ったイベントを記録する。
type SpyOutbox struct {
	Stored   []domain.Event // Store で受け取ったイベント
	StoreErr error          // Store で返すエラー
}

func (s *SpyOutbox) Store(_ context.Context, events []domain.Event) error {
	if s.StoreErr != nil {
		return s.StoreErr
	}
	s.Stored = append(s.Stored, events...)
	return nil
}

func (s *SpyOutbox) Fetch(_ context.Context, _ int) ([]domain.Event, error) {
	return nil, nil
}

func (s *SpyOutbox) MarkDelivered(_ context.Context, _ []string) error {
	return nil
}
