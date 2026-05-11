package testutil

import (
	"context"
	"errors"

	"mendo/internal/domain/specialorder"
)

// StubSpecialOrderReader は specialorder.Reader の Stub。
type StubSpecialOrderReader struct {
	SpecialOrder *specialorder.SpecialOrder
	FindErr      error
}

func (s *StubSpecialOrderReader) FindByID(_ context.Context, _ specialorder.SpecialOrderID) (*specialorder.SpecialOrder, error) {
	if s.FindErr != nil {
		return nil, s.FindErr
	}
	if s.SpecialOrder == nil {
		return nil, errors.New("not found")
	}
	return s.SpecialOrder, nil
}

// SpySpecialOrderWriter は specialorder.Writer の Spy。
type SpySpecialOrderWriter struct {
	SavedSpecialOrder *specialorder.SpecialOrder
	SaveErr           error
}

func (s *SpySpecialOrderWriter) Save(_ context.Context, so *specialorder.SpecialOrder) error {
	if s.SaveErr != nil {
		return s.SaveErr
	}
	s.SavedSpecialOrder = so
	return nil
}
