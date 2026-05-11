package repository_test

import (
	"context"
	"testing"

	"mendo/internal/infrastructure/repository"
	"mendo/internal/staff"
	"mendo/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Save + FindByID ラウンドトリップ
// =============================================================================

func Test_StaffRepo_Save_FindByID_ラウンドトリップ(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubStaffDataSource{}
	repo := repository.NewStaffRepository(ds)

	s := &staff.Staff{
		ID:        "staff-1",
		Name:      "山田太郎",
		Phone:     "090-1234-5678",
		ShiftType: "morning",
	}

	err := repo.Save(context.Background(), s)
	require.NoError(t, err)

	found, err := repo.FindByID(context.Background(), "staff-1")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "staff-1", found.ID)
	assert.Equal(t, "山田太郎", found.Name)
	assert.Equal(t, "090-1234-5678", found.Phone)
	assert.Equal(t, "morning", found.ShiftType)
}

func Test_StaffRepo_Save_バリデーションエラー(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubStaffDataSource{}
	repo := repository.NewStaffRepository(ds)

	s := &staff.Staff{
		ID:        "staff-1",
		Name:      "", // 空は invalid
		Phone:     "090-1234-5678",
		ShiftType: "morning",
	}

	err := repo.Save(context.Background(), s)

	require.Error(t, err)
	assert.Nil(t, ds.UpsertedRow) // UpsertStaff は呼ばれない
}

func Test_StaffRepo_FindByID_見つからない場合エラー(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubStaffDataSource{}
	repo := repository.NewStaffRepository(ds)

	_, err := repo.FindByID(context.Background(), "nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// =============================================================================
// FindAll
// =============================================================================

func Test_StaffRepo_FindAll_正常系(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubStaffDataSource{}
	repo := repository.NewStaffRepository(ds)

	staffs := []*staff.Staff{
		{ID: "staff-1", Name: "山田太郎", Phone: "090-1111-2222", ShiftType: "morning"},
		{ID: "staff-2", Name: "鈴木花子", Phone: "080-3333-4444", ShiftType: "night"},
	}
	for _, s := range staffs {
		require.NoError(t, repo.Save(context.Background(), s))
	}

	results, err := repo.FindAll(context.Background())

	require.NoError(t, err)
	assert.Len(t, results, 2)
}

// =============================================================================
// Delete
// =============================================================================

func Test_StaffRepo_Delete_正常系(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubStaffDataSource{}
	repo := repository.NewStaffRepository(ds)

	s := &staff.Staff{
		ID:        "staff-1",
		Name:      "山田太郎",
		Phone:     "090-1234-5678",
		ShiftType: "morning",
	}
	require.NoError(t, repo.Save(context.Background(), s))

	err := repo.Delete(context.Background(), "staff-1")

	require.NoError(t, err)
	assert.Equal(t, "staff-1", ds.DeletedID)

	// 削除後は FindByID でエラーになる
	_, err = repo.FindByID(context.Background(), "staff-1")
	require.Error(t, err)
}

func Test_StaffRepo_Delete_存在しないIDはエラー(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubStaffDataSource{}
	repo := repository.NewStaffRepository(ds)

	err := repo.Delete(context.Background(), "nonexistent")

	require.Error(t, err)
}
