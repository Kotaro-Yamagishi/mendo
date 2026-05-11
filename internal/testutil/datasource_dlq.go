package testutil

import (
	"context"
	"fmt"

	"mendo/internal/infrastructure/datasource"
)

// StubDLQDataSource は datasource.DLQDataSource の Stub + Spy。
type StubDLQDataSource struct {
	Rows        []datasource.DeadLetterRow
	SingleRow   *datasource.DeadLetterRow
	InsertErr   error
	FindAllErr  error
	FindByIDErr error
	DeleteErr   error

	InsertedRow *datasource.DeadLetterRow
	DeletedID   string
}

func (s *StubDLQDataSource) InsertDeadLetterRow(_ context.Context, row *datasource.DeadLetterRow) error {
	if s.InsertErr != nil {
		return s.InsertErr
	}
	s.InsertedRow = row
	return nil
}

func (s *StubDLQDataSource) FindAllDeadLetterRows(_ context.Context) ([]datasource.DeadLetterRow, error) {
	if s.FindAllErr != nil {
		return nil, s.FindAllErr
	}
	return s.Rows, nil
}

func (s *StubDLQDataSource) FindDeadLetterRowByID(_ context.Context, id string) (*datasource.DeadLetterRow, error) {
	if s.FindByIDErr != nil {
		return nil, s.FindByIDErr
	}
	if s.SingleRow != nil {
		return s.SingleRow, nil
	}
	// Rows から探す
	for i := range s.Rows {
		if s.Rows[i].ID == id {
			return &s.Rows[i], nil
		}
	}
	return nil, fmt.Errorf("dead letter not found: %s", id)
}

func (s *StubDLQDataSource) DeleteDeadLetterRow(_ context.Context, id string) error {
	if s.DeleteErr != nil {
		return s.DeleteErr
	}
	s.DeletedID = id
	return nil
}
