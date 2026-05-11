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

func newEventStoreDS() *dsmysql.MySQLEventStoreDataSource {
	return dsmysql.NewMySQLEventStoreDataSource(testDB)
}

func cleanupEvents(t *testing.T, aggregateID string) {
	t.Helper()
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(),
			"DELETE FROM events WHERE aggregate_id = ?", aggregateID)
	})
}

// InsertEvents → FindEventsByAggregateID のラウンドトリップ。
// SQL の typo、カラムマッピング漏れをここで検出する。
func Test_EventStoreDS_InsertAndFind(t *testing.T) {
	ctx := context.Background()
	ds := newEventStoreDS()
	aggregateID := "test-ds-insert-find"
	cleanupEvents(t, aggregateID)

	rows := []datasource.EventRow{
		{
			EventID:       "evt-1",
			AggregateID:   aggregateID,
			AggregateType: "Order",
			EventType:     "order.created",
			Version:       0,
			Payload:       `{"seat_no":3}`,
			CorrelationID: "corr-1",
			CausationID:   "",
			CreatedAt:     time.Now().UTC(),
		},
		{
			EventID:       "evt-2",
			AggregateID:   aggregateID,
			AggregateType: "Order",
			EventType:     "order.item_added",
			Version:       1,
			Payload:       `{"menu_id":"menu-1"}`,
			CorrelationID: "corr-1",
			CausationID:   "",
			CreatedAt:     time.Now().UTC(),
		},
	}

	// Insert
	err := ds.InsertEvents(ctx, rows)
	require.NoError(t, err)

	// Find
	found, err := ds.FindEventsByAggregateID(ctx, aggregateID)
	require.NoError(t, err)
	require.Len(t, found, 2)

	// 順序検証: version ASC で返る
	assert.Equal(t, "evt-1", found[0].EventID)
	assert.Equal(t, "order.created", found[0].EventType)
	assert.Equal(t, `{"seat_no":3}`, found[0].Payload)
	assert.Equal(t, 0, found[0].Version)

	assert.Equal(t, "evt-2", found[1].EventID)
	assert.Equal(t, "order.item_added", found[1].EventType)
	assert.Equal(t, 1, found[1].Version)
}

// version 順序の検証。後から追加したイベントが正しい位置に来るか。
func Test_EventStoreDS_VersionOrdering(t *testing.T) {
	ctx := context.Background()
	ds := newEventStoreDS()
	aggregateID := "test-ds-version-order"
	cleanupEvents(t, aggregateID)

	// 1回目: version 0, 1
	err := ds.InsertEvents(ctx, []datasource.EventRow{
		{EventID: "evt-a0", AggregateID: aggregateID, EventType: "order.created", Version: 0, Payload: "{}", CreatedAt: time.Now().UTC()},
		{EventID: "evt-a1", AggregateID: aggregateID, EventType: "order.item_added", Version: 1, Payload: "{}", CreatedAt: time.Now().UTC()},
	})
	require.NoError(t, err)

	// 2回目: version 2（追記）
	err = ds.InsertEvents(ctx, []datasource.EventRow{
		{EventID: "evt-a2", AggregateID: aggregateID, EventType: "order.confirmed", Version: 2, Payload: "{}", CreatedAt: time.Now().UTC()},
	})
	require.NoError(t, err)

	// 全件取得 → version 順
	found, err := ds.FindEventsByAggregateID(ctx, aggregateID)
	require.NoError(t, err)
	require.Len(t, found, 3)
	assert.Equal(t, 0, found[0].Version)
	assert.Equal(t, 1, found[1].Version)
	assert.Equal(t, 2, found[2].Version)
}

// 存在しない aggregate_id → 空スライス
func Test_EventStoreDS_FindNonExistent(t *testing.T) {
	ctx := context.Background()
	ds := newEventStoreDS()

	found, err := ds.FindEventsByAggregateID(ctx, "nonexistent")

	require.NoError(t, err)
	assert.Empty(t, found)
}

// 重複 EventID → エラー（PRIMARY KEY 制約）
func Test_EventStoreDS_DuplicateEventID(t *testing.T) {
	ctx := context.Background()
	ds := newEventStoreDS()
	aggregateID := "test-ds-duplicate"
	cleanupEvents(t, aggregateID)

	row := datasource.EventRow{
		EventID: "evt-dup", AggregateID: aggregateID, EventType: "order.created",
		Version: 0, Payload: "{}", CreatedAt: time.Now().UTC(),
	}
	require.NoError(t, ds.InsertEvents(ctx, []datasource.EventRow{row}))

	// 同じ EventID で再 Insert → エラー
	err := ds.InsertEvents(ctx, []datasource.EventRow{row})
	require.Error(t, err, "重複 EventID は PRIMARY KEY 制約でエラーになる")
}
