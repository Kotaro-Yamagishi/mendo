package testutil

import (
	"context"

	"mendo/internal/domain"
)

// SpyEventPublisher は EventPublisher の Spy。
type SpyEventPublisher struct {
	Published  []domain.Event
	PublishErr error
}

func (s *SpyEventPublisher) Publish(_ context.Context, events ...domain.Event) error {
	if s.PublishErr != nil {
		return s.PublishErr
	}
	s.Published = append(s.Published, events...)
	return nil
}
