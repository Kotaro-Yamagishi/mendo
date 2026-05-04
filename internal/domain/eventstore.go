package domain

import "context"

// EventStore はイベントソーシング用のストア。
// 全集約のイベントを1つのストアに保存する（aggregate_id でフィルタ）。
// Insert only。UPDATE も DELETE もしない。
type EventStore interface {
	// Save はイベントをストアに追記する。
	Save(ctx context.Context, events []Event) error

	// Load は指定した集約のイベント列を時系列順に全て取得する。
	// この列を順に Apply することで集約の現在の状態を復元する。
	Load(ctx context.Context, aggregateID string) ([]Event, error)
}
