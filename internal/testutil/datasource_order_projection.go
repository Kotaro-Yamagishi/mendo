package testutil

import (
	"context"
	"fmt"

	"mendo/internal/infrastructure/datasource"
)

// StubOrderProjectionDataSource は datasource.OrderProjectionDataSource の Stub + Spy。
type StubOrderProjectionDataSource struct {
	Row       *datasource.OrderProjectionRow
	Rows      []datasource.OrderProjectionRow
	FindErr   error
	AllErr    error
	UpsertErr error

	UpsertedRow *datasource.OrderProjectionRow
}

func (s *StubOrderProjectionDataSource) FindOrderProjectionByID(_ context.Context, orderID string) (*datasource.OrderProjectionRow, error) {
	if s.FindErr != nil {
		return nil, s.FindErr
	}
	if s.Row != nil {
		return s.Row, nil
	}
	for i := range s.Rows {
		if s.Rows[i].OrderID == orderID {
			return &s.Rows[i], nil
		}
	}
	return nil, fmt.Errorf("order not found: %s", orderID)
}

func (s *StubOrderProjectionDataSource) FindAllOrderProjections(_ context.Context) ([]datasource.OrderProjectionRow, error) {
	if s.AllErr != nil {
		return nil, s.AllErr
	}
	return s.Rows, nil
}

func (s *StubOrderProjectionDataSource) UpsertOrderProjection(_ context.Context, row *datasource.OrderProjectionRow) error {
	if s.UpsertErr != nil {
		return s.UpsertErr
	}
	s.UpsertedRow = row
	return nil
}
