package mysql

import (
	"context"
	"database/sql"
	"fmt"

	datasource "mendo/internal/infrastructure/datasource"
	infraMysql "mendo/internal/infrastructure/mysql"
)

// MySQLEventStoreDataSource は EventStoreDataSource の MySQL 実装。
//
// テーブル: events (event_id, aggregate_id, aggregate_type, event_type, version, payload, correlation_id, causation_id, created_at)
// append-only。UPDATE / DELETE は行わない。
type MySQLEventStoreDataSource struct {
	db *sql.DB
}

func NewMySQLEventStoreDataSource(db *sql.DB) *MySQLEventStoreDataSource {
	return &MySQLEventStoreDataSource{db: db}
}

// InsertEvents はイベント列を events テーブルに追記する。
func (ds *MySQLEventStoreDataSource) InsertEvents(ctx context.Context, rows []datasource.EventRow) error {
	c := infraMysql.GetConn(ctx, ds.db)

	for _, row := range rows {
		_, err := c.ExecContext(ctx,
			`INSERT INTO events
			   (event_id, aggregate_id, aggregate_type, event_type, version, payload, correlation_id, causation_id, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			row.EventID,
			row.AggregateID,
			row.AggregateType,
			row.EventType,
			row.Version,
			row.Payload,
			row.CorrelationID,
			row.CausationID,
			row.CreatedAt.UTC(),
		)
		if err != nil {
			return fmt.Errorf("insert event %s: %w", row.EventID, err)
		}
	}
	return nil
}

// FindEventsByAggregateID は aggregate_id でイベントを時系列順に全件取得する。
func (ds *MySQLEventStoreDataSource) FindEventsByAggregateID(ctx context.Context, aggregateID string) ([]datasource.EventRow, error) {
	c := infraMysql.GetConn(ctx, ds.db)

	rows, err := c.QueryContext(ctx,
		`SELECT event_id, aggregate_id, aggregate_type, event_type, version, payload, correlation_id, causation_id, created_at
		   FROM events
		  WHERE aggregate_id = ?
		  ORDER BY version ASC, created_at ASC`,
		aggregateID,
	)
	if err != nil {
		return nil, fmt.Errorf("find events by aggregate id: %w", err)
	}
	defer rows.Close()

	var result []datasource.EventRow
	for rows.Next() {
		var row datasource.EventRow
		if err := rows.Scan(
			&row.EventID,
			&row.AggregateID,
			&row.AggregateType,
			&row.EventType,
			&row.Version,
			&row.Payload,
			&row.CorrelationID,
			&row.CausationID,
			&row.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}
	return result, nil
}

// Verify interface compliance at compile time.
var _ datasource.EventStoreDataSource = (*MySQLEventStoreDataSource)(nil)
