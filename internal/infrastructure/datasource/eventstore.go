package datasource

import "context"

// EventStoreDataSource は events テーブルへのアクセス IF。
// append-only。UPDATE / DELETE は行わない。
// 引数・戻り値はすべてプリミティブ型または DTO。ドメイン型を使わない。
type EventStoreDataSource interface {
	// InsertEvents はイベント列を events テーブルに追記する。
	InsertEvents(ctx context.Context, rows []EventRow) error

	// FindEventsByAggregateID は aggregate_id でイベントを時系列順に全件取得する。
	FindEventsByAggregateID(ctx context.Context, aggregateID string) ([]EventRow, error)
}
