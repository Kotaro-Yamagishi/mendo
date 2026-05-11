package testutil

import (
	"context"

	"mendo/internal/infrastructure/datasource"
)

// StubKitchenDataSource は datasource.KitchenDataSource の Stub + Spy。
type StubKitchenDataSource struct {
	KitchenRow    *datasource.KitchenRow
	TaskRows      []datasource.CookingTaskRow
	FindErr       error
	FindTasksErr  error
	UpsertErr     error
	InsertTaskErr error
	UpdateTaskErr error

	UpsertedKitchen  *datasource.KitchenRow
	InsertedTask     *datasource.CookingTaskRow
	UpdatedKitchenID string
	UpdatedOrderID   string
	UpdatedStatus    string
}

func (s *StubKitchenDataSource) FindKitchenByID(_ context.Context, _ string) (*datasource.KitchenRow, error) {
	if s.FindErr != nil {
		return nil, s.FindErr
	}
	return s.KitchenRow, nil
}

func (s *StubKitchenDataSource) FindCookingTasksByKitchenID(_ context.Context, _ string) ([]datasource.CookingTaskRow, error) {
	if s.FindTasksErr != nil {
		return nil, s.FindTasksErr
	}
	return s.TaskRows, nil
}

func (s *StubKitchenDataSource) UpsertKitchen(_ context.Context, row *datasource.KitchenRow) error {
	if s.UpsertErr != nil {
		return s.UpsertErr
	}
	s.UpsertedKitchen = row
	return nil
}

func (s *StubKitchenDataSource) InsertCookingTask(_ context.Context, row *datasource.CookingTaskRow) error {
	if s.InsertTaskErr != nil {
		return s.InsertTaskErr
	}
	s.InsertedTask = row
	return nil
}

func (s *StubKitchenDataSource) UpdateCookingTaskStatus(_ context.Context, kitchenID, orderID, status string) error {
	if s.UpdateTaskErr != nil {
		return s.UpdateTaskErr
	}
	s.UpdatedKitchenID = kitchenID
	s.UpdatedOrderID = orderID
	s.UpdatedStatus = status
	return nil
}
