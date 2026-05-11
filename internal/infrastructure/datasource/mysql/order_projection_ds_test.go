//go:build integration

package mysql_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"mendo/internal/infrastructure/datasource"
	dsmysql "mendo/internal/infrastructure/datasource/mysql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newOrderProjectionDS() *dsmysql.OrderProjectionDataSource {
	return dsmysql.NewOrderProjectionDataSource(testDB)
}

func cleanupOrderProjection(t *testing.T, orderIDs ...string) {
	t.Helper()
	t.Cleanup(func() {
		for _, id := range orderIDs {
			_, _ = testDB.ExecContext(context.Background(), "DELETE FROM oc_order_projections WHERE order_id = ?", id)
		}
	})
}

func Test_OrderProjectionDS_UpsertAndFindByID(t *testing.T) {
	ctx := context.Background()
	ds := newOrderProjectionDS()
	cleanupOrderProjection(t, "op-rt-1")

	now := time.Now().UTC().Truncate(time.Second)
	row := &datasource.OrderProjectionRow{
		OrderID:   "op-rt-1",
		SeatNo:    5,
		Items:     `[{"menu_id":"m-1","menu_name":"カレー","price":800,"quantity":2}]`,
		Total:     1600,
		Status:    "pending",
		CreatedAt: now,
		UpdatedAt: now,
	}

	require.NoError(t, ds.UpsertOrderProjection(ctx, row))

	got, err := ds.FindOrderProjectionByID(ctx, "op-rt-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, row.OrderID, got.OrderID)
	assert.Equal(t, row.SeatNo, got.SeatNo)
	// MySQL は JSON オブジェクトのキーをアルファベット順に並べ替えるため、
	// 文字列比較ではなく unmarshal してから比較する。
	var wantItems, gotItems any
	require.NoError(t, json.Unmarshal([]byte(row.Items), &wantItems))
	require.NoError(t, json.Unmarshal([]byte(got.Items), &gotItems))
	assert.Equal(t, wantItems, gotItems)
	assert.Equal(t, row.Total, got.Total)
	assert.Equal(t, row.Status, got.Status)
}

func Test_OrderProjectionDS_FindAllOrderProjections(t *testing.T) {
	ctx := context.Background()
	ds := newOrderProjectionDS()
	cleanupOrderProjection(t, "op-all-1", "op-all-2")

	now := time.Now().UTC().Truncate(time.Second)
	rows := []*datasource.OrderProjectionRow{
		{OrderID: "op-all-1", SeatNo: 1, Items: `[]`, Total: 0, Status: "pending", CreatedAt: now, UpdatedAt: now},
		{OrderID: "op-all-2", SeatNo: 2, Items: `[]`, Total: 500, Status: "confirmed", CreatedAt: now, UpdatedAt: now},
	}
	for _, r := range rows {
		require.NoError(t, ds.UpsertOrderProjection(ctx, r))
	}

	all, err := ds.FindAllOrderProjections(ctx)
	require.NoError(t, err)

	found := make(map[string]bool)
	for _, r := range all {
		found[r.OrderID] = true
	}
	assert.True(t, found["op-all-1"])
	assert.True(t, found["op-all-2"])
}

func Test_OrderProjectionDS_FindByID_NotFound(t *testing.T) {
	ctx := context.Background()
	ds := newOrderProjectionDS()

	got, err := ds.FindOrderProjectionByID(ctx, "op-nonexistent-id")
	require.NoError(t, err)
	assert.Nil(t, got)
}
