package repository

import (
	"context"
	"fmt"

	"mendo/internal/infrastructure/datasource"
	"mendo/internal/staff"
)

// StaffRepository は datasource を使った Staff の永続化実装。
// staff.Store の datasource 版。アクティブレコードパターンに対応する。
type StaffRepository struct {
	ds datasource.StaffDataSource
}

func NewStaffRepository(ds datasource.StaffDataSource) *StaffRepository {
	return &StaffRepository{ds: ds}
}

// Save は Staff を StaffRow に変換して永続化する。
func (r *StaffRepository) Save(ctx context.Context, s *staff.Staff) error {
	if err := s.Validate(); err != nil {
		return fmt.Errorf("StaffRepository.Save validation: %w", err)
	}
	row := &datasource.StaffRow{
		ID:        s.ID,
		Name:      s.Name,
		Phone:     s.Phone,
		ShiftType: s.ShiftType,
	}
	if err := r.ds.UpsertStaff(ctx, row); err != nil {
		return fmt.Errorf("StaffRepository.Save UpsertStaff: %w", err)
	}
	return nil
}

// FindByID は StaffRow を取得して Staff を返す。
func (r *StaffRepository) FindByID(ctx context.Context, id string) (*staff.Staff, error) {
	row, err := r.ds.FindStaffByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("StaffRepository.FindByID: %w", err)
	}
	if row == nil {
		return nil, fmt.Errorf("staff not found: %s", id)
	}
	return rowToStaff(row), nil
}

// FindAll は全 StaffRow を取得して Staff スライスを返す。
func (r *StaffRepository) FindAll(ctx context.Context) ([]*staff.Staff, error) {
	rows, err := r.ds.FindAllStaffs(ctx)
	if err != nil {
		return nil, fmt.Errorf("StaffRepository.FindAll: %w", err)
	}
	staffs := make([]*staff.Staff, 0, len(rows))
	for i := range rows {
		staffs = append(staffs, rowToStaff(&rows[i]))
	}
	return staffs, nil
}

// Delete は指定した ID の Staff を削除する。
func (r *StaffRepository) Delete(ctx context.Context, id string) error {
	if err := r.ds.DeleteStaff(ctx, id); err != nil {
		return fmt.Errorf("StaffRepository.Delete: %w", err)
	}
	return nil
}

func rowToStaff(row *datasource.StaffRow) *staff.Staff {
	return &staff.Staff{
		ID:        row.ID,
		Name:      row.Name,
		Phone:     row.Phone,
		ShiftType: row.ShiftType,
	}
}
