//go:build integration

package mysql_test

import (
	"context"
	"testing"
	"time"

	"mendo/internal/infrastructure/datasource"
	dsmysql "mendo/internal/infrastructure/datasource/mysql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newStaffDS() *dsmysql.StaffDataSource {
	return dsmysql.NewStaffDataSource(testDB)
}

func cleanupStaff(t *testing.T, ids ...string) {
	t.Helper()
	t.Cleanup(func() {
		for _, id := range ids {
			_, _ = testDB.ExecContext(context.Background(), "DELETE FROM staffs WHERE id = ?", id)
		}
	})
}

func Test_StaffDS_UpsertAndFindByID(t *testing.T) {
	ctx := context.Background()
	ds := newStaffDS()
	cleanupStaff(t, "staff-rt-1")

	now := time.Now().UTC().Truncate(time.Second)
	row := &datasource.StaffRow{
		ID:        "staff-rt-1",
		Name:      "山田太郎",
		Phone:     "090-1234-5678",
		ShiftType: "morning",
		CreatedAt: now,
		UpdatedAt: now,
	}

	require.NoError(t, ds.UpsertStaff(ctx, row))

	got, err := ds.FindStaffByID(ctx, "staff-rt-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, row.ID, got.ID)
	assert.Equal(t, row.Name, got.Name)
	assert.Equal(t, row.Phone, got.Phone)
	assert.Equal(t, row.ShiftType, got.ShiftType)
}

func Test_StaffDS_FindAllStaffs(t *testing.T) {
	ctx := context.Background()
	ds := newStaffDS()
	cleanupStaff(t, "staff-all-1", "staff-all-2")

	now := time.Now().UTC().Truncate(time.Second)
	rows := []*datasource.StaffRow{
		{ID: "staff-all-1", Name: "スタッフA", Phone: "090-0001-0001", ShiftType: "morning", CreatedAt: now, UpdatedAt: now},
		{ID: "staff-all-2", Name: "スタッフB", Phone: "090-0002-0002", ShiftType: "evening", CreatedAt: now, UpdatedAt: now},
	}
	for _, r := range rows {
		require.NoError(t, ds.UpsertStaff(ctx, r))
	}

	all, err := ds.FindAllStaffs(ctx)
	require.NoError(t, err)

	found := make(map[string]bool)
	for _, r := range all {
		found[r.ID] = true
	}
	assert.True(t, found["staff-all-1"])
	assert.True(t, found["staff-all-2"])
}

func Test_StaffDS_DeleteStaff(t *testing.T) {
	ctx := context.Background()
	ds := newStaffDS()
	cleanupStaff(t, "staff-del-1")

	now := time.Now().UTC().Truncate(time.Second)
	row := &datasource.StaffRow{
		ID:        "staff-del-1",
		Name:      "削除対象",
		Phone:     "",
		ShiftType: "morning",
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, ds.UpsertStaff(ctx, row))

	require.NoError(t, ds.DeleteStaff(ctx, "staff-del-1"))

	got, err := ds.FindStaffByID(ctx, "staff-del-1")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func Test_StaffDS_FindByID_NotFound(t *testing.T) {
	ctx := context.Background()
	ds := newStaffDS()

	got, err := ds.FindStaffByID(ctx, "staff-nonexistent-id")
	require.NoError(t, err)
	assert.Nil(t, got)
}
