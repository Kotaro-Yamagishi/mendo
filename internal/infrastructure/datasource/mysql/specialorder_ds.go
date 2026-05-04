package mysql

import (
	"context"
	"database/sql"
	"fmt"

	datasource "mendo/internal/infrastructure/datasource"
	infraMysql "mendo/internal/infrastructure/mysql"
)

// MySQLSpecialOrderDataSource は SpecialOrderDataSource の MySQL 実装。
//
// テーブル: sc_special_orders (id, order_id, menu_name, status, reject_reason, suggested_menu, created_at, updated_at)
type MySQLSpecialOrderDataSource struct {
	db *sql.DB
}

func NewMySQLSpecialOrderDataSource(db *sql.DB) *MySQLSpecialOrderDataSource {
	return &MySQLSpecialOrderDataSource{db: db}
}

// FindSpecialOrderByID は sc_special_orders から id を指定して SpecialOrderRow を返す。
// 見つからない場合は nil, nil を返す。
func (ds *MySQLSpecialOrderDataSource) FindSpecialOrderByID(ctx context.Context, id string) (*datasource.SpecialOrderRow, error) {
	c := infraMysql.GetConn(ctx, ds.db)

	var row datasource.SpecialOrderRow
	err := c.QueryRowContext(ctx,
		`SELECT id, order_id, menu_name, status, reject_reason, suggested_menu, created_at, updated_at
		   FROM sc_special_orders
		  WHERE id = ?`,
		id,
	).Scan(
		&row.ID,
		&row.OrderID,
		&row.MenuName,
		&row.Status,
		&row.RejectReason,
		&row.SuggestedMenu,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find special order by id: %w", err)
	}
	return &row, nil
}

// UpsertSpecialOrder は sc_special_orders を INSERT OR UPDATE する。
func (ds *MySQLSpecialOrderDataSource) UpsertSpecialOrder(ctx context.Context, row *datasource.SpecialOrderRow) error {
	c := infraMysql.GetConn(ctx, ds.db)

	_, err := c.ExecContext(ctx,
		`INSERT INTO sc_special_orders (id, order_id, menu_name, status, reject_reason, suggested_menu, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   menu_name      = VALUES(menu_name),
		   status         = VALUES(status),
		   reject_reason  = VALUES(reject_reason),
		   suggested_menu = VALUES(suggested_menu),
		   updated_at     = VALUES(updated_at)`,
		row.ID,
		row.OrderID,
		row.MenuName,
		row.Status,
		row.RejectReason,
		row.SuggestedMenu,
		row.CreatedAt.UTC(),
		row.UpdatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("upsert special order: %w", err)
	}
	return nil
}
