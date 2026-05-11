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

func newOrderBoardDS() *dsmysql.OrderBoardDataSource {
	return dsmysql.NewOrderBoardDataSource(testDB)
}

func cleanupOrderBoard(t *testing.T, orderIDs ...string) {
	t.Helper()
	t.Cleanup(func() {
		for _, id := range orderIDs {
			_, _ = testDB.ExecContext(context.Background(), "DELETE FROM kc_order_board WHERE order_id = ?", id)
		}
	})
}

func Test_OrderBoardDS_UpsertAndFindByID(t *testing.T) {
	ctx := context.Background()
	ds := newOrderBoardDS()
	cleanupOrderBoard(t, "ob-rt-1")

	now := time.Now().UTC().Truncate(time.Second)
	row := &datasource.OrderBoardRow{
		OrderID:       "ob-rt-1",
		SeatNo:        3,
		OrderStatus:   "confirmed",
		CookingStatus: "",
		OrderedAt:     &now,
		CookingAt:     nil,
	}

	require.NoError(t, ds.UpsertOrderBoardRow(ctx, row))

	got, err := ds.FindOrderBoardRowByID(ctx, "ob-rt-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, row.OrderID, got.OrderID)
	assert.Equal(t, row.SeatNo, got.SeatNo)
	assert.Equal(t, row.OrderStatus, got.OrderStatus)
	assert.Equal(t, row.CookingStatus, got.CookingStatus)
	require.NotNil(t, got.OrderedAt)
	assert.Equal(t, *row.OrderedAt, *got.OrderedAt)
	assert.Nil(t, got.CookingAt)
}

func Test_OrderBoardDS_UpsertUpdates(t *testing.T) {
	ctx := context.Background()
	ds := newOrderBoardDS()
	cleanupOrderBoard(t, "ob-upd-1")

	now := time.Now().UTC().Truncate(time.Second)
	row := &datasource.OrderBoardRow{
		OrderID:       "ob-upd-1",
		SeatNo:        1,
		OrderStatus:   "pending",
		CookingStatus: "",
		OrderedAt:     &now,
		CookingAt:     nil,
	}
	require.NoError(t, ds.UpsertOrderBoardRow(ctx, row))

	// 同じ order_id で status を変更して再 Upsert
	cookingAt := now.Add(5 * time.Minute)
	updated := &datasource.OrderBoardRow{
		OrderID:       "ob-upd-1",
		SeatNo:        1,
		OrderStatus:   "confirmed",
		CookingStatus: "cooking",
		OrderedAt:     &now,
		CookingAt:     &cookingAt,
	}
	require.NoError(t, ds.UpsertOrderBoardRow(ctx, updated))

	got, err := ds.FindOrderBoardRowByID(ctx, "ob-upd-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "confirmed", got.OrderStatus)
	assert.Equal(t, "cooking", got.CookingStatus)
	require.NotNil(t, got.CookingAt)
	assert.Equal(t, cookingAt, *got.CookingAt)
}

func Test_OrderBoardDS_FindAllOrderBoardRows(t *testing.T) {
	ctx := context.Background()
	ds := newOrderBoardDS()
	cleanupOrderBoard(t, "ob-all-1", "ob-all-2")

	now := time.Now().UTC().Truncate(time.Second)
	rows := []*datasource.OrderBoardRow{
		{OrderID: "ob-all-1", SeatNo: 1, OrderStatus: "pending", CookingStatus: "", OrderedAt: &now},
		{OrderID: "ob-all-2", SeatNo: 2, OrderStatus: "confirmed", CookingStatus: "cooking", OrderedAt: &now},
	}
	for _, r := range rows {
		require.NoError(t, ds.UpsertOrderBoardRow(ctx, r))
	}

	all, err := ds.FindAllOrderBoardRows(ctx)
	require.NoError(t, err)

	found := make(map[string]bool)
	for _, r := range all {
		found[r.OrderID] = true
	}
	assert.True(t, found["ob-all-1"])
	assert.True(t, found["ob-all-2"])
}

func Test_OrderBoardDS_FindByID_NotFound(t *testing.T) {
	ctx := context.Background()
	ds := newOrderBoardDS()

	got, err := ds.FindOrderBoardRowByID(ctx, "ob-nonexistent-id")
	require.NoError(t, err)
	assert.Nil(t, got)
}
