package testutil

import (
	"context"
	"fmt"

	"mendo/internal/domain/order"
)

// StubProjectionReader は order.ProjectionReader の Stub。
type StubProjectionReader struct {
	Rows    []order.OrderStateRow
	Row     *order.OrderStateRow
	FindErr error
	AllErr  error
}

func (s *StubProjectionReader) FindByID(_ context.Context, orderID string) (*order.OrderStateRow, error) {
	if s.FindErr != nil {
		return nil, s.FindErr
	}
	if s.Row != nil {
		return s.Row, nil
	}
	return nil, fmt.Errorf("order not found: %s", orderID)
}

func (s *StubProjectionReader) FindAll(_ context.Context) ([]order.OrderStateRow, error) {
	if s.AllErr != nil {
		return nil, s.AllErr
	}
	return s.Rows, nil
}
