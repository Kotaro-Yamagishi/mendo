package testutil

import (
	"context"
	"mendo/internal/infrastructure/datasource"
)

// StubEventStoreDataSource は EventStoreDataSource の Stub + Spy。
// datasource 層のインターフェースに対する Stub で、repository 層のユニットテストで使う。
type StubEventStoreDataSource struct {
	Events    []datasource.EventRow // FindEventsByAggregateID で返す行
	FindErr   error                 // FindEventsByAggregateID で返すエラー
	InsertErr error                 // InsertEvents で返すエラー
	Inserted  []datasource.EventRow // InsertEvents で受け取った行（Spy）
}

func (s *StubEventStoreDataSource) FindEventsByAggregateID(_ context.Context, _ string) ([]datasource.EventRow, error) {
	if s.FindErr != nil {
		return nil, s.FindErr
	}
	return s.Events, nil
}

func (s *StubEventStoreDataSource) InsertEvents(_ context.Context, rows []datasource.EventRow) error {
	if s.InsertErr != nil {
		return s.InsertErr
	}
	s.Inserted = append(s.Inserted, rows...)
	return nil
}
