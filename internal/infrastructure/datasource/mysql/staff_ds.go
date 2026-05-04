package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"mendo/internal/infrastructure/datasource"
)

// StaffDataSource は datasource.StaffDataSource の MySQL 実装。
//
// テーブル: staffs (id, name, phone, shift_type, created_at, updated_at)
//
// staff パッケージはアクティブレコードパターン。
// ID の採番は呼び出し元（Repository / UseCase）の責務。
type StaffDataSource struct {
	db *sql.DB
}

func NewStaffDataSource(db *sql.DB) *StaffDataSource {
	return &StaffDataSource{db: db}
}

// FindStaffByID は id を指定して StaffRow を返す。
// 見つからない場合は nil, nil を返す。
func (s *StaffDataSource) FindStaffByID(ctx context.Context, id string) (*datasource.StaffRow, error) {
	c := getConn(ctx, s.db)

	var row datasource.StaffRow
	err := c.QueryRowContext(ctx,
		`SELECT id, name, phone, shift_type, created_at, updated_at
		   FROM staffs
		  WHERE id = ?`,
		id,
	).Scan(
		&row.ID,
		&row.Name,
		&row.Phone,
		&row.ShiftType,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find staff by id: %w", err)
	}
	return &row, nil
}

// FindAllStaffs は全スタッフを返す。
func (s *StaffDataSource) FindAllStaffs(ctx context.Context) ([]datasource.StaffRow, error) {
	c := getConn(ctx, s.db)

	rows, err := c.QueryContext(ctx,
		`SELECT id, name, phone, shift_type, created_at, updated_at
		   FROM staffs
		  ORDER BY id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("find all staffs: %w", err)
	}
	defer rows.Close()

	var result []datasource.StaffRow
	for rows.Next() {
		var row datasource.StaffRow
		if err := rows.Scan(
			&row.ID,
			&row.Name,
			&row.Phone,
			&row.ShiftType,
			&row.CreatedAt,
			&row.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan staff row: %w", err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate staffs: %w", err)
	}

	return result, nil
}

// UpsertStaff はスタッフを INSERT OR UPDATE する。
// ID が空文字の場合は INSERT として扱い、呼び出し元が採番済み ID をセットして渡す。
func (s *StaffDataSource) UpsertStaff(ctx context.Context, row *datasource.StaffRow) error {
	c := getConn(ctx, s.db)

	_, err := c.ExecContext(ctx,
		`INSERT INTO staffs (id, name, phone, shift_type, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   name       = VALUES(name),
		   phone      = VALUES(phone),
		   shift_type = VALUES(shift_type),
		   updated_at = VALUES(updated_at)`,
		row.ID,
		row.Name,
		row.Phone,
		row.ShiftType,
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert staff: %w", err)
	}
	return nil
}

// DeleteStaff は指定した id のスタッフを削除する。
func (s *StaffDataSource) DeleteStaff(ctx context.Context, id string) error {
	c := getConn(ctx, s.db)

	_, err := c.ExecContext(ctx, `DELETE FROM staffs WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete staff: %w", err)
	}
	return nil
}
