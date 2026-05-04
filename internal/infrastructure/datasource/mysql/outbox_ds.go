package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	datasource "mendo/internal/infrastructure/datasource"
	infraMysql "mendo/internal/infrastructure/mysql"
)

// MySQLOutboxDataSource は OutboxDataSource の MySQL 実装。
//
// テーブル: outbox (id, event_type, aggregate_id, payload, delivered, created_at)
type MySQLOutboxDataSource struct {
	db *sql.DB
}

func NewMySQLOutboxDataSource(db *sql.DB) *MySQLOutboxDataSource {
	return &MySQLOutboxDataSource{db: db}
}

// InsertOutboxRows はイベントを outbox テーブルに保存する。
// EventStore への INSERT と同一トランザクション内で呼ぶ。
func (ds *MySQLOutboxDataSource) InsertOutboxRows(ctx context.Context, rows []datasource.OutboxRow) error {
	c := infraMysql.GetConn(ctx, ds.db)

	for _, row := range rows {
		_, err := c.ExecContext(ctx,
			`INSERT INTO outbox (id, event_type, aggregate_id, payload, delivered, created_at)
			 VALUES (?, ?, ?, ?, FALSE, ?)`,
			row.ID,
			row.EventType,
			row.AggregateID,
			row.Payload,
			row.CreatedAt.UTC(),
		)
		if err != nil {
			return fmt.Errorf("insert outbox row %s: %w", row.ID, err)
		}
	}
	return nil
}

// FindUndeliveredOutboxRows は未配信（delivered=false）の行を limit 件取得する。
func (ds *MySQLOutboxDataSource) FindUndeliveredOutboxRows(ctx context.Context, limit int) ([]datasource.OutboxRow, error) {
	c := infraMysql.GetConn(ctx, ds.db)

	rows, err := c.QueryContext(ctx,
		`SELECT id, event_type, aggregate_id, payload, delivered, created_at
		   FROM outbox
		  WHERE delivered = FALSE
		  ORDER BY created_at ASC
		  LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("find undelivered outbox rows: %w", err)
	}
	defer rows.Close()

	var result []datasource.OutboxRow
	for rows.Next() {
		var row datasource.OutboxRow
		if err := rows.Scan(
			&row.ID,
			&row.EventType,
			&row.AggregateID,
			&row.Payload,
			&row.Delivered,
			&row.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan outbox row: %w", err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate outbox rows: %w", err)
	}
	return result, nil
}

// MarkOutboxRowsDelivered は指定した id 群を配信済み（delivered=true）にする。
func (ds *MySQLOutboxDataSource) MarkOutboxRowsDelivered(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	c := infraMysql.GetConn(ctx, ds.db)

	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1] // trailing comma を除去

	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	_, err := c.ExecContext(ctx,
		`UPDATE outbox SET delivered = TRUE WHERE id IN (`+placeholders+`)`,
		args...,
	)
	if err != nil {
		return fmt.Errorf("mark outbox rows delivered: %w", err)
	}
	return nil
}

// Verify interface compliance at compile time.
var _ datasource.OutboxDataSource = (*MySQLOutboxDataSource)(nil)
