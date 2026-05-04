package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"mendo/internal/infrastructure/datasource"
)

// OrderBoardDataSource は datasource.OrderBoardDataSource の MySQL 実装。
//
// テーブル: kc_order_board (order_id PK, seat_no, order_status, cooking_status, ordered_at, cooking_at)
//
// Order と Kitchen の両方のイベントから構築される横断 Projection。
type OrderBoardDataSource struct {
	db *sql.DB
}

func NewOrderBoardDataSource(db *sql.DB) *OrderBoardDataSource {
	return &OrderBoardDataSource{db: db}
}

// FindOrderBoardRowByID は order_id を指定して OrderBoardRow を返す。
// 見つからない場合は nil, nil を返す。
func (b *OrderBoardDataSource) FindOrderBoardRowByID(ctx context.Context, orderID string) (*datasource.OrderBoardRow, error) {
	c := getConn(ctx, b.db)

	var (
		row       datasource.OrderBoardRow
		orderedAt sql.NullTime
		cookingAt sql.NullTime
	)
	err := c.QueryRowContext(ctx,
		`SELECT order_id, seat_no, order_status, cooking_status, ordered_at, cooking_at
		   FROM kc_order_board
		  WHERE order_id = ?`,
		orderID,
	).Scan(
		&row.OrderID,
		&row.SeatNo,
		&row.OrderStatus,
		&row.CookingStatus,
		&orderedAt,
		&cookingAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find order board row by id: %w", err)
	}

	if orderedAt.Valid {
		row.OrderedAt = &orderedAt.Time
	}
	if cookingAt.Valid {
		row.CookingAt = &cookingAt.Time
	}

	return &row, nil
}

// FindAllOrderBoardRows は kc_order_board テーブルの全行を返す。
func (b *OrderBoardDataSource) FindAllOrderBoardRows(ctx context.Context) ([]datasource.OrderBoardRow, error) {
	c := getConn(ctx, b.db)

	rows, err := c.QueryContext(ctx,
		`SELECT order_id, seat_no, order_status, cooking_status, ordered_at, cooking_at
		   FROM kc_order_board
		  ORDER BY ordered_at ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("find all order board rows: %w", err)
	}
	defer rows.Close()

	var result []datasource.OrderBoardRow
	for rows.Next() {
		var (
			row       datasource.OrderBoardRow
			orderedAt sql.NullTime
			cookingAt sql.NullTime
		)
		if err := rows.Scan(
			&row.OrderID,
			&row.SeatNo,
			&row.OrderStatus,
			&row.CookingStatus,
			&orderedAt,
			&cookingAt,
		); err != nil {
			return nil, fmt.Errorf("scan order board row: %w", err)
		}
		if orderedAt.Valid {
			row.OrderedAt = &orderedAt.Time
		}
		if cookingAt.Valid {
			row.CookingAt = &cookingAt.Time
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order board rows: %w", err)
	}

	return result, nil
}

// UpsertOrderBoardRow は OrderBoard エントリを INSERT OR UPDATE する。
func (b *OrderBoardDataSource) UpsertOrderBoardRow(ctx context.Context, row *datasource.OrderBoardRow) error {
	c := getConn(ctx, b.db)

	_, err := c.ExecContext(ctx,
		`INSERT INTO kc_order_board (order_id, seat_no, order_status, cooking_status, ordered_at, cooking_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   seat_no        = VALUES(seat_no),
		   order_status   = VALUES(order_status),
		   cooking_status = VALUES(cooking_status),
		   ordered_at     = VALUES(ordered_at),
		   cooking_at     = VALUES(cooking_at)`,
		row.OrderID,
		row.SeatNo,
		row.OrderStatus,
		row.CookingStatus,
		row.OrderedAt,
		row.CookingAt,
	)
	if err != nil {
		return fmt.Errorf("upsert order board row: %w", err)
	}
	return nil
}
