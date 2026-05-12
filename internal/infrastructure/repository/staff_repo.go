package repository

import (
	"context"

	"mendo/internal/apperrors"
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
		return err
	}
	row := &datasource.StaffRow{
		ID:        s.ID,
		Name:      s.Name,
		Phone:     s.Phone,
		ShiftType: s.ShiftType,
	}
	if err := r.ds.UpsertStaff(ctx, row); err != nil {
		return apperrors.Infrastructure("スタッフの保存に失敗", err)
	}
	return nil
}

// FindByID は StaffRow を取得して Staff を返す。
func (r *StaffRepository) FindByID(ctx context.Context, id string) (*staff.Staff, error) {
	row, err := r.ds.FindStaffByID(ctx, id)
	if err != nil {
		return nil, apperrors.Infrastructure("スタッフの取得に失敗", err)
	}
	if row == nil {
		return nil, apperrors.NotFound("staff", id)
	}
	return rowToStaff(row), nil
}

// FindAll は全 StaffRow を取得して Staff スライスを返す。
func (r *StaffRepository) FindAll(ctx context.Context) ([]*staff.Staff, error) {
	rows, err := r.ds.FindAllStaffs(ctx)
	if err != nil {
		return nil, apperrors.Infrastructure("スタッフ一覧の取得に失敗", err)
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
		return apperrors.Infrastructure("スタッフの削除に失敗", err)
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
