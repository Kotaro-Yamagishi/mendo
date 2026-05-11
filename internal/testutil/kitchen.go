package testutil

import (
	"context"
	"errors"

	"mendo/internal/domain/kitchen"
)

// StubKitchenReader は kitchen.Reader の Stub。
type StubKitchenReader struct {
	Kitchen *kitchen.Kitchen
	FindErr error
}

func (s *StubKitchenReader) FindByID(_ context.Context, _ kitchen.KitchenID) (*kitchen.Kitchen, error) {
	if s.FindErr != nil {
		return nil, s.FindErr
	}
	if s.Kitchen == nil {
		return nil, errors.New("not found")
	}
	return s.Kitchen, nil
}

// SpyKitchenWriter は kitchen.Writer の Spy。
type SpyKitchenWriter struct {
	SavedKitchen *kitchen.Kitchen
	SaveErr      error
}

func (s *SpyKitchenWriter) Save(_ context.Context, k *kitchen.Kitchen) error {
	if s.SaveErr != nil {
		return s.SaveErr
	}
	s.SavedKitchen = k
	return nil
}
