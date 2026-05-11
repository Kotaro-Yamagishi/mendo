package testutil

import (
	"context"
	"fmt"

	"mendo/internal/infrastructure/datasource"
)

// StubStaffDataSource は datasource.StaffDataSource の Stub + Spy。
type StubStaffDataSource struct {
	Rows      []datasource.StaffRow
	FindErr   error
	AllErr    error
	UpsertErr error
	DeleteErr error

	UpsertedRow    *datasource.StaffRow
	DeletedID      string
}

func (s *StubStaffDataSource) FindStaffByID(_ context.Context, id string) (*datasource.StaffRow, error) {
	if s.FindErr != nil {
		return nil, s.FindErr
	}
	for i := range s.Rows {
		if s.Rows[i].ID == id {
			return &s.Rows[i], nil
		}
	}
	return nil, nil
}

func (s *StubStaffDataSource) FindAllStaffs(_ context.Context) ([]datasource.StaffRow, error) {
	if s.AllErr != nil {
		return nil, s.AllErr
	}
	return s.Rows, nil
}

func (s *StubStaffDataSource) UpsertStaff(_ context.Context, row *datasource.StaffRow) error {
	if s.UpsertErr != nil {
		return s.UpsertErr
	}
	s.UpsertedRow = row
	for i := range s.Rows {
		if s.Rows[i].ID == row.ID {
			s.Rows[i] = *row
			return nil
		}
	}
	s.Rows = append(s.Rows, *row)
	return nil
}

func (s *StubStaffDataSource) DeleteStaff(_ context.Context, id string) error {
	if s.DeleteErr != nil {
		return s.DeleteErr
	}
	for i := range s.Rows {
		if s.Rows[i].ID == id {
			s.Rows = append(s.Rows[:i], s.Rows[i+1:]...)
			s.DeletedID = id
			return nil
		}
	}
	return fmt.Errorf("staff not found: %s", id)
}
