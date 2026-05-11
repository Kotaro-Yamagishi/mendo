package testutil

import (
	"context"

	"mendo/internal/infrastructure/datasource"
)

type StubMenuDataSource struct {
	Menu         *datasource.MenuRow
	Menus        []datasource.MenuRow
	FindErr      error
	InsertErr    error
	UpdateErr    error
	InsertedMenu *datasource.MenuRow
	UpdatedID    string
	UpdatedAvail bool
}

func (s *StubMenuDataSource) FindMenuByID(_ context.Context, _ string) (*datasource.MenuRow, error) {
	if s.FindErr != nil {
		return nil, s.FindErr
	}
	return s.Menu, nil
}

func (s *StubMenuDataSource) FindAllMenus(_ context.Context) ([]datasource.MenuRow, error) {
	return s.Menus, nil
}

func (s *StubMenuDataSource) InsertMenu(_ context.Context, row *datasource.MenuRow) error {
	if s.InsertErr != nil {
		return s.InsertErr
	}
	s.InsertedMenu = row
	return nil
}

func (s *StubMenuDataSource) UpdateMenuAvailability(_ context.Context, id string, available bool) error {
	if s.UpdateErr != nil {
		return s.UpdateErr
	}
	s.UpdatedID = id
	s.UpdatedAvail = available
	return nil
}
