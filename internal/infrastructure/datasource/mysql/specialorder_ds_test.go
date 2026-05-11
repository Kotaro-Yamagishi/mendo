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

func newSpecialOrderDS() *dsmysql.MySQLSpecialOrderDataSource {
	return dsmysql.NewMySQLSpecialOrderDataSource(testDB)
}

func cleanupSpecialOrders(t *testing.T, id string) {
	t.Helper()
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(), "DELETE FROM sc_special_orders WHERE id = ?", id)
	})
}

func Test_SpecialOrderDS_UpsertAndFind(t *testing.T) {
	ctx := context.Background()
	ds := newSpecialOrderDS()
	cleanupSpecialOrders(t, "so-test-1")

	now := time.Now().UTC().Truncate(time.Second)
	row := &datasource.SpecialOrderRow{
		ID: "so-test-1", OrderID: "order-1", MenuName: "特製つけ麺",
		Status: 1, SuggestedMenu: "", CreatedAt: now, UpdatedAt: now,
	}

	require.NoError(t, ds.UpsertSpecialOrder(ctx, row))

	found, err := ds.FindSpecialOrderByID(ctx, "so-test-1")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "特製つけ麺", found.MenuName)
	assert.Equal(t, 1, found.Status)
}

func Test_SpecialOrderDS_Upsert_更新(t *testing.T) {
	ctx := context.Background()
	ds := newSpecialOrderDS()
	cleanupSpecialOrders(t, "so-test-2")

	now := time.Now().UTC().Truncate(time.Second)
	row := &datasource.SpecialOrderRow{
		ID: "so-test-2", OrderID: "order-1", MenuName: "特製つけ麺",
		Status: 1, CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, ds.UpsertSpecialOrder(ctx, row))

	// Status を更新（Rejected に変更）
	row.Status = 3
	row.SuggestedMenu = "醤油ラーメン"
	row.UpdatedAt = time.Now().UTC().Truncate(time.Second)
	require.NoError(t, ds.UpsertSpecialOrder(ctx, row))

	found, err := ds.FindSpecialOrderByID(ctx, "so-test-2")
	require.NoError(t, err)
	assert.Equal(t, 3, found.Status)
	assert.Equal(t, "醤油ラーメン", found.SuggestedMenu)
}

func Test_SpecialOrderDS_Find_存在しない(t *testing.T) {
	ctx := context.Background()
	ds := newSpecialOrderDS()

	found, err := ds.FindSpecialOrderByID(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, found)
}
