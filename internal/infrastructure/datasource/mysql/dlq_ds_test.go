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

func newDLQDS() *dsmysql.DLQDataSource {
	return dsmysql.NewDLQDataSource(testDB)
}

func cleanupDLQ(t *testing.T, ids ...string) {
	t.Helper()
	t.Cleanup(func() {
		for _, id := range ids {
			_, _ = testDB.ExecContext(context.Background(), "DELETE FROM dlq WHERE id = ?", id)
		}
	})
}

func Test_DLQDS_InsertAndFindByID(t *testing.T) {
	ctx := context.Background()
	ds := newDLQDS()
	cleanupDLQ(t, "dlq-rt-1")

	now := time.Now().UTC().Truncate(time.Second)
	row := &datasource.DeadLetterRow{
		ID:          "dlq-rt-1",
		EventType:   "order.created",
		Payload:     `{"order_id":"o-1"}`,
		Error:       "handler panicked",
		FailCount:   3,
		HandlerName: "OrderCreatedHandler",
		LastFailAt:  now,
	}

	require.NoError(t, ds.InsertDeadLetterRow(ctx, row))

	got, err := ds.FindDeadLetterRowByID(ctx, "dlq-rt-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, row.ID, got.ID)
	assert.Equal(t, row.EventType, got.EventType)
	assert.Equal(t, row.Payload, got.Payload)
	assert.Equal(t, row.Error, got.Error)
	assert.Equal(t, row.FailCount, got.FailCount)
	assert.Equal(t, row.HandlerName, got.HandlerName)
	assert.Equal(t, row.LastFailAt, got.LastFailAt)
}

func Test_DLQDS_FindAllDeadLetterRows(t *testing.T) {
	ctx := context.Background()
	ds := newDLQDS()
	cleanupDLQ(t, "dlq-all-1", "dlq-all-2")

	now := time.Now().UTC().Truncate(time.Second)
	rows := []*datasource.DeadLetterRow{
		{ID: "dlq-all-1", EventType: "order.created", Payload: `{}`, Error: "err1", FailCount: 1, HandlerName: "H1", LastFailAt: now},
		{ID: "dlq-all-2", EventType: "order.confirmed", Payload: `{}`, Error: "err2", FailCount: 2, HandlerName: "H2", LastFailAt: now},
	}
	for _, r := range rows {
		require.NoError(t, ds.InsertDeadLetterRow(ctx, r))
	}

	all, err := ds.FindAllDeadLetterRows(ctx)
	require.NoError(t, err)

	found := make(map[string]bool)
	for _, r := range all {
		found[r.ID] = true
	}
	assert.True(t, found["dlq-all-1"])
	assert.True(t, found["dlq-all-2"])
}

func Test_DLQDS_DeleteDeadLetterRow(t *testing.T) {
	ctx := context.Background()
	ds := newDLQDS()
	cleanupDLQ(t, "dlq-del-1")

	now := time.Now().UTC().Truncate(time.Second)
	row := &datasource.DeadLetterRow{
		ID:          "dlq-del-1",
		EventType:   "order.created",
		Payload:     `{}`,
		Error:       "err",
		FailCount:   1,
		HandlerName: "H",
		LastFailAt:  now,
	}
	require.NoError(t, ds.InsertDeadLetterRow(ctx, row))

	require.NoError(t, ds.DeleteDeadLetterRow(ctx, "dlq-del-1"))

	got, err := ds.FindDeadLetterRowByID(ctx, "dlq-del-1")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func Test_DLQDS_FindByID_NotFound(t *testing.T) {
	ctx := context.Background()
	ds := newDLQDS()

	got, err := ds.FindDeadLetterRowByID(ctx, "dlq-nonexistent-id")
	require.NoError(t, err)
	assert.Nil(t, got)
}
