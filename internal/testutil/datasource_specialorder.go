package testutil

import (
	"context"

	"mendo/internal/infrastructure/datasource"
)

type StubSpecialOrderDataSource struct {
	SpecialOrder *datasource.SpecialOrderRow
	FindErr      error
	UpsertErr    error
	UpsertedRow  *datasource.SpecialOrderRow
}

func (s *StubSpecialOrderDataSource) FindSpecialOrderByID(_ context.Context, _ string) (*datasource.SpecialOrderRow, error) {
	if s.FindErr != nil {
		return nil, s.FindErr
	}
	return s.SpecialOrder, nil
}

func (s *StubSpecialOrderDataSource) UpsertSpecialOrder(_ context.Context, row *datasource.SpecialOrderRow) error {
	if s.UpsertErr != nil {
		return s.UpsertErr
	}
	s.UpsertedRow = row
	return nil
}
