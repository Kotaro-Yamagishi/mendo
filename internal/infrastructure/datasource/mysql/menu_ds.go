package mysql

import (
	"context"
	"database/sql"
	"fmt"

	datasource "mendo/internal/infrastructure/datasource"
	infraMysql "mendo/internal/infrastructure/mysql"
)

// MySQLMenuDataSource は MenuDataSource の MySQL 実装。
//
// テーブル: oc_menus (menu_id, name, price, available, created_at, updated_at)
type MySQLMenuDataSource struct {
	db *sql.DB
}

func NewMySQLMenuDataSource(db *sql.DB) *MySQLMenuDataSource {
	return &MySQLMenuDataSource{db: db}
}

// FindMenuByID は oc_menus から menu_id を指定して MenuRow を返す。
// 見つからない場合は nil, nil を返す。
func (ds *MySQLMenuDataSource) FindMenuByID(ctx context.Context, id string) (*datasource.MenuRow, error) {
	c := infraMysql.GetConn(ctx, ds.db)

	var row datasource.MenuRow
	err := c.QueryRowContext(ctx,
		`SELECT menu_id, name, price, available, created_at, updated_at FROM oc_menus WHERE menu_id = ?`,
		id,
	).Scan(&row.MenuID, &row.Name, &row.Price, &row.Available, &row.CreatedAt, &row.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find menu by id: %w", err)
	}
	return &row, nil
}

// FindAllMenus は oc_menus の全メニューを返す。
func (ds *MySQLMenuDataSource) FindAllMenus(ctx context.Context) ([]datasource.MenuRow, error) {
	c := infraMysql.GetConn(ctx, ds.db)

	rows, err := c.QueryContext(ctx,
		`SELECT menu_id, name, price, available, created_at, updated_at FROM oc_menus ORDER BY name ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("find all menus: %w", err)
	}
	defer rows.Close()

	var result []datasource.MenuRow
	for rows.Next() {
		var row datasource.MenuRow
		if err := rows.Scan(&row.MenuID, &row.Name, &row.Price, &row.Available, &row.CreatedAt, &row.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan menu: %w", err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate menus: %w", err)
	}
	return result, nil
}

// InsertMenu は oc_menus に新しいメニューを INSERT する。
func (ds *MySQLMenuDataSource) InsertMenu(ctx context.Context, row *datasource.MenuRow) error {
	c := infraMysql.GetConn(ctx, ds.db)

	_, err := c.ExecContext(ctx,
		`INSERT INTO oc_menus (menu_id, name, price, available, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE name = VALUES(name), price = VALUES(price), available = VALUES(available), updated_at = VALUES(updated_at)`,
		row.MenuID,
		row.Name,
		row.Price,
		row.Available,
		row.CreatedAt.UTC(),
		row.UpdatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("insert menu: %w", err)
	}
	return nil
}

// UpdateMenuAvailability は指定した menu_id の availability を更新する。
func (ds *MySQLMenuDataSource) UpdateMenuAvailability(ctx context.Context, id string, available bool) error {
	c := infraMysql.GetConn(ctx, ds.db)

	_, err := c.ExecContext(ctx,
		`UPDATE oc_menus SET available = ? WHERE menu_id = ?`,
		available,
		id,
	)
	if err != nil {
		return fmt.Errorf("update menu availability: %w", err)
	}
	return nil
}
