package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"mendo/internal/infrastructure/datasource"
)

// OrderProjectionDataSource は datasource.OrderProjectionDataSource の MySQL 実装。
//
// テーブル: oc_order_projections (order_id PK, seat_no, items, total, status, created_at, updated_at)
//
// データの流れ:
//
//	OrderCreated / ItemAdded / OrderConfirmed / OrderCancelled イベント
//	  ↓ UpsertOrderProjection
//	oc_order_projections に INSERT or UPDATE
//	  ↓ FindOrderProjectionByID / FindAllOrderProjections
//	handler へレスポンス
type OrderProjectionDataSource struct {
	db *sql.DB
}

func NewOrderProjectionDataSource(db *sql.DB) *OrderProjectionDataSource {
	return &OrderProjectionDataSource{db: db}
}

// FindOrderProjectionByID は order_id を指定して OrderProjectionRow を返す。
// 見つからない場合は nil, nil を返す。
func (p *OrderProjectionDataSource) FindOrderProjectionByID(ctx context.Context, orderID string) (*datasource.OrderProjectionRow, error) {
	c := getConn(ctx, p.db)

	var row datasource.OrderProjectionRow
	err := c.QueryRowContext(ctx,
		`SELECT order_id, seat_no, items, total, status, created_at, updated_at
		   FROM oc_order_projections
		  WHERE order_id = ?`,
		orderID,
	).Scan(
		&row.OrderID,
		&row.SeatNo,
		&row.Items,
		&row.Total,
		&row.Status,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find order projection by id: %w", err)
	}
	return &row, nil
}

// FindAllOrderProjections は全 Order の Projection を返す。
func (p *OrderProjectionDataSource) FindAllOrderProjections(ctx context.Context) ([]datasource.OrderProjectionRow, error) {
	c := getConn(ctx, p.db)

	rows, err := c.QueryContext(ctx,
		`SELECT order_id, seat_no, items, total, status, created_at, updated_at
		   FROM oc_order_projections
		  ORDER BY order_id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("find all order projections: %w", err)
	}
	defer rows.Close()

	var result []datasource.OrderProjectionRow
	for rows.Next() {
		var row datasource.OrderProjectionRow
		if err := rows.Scan(
			&row.OrderID,
			&row.SeatNo,
			&row.Items,
			&row.Total,
			&row.Status,
			&row.CreatedAt,
			&row.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan order projection: %w", err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order projections: %w", err)
	}

	return result, nil
}

// UpsertOrderProjection は Order Projection を INSERT OR UPDATE する。
func (p *OrderProjectionDataSource) UpsertOrderProjection(ctx context.Context, row *datasource.OrderProjectionRow) error {
	c := getConn(ctx, p.db)

	_, err := c.ExecContext(ctx,
		`INSERT INTO oc_order_projections (order_id, seat_no, items, total, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   seat_no    = VALUES(seat_no),
		   items      = VALUES(items),
		   total      = VALUES(total),
		   status     = VALUES(status),
		   updated_at = VALUES(updated_at)`,
		row.OrderID,
		row.SeatNo,
		row.Items,
		row.Total,
		row.Status,
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert order projection: %w", err)
	}
	return nil
}
