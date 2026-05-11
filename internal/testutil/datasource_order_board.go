package testutil

import (
	"context"

	"mendo/internal/infrastructure/datasource"
)

// StubOrderBoardDataSource は datasource.OrderBoardDataSource の Stub + Spy。
type StubOrderBoardDataSource struct {
	Row       *datasource.OrderBoardRow
	Rows      []datasource.OrderBoardRow
	FindErr   error
	AllErr    error
	UpsertErr error

	UpsertedRow *datasource.OrderBoardRow
}

func (s *StubOrderBoardDataSource) FindOrderBoardRowByID(_ context.Context, orderID string) (*datasource.OrderBoardRow, error) {
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
	return nil, nil
}

func (s *StubOrderBoardDataSource) FindAllOrderBoardRows(_ context.Context) ([]datasource.OrderBoardRow, error) {
	if s.AllErr != nil {
		return nil, s.AllErr
	}
	return s.Rows, nil
}

func (s *StubOrderBoardDataSource) UpsertOrderBoardRow(_ context.Context, row *datasource.OrderBoardRow) error {
	if s.UpsertErr != nil {
		return s.UpsertErr
	}
	s.UpsertedRow = row
	// Rows にも反映（Find系のラウンドトリップを支えるため）
	for i := range s.Rows {
		if s.Rows[i].OrderID == row.OrderID {
			s.Rows[i] = *row
			return nil
		}
	}
	s.Rows = append(s.Rows, *row)
	return nil
}
