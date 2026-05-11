package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	datasource "mendo/internal/infrastructure/datasource"
	infraMysql "mendo/internal/infrastructure/mysql"
)

// MySQLKitchenDataSource は KitchenDataSource の MySQL 実装。
//
// テーブル:
//
//	kc_kitchens      (kitchen_id, max_capacity, created_at)
//	kc_cooking_tasks (task_id, kitchen_id, order_id, instructions JSON, status, started_at, completed_at)
type MySQLKitchenDataSource struct {
	db *sql.DB
}

func NewMySQLKitchenDataSource(db *sql.DB) *MySQLKitchenDataSource {
	return &MySQLKitchenDataSource{db: db}
}

// FindKitchenByID は kc_kitchens から kitchen_id を指定して KitchenRow を返す。
func (ds *MySQLKitchenDataSource) FindKitchenByID(ctx context.Context, id string) (*datasource.KitchenRow, error) {
	c := infraMysql.GetConn(ctx, ds.db)

	var row datasource.KitchenRow
	err := c.QueryRowContext(ctx,
		`SELECT kitchen_id, max_capacity, created_at FROM kc_kitchens WHERE kitchen_id = ?`,
		id,
	).Scan(&row.KitchenID, &row.MaxCapacity, &row.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find kitchen by id: %w", err)
	}
	return &row, nil
}

// FindCookingTasksByKitchenID は kc_cooking_tasks から kitchen_id に紐づく全タスクを返す。
func (ds *MySQLKitchenDataSource) FindCookingTasksByKitchenID(ctx context.Context, kitchenID string) ([]datasource.CookingTaskRow, error) {
	c := infraMysql.GetConn(ctx, ds.db)

	rows, err := c.QueryContext(ctx,
		`SELECT task_id, kitchen_id, order_id, status, instructions, started_at, completed_at
		   FROM kc_cooking_tasks
		  WHERE kitchen_id = ?
		  ORDER BY started_at ASC`,
		kitchenID,
	)
	if err != nil {
		return nil, fmt.Errorf("find cooking tasks by kitchen id: %w", err)
	}
	defer rows.Close()

	var result []datasource.CookingTaskRow
	for rows.Next() {
		var (
			row             datasource.CookingTaskRow
			instructionsRaw []byte
			completedAt     sql.NullTime
		)
		if err := rows.Scan(
			&row.TaskID,
			&row.KitchenID,
			&row.OrderID,
			&row.Status,
			&instructionsRaw,
			&row.StartedAt,
			&completedAt,
		); err != nil {
			return nil, fmt.Errorf("scan cooking task: %w", err)
		}

		var dtos []datasource.CookingInstructionDTO
		if err := json.Unmarshal(instructionsRaw, &dtos); err != nil {
			return nil, fmt.Errorf("unmarshal instructions: %w", err)
		}
		instructionsJSON, err := json.Marshal(dtos)
		if err != nil {
			return nil, fmt.Errorf("marshal instructions to string: %w", err)
		}
		row.Instructions = string(instructionsJSON)

		if completedAt.Valid {
			row.CompletedAt = &completedAt.Time
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cooking tasks: %w", err)
	}
	return result, nil
}

// UpsertKitchen は kc_kitchens を INSERT OR UPDATE する。
func (ds *MySQLKitchenDataSource) UpsertKitchen(ctx context.Context, row *datasource.KitchenRow) error {
	c := infraMysql.GetConn(ctx, ds.db)

	_, err := c.ExecContext(ctx,
		`INSERT INTO kc_kitchens (kitchen_id, max_capacity, created_at) VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE max_capacity = VALUES(max_capacity)`,
		row.KitchenID,
		row.MaxCapacity,
		row.CreatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("upsert kitchen: %w", err)
	}
	return nil
}

// InsertCookingTask は kc_cooking_tasks に新しいタスクを INSERT する。
func (ds *MySQLKitchenDataSource) InsertCookingTask(ctx context.Context, row *datasource.CookingTaskRow) error {
	c := infraMysql.GetConn(ctx, ds.db)

	_, err := c.ExecContext(ctx,
		`INSERT IGNORE INTO kc_cooking_tasks (task_id, kitchen_id, order_id, status, instructions, started_at, completed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		row.TaskID,
		row.KitchenID,
		row.OrderID,
		row.Status,
		row.Instructions,
		row.StartedAt.UTC(),
		row.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("insert cooking task: %w", err)
	}
	return nil
}

// UpdateCookingTaskStatus は指定した kitchen_id + order_id のタスクのステータスを更新する。
func (ds *MySQLKitchenDataSource) UpdateCookingTaskStatus(ctx context.Context, kitchenID, orderID string, status string) error {
	c := infraMysql.GetConn(ctx, ds.db)

	_, err := c.ExecContext(ctx,
		`UPDATE kc_cooking_tasks SET status = ?
		  WHERE kitchen_id = ? AND order_id = ?`,
		status,
		kitchenID,
		orderID,
	)
	if err != nil {
		return fmt.Errorf("update cooking task status: %w", err)
	}
	return nil
}
