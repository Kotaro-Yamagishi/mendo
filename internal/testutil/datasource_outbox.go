package testutil

import (
	"context"

	"mendo/internal/infrastructure/datasource"
)

type StubOutboxDataSource struct {
	UndeliveredRows []datasource.OutboxRow
	InsertErr       error
	FindErr         error
	MarkErr         error
	InsertedRows    []datasource.OutboxRow
	MarkedIDs       []string
}

func (s *StubOutboxDataSource) InsertOutboxRows(_ context.Context, rows []datasource.OutboxRow) error {
	if s.InsertErr != nil {
		return s.InsertErr
	}
	s.InsertedRows = append(s.InsertedRows, rows...)
	return nil
}

func (s *StubOutboxDataSource) FindUndeliveredOutboxRows(_ context.Context, _ int) ([]datasource.OutboxRow, error) {
	if s.FindErr != nil {
		return nil, s.FindErr
	}
	return s.UndeliveredRows, nil
}

func (s *StubOutboxDataSource) MarkOutboxRowsDelivered(_ context.Context, ids []string) error {
	if s.MarkErr != nil {
		return s.MarkErr
	}
	s.MarkedIDs = append(s.MarkedIDs, ids...)
	return nil
}
