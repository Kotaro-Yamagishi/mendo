package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"mendo/internal/infrastructure/datasource"
)

// DLQDataSource は datasource.DLQDataSource の MySQL 実装。
//
// テーブル: dlq (id, event_type, payload, error, fail_count, handler_name, last_fail_at)
type DLQDataSource struct {
	db *sql.DB
}

func NewDLQDataSource(db *sql.DB) *DLQDataSource {
	return &DLQDataSource{db: db}
}

// InsertDeadLetterRow はリトライ失敗したイベントを dlq テーブルに保存する。
func (d *DLQDataSource) InsertDeadLetterRow(ctx context.Context, row *datasource.DeadLetterRow) error {
	_, err := d.db.ExecContext(ctx,
		`INSERT INTO dlq (id, event_type, payload, error, fail_count, handler_name, last_fail_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		row.ID,
		row.EventType,
		row.Payload,
		row.Error,
		row.FailCount,
		row.HandlerName,
		row.LastFailAt,
	)
	if err != nil {
		return fmt.Errorf("insert dead letter row: %w", err)
	}
	return nil
}

// FindAllDeadLetterRows は dlq テーブルの全行を返す。
func (d *DLQDataSource) FindAllDeadLetterRows(ctx context.Context) ([]datasource.DeadLetterRow, error) {
	rows, err := d.db.QueryContext(ctx,
		`SELECT id, event_type, payload, error, fail_count, handler_name, last_fail_at
		   FROM dlq
		  ORDER BY last_fail_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("find all dead letter rows: %w", err)
	}
	defer rows.Close()

	var result []datasource.DeadLetterRow
	for rows.Next() {
		var row datasource.DeadLetterRow
		if err := rows.Scan(
			&row.ID,
			&row.EventType,
			&row.Payload,
			&row.Error,
			&row.FailCount,
			&row.HandlerName,
			&row.LastFailAt,
		); err != nil {
			return nil, fmt.Errorf("scan dead letter row: %w", err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dead letter rows: %w", err)
	}

	return result, nil
}

// FindDeadLetterRowByID は id を指定して DeadLetterRow を返す。
// 見つからない場合は nil, nil を返す。
func (d *DLQDataSource) FindDeadLetterRowByID(ctx context.Context, id string) (*datasource.DeadLetterRow, error) {
	var row datasource.DeadLetterRow
	err := d.db.QueryRowContext(ctx,
		`SELECT id, event_type, payload, error, fail_count, handler_name, last_fail_at
		   FROM dlq
		  WHERE id = ?`,
		id,
	).Scan(
		&row.ID,
		&row.EventType,
		&row.Payload,
		&row.Error,
		&row.FailCount,
		&row.HandlerName,
		&row.LastFailAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find dead letter row by id: %w", err)
	}
	return &row, nil
}

// DeleteDeadLetterRow は指定した id の行を削除する。
func (d *DLQDataSource) DeleteDeadLetterRow(ctx context.Context, id string) error {
	_, err := d.db.ExecContext(ctx, `DELETE FROM dlq WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete dead letter row: %w", err)
	}
	return nil
}
