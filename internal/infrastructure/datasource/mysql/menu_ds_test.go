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

func newMenuDS() *dsmysql.MySQLMenuDataSource {
	return dsmysql.NewMySQLMenuDataSource(testDB)
}

func cleanupMenus(t *testing.T, id string) {
	t.Helper()
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(), "DELETE FROM oc_menus WHERE menu_id = ?", id)
	})
}

func Test_MenuDS_InsertAndFind(t *testing.T) {
	ctx := context.Background()
	ds := newMenuDS()
	cleanupMenus(t, "menu-test-1")

	now := time.Now().UTC().Truncate(time.Second)
	row := &datasource.MenuRow{
		MenuID:    "menu-test-1",
		Name:      "醤油ラーメン",
		Price:     800,
		Available: true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	require.NoError(t, ds.InsertMenu(ctx, row))

	found, err := ds.FindMenuByID(ctx, "menu-test-1")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "醤油ラーメン", found.Name)
	assert.Equal(t, 800, found.Price)
	assert.True(t, found.Available)
}

func Test_MenuDS_UpdateAvailability(t *testing.T) {
	ctx := context.Background()
	ds := newMenuDS()
	cleanupMenus(t, "menu-test-2")

	now := time.Now().UTC().Truncate(time.Second)
	row := &datasource.MenuRow{
		MenuID: "menu-test-2", Name: "味噌ラーメン", Price: 900,
		Available: true, CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, ds.InsertMenu(ctx, row))

	require.NoError(t, ds.UpdateMenuAvailability(ctx, "menu-test-2", false))

	found, err := ds.FindMenuByID(ctx, "menu-test-2")
	require.NoError(t, err)
	assert.False(t, found.Available)
}

func Test_MenuDS_FindByID_存在しない(t *testing.T) {
	ctx := context.Background()
	ds := newMenuDS()

	found, err := ds.FindMenuByID(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, found)
}

func Test_MenuDS_FindAll(t *testing.T) {
	ctx := context.Background()
	ds := newMenuDS()
	cleanupMenus(t, "menu-all-1")
	cleanupMenus(t, "menu-all-2")

	now := time.Now().UTC().Truncate(time.Second)
	require.NoError(t, ds.InsertMenu(ctx, &datasource.MenuRow{
		MenuID: "menu-all-1", Name: "A_ラーメン", Price: 700, Available: true, CreatedAt: now, UpdatedAt: now,
	}))
	require.NoError(t, ds.InsertMenu(ctx, &datasource.MenuRow{
		MenuID: "menu-all-2", Name: "B_つけ麺", Price: 900, Available: true, CreatedAt: now, UpdatedAt: now,
	}))

	all, err := ds.FindAllMenus(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(all), 2)
}
