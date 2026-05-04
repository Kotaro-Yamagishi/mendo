package datasource

import "context"

// OutboxDataSource は outbox テーブルへのアクセス IF。
// 引数・戻り値はすべてプリミティブ型または DTO。ドメイン型を使わない。
type OutboxDataSource interface {
	// InsertOutboxRows はイベントを outbox テーブルに保存する。
	// EventStore への INSERT と同一トランザクション内で呼ぶ。
	InsertOutboxRows(ctx context.Context, rows []OutboxRow) error

	// FindUndeliveredOutboxRows は未配信（delivered=false）の行を limit 件取得する。
	FindUndeliveredOutboxRows(ctx context.Context, limit int) ([]OutboxRow, error)

	// MarkOutboxRowsDelivered は指定した id 群を配信済み（delivered=true）にする。
	MarkOutboxRowsDelivered(ctx context.Context, ids []string) error
}
