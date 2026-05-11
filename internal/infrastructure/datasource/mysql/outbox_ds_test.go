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

func newOutboxDS() *dsmysql.MySQLOutboxDataSource {
	return dsmysql.NewMySQLOutboxDataSource(testDB)
}

func cleanupOutbox(t *testing.T, ids ...string) {
	t.Helper()
	t.Cleanup(func() {
		for _, id := range ids {
			_, _ = testDB.ExecContext(context.Background(), "DELETE FROM outbox WHERE id = ?", id)
		}
	})
}

func Test_OutboxDS_FullLifecycle(t *testing.T) {
	ctx := context.Background()
	ds := newOutboxDS()
	cleanupOutbox(t, "outbox-1", "outbox-2")

	now := time.Now().UTC().Truncate(time.Second)
	rows := []datasource.OutboxRow{
		{ID: "outbox-1", EventType: "order.created", AggregateID: "order-1", Payload: `{"test":1}`, CreatedAt: now},
		{ID: "outbox-2", EventType: "order.confirmed", AggregateID: "order-1", Payload: `{"test":2}`, CreatedAt: now},
	}

	// Insert
	require.NoError(t, ds.InsertOutboxRows(ctx, rows))

	// Fetch undelivered
	undelivered, err := ds.FindUndeliveredOutboxRows(ctx, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(undelivered), 2)

	// Mark delivered
	require.NoError(t, ds.MarkOutboxRowsDelivered(ctx, []string{"outbox-1"}))

	// Fetch again — outbox-1 は配信済みなので含まれない
	undelivered2, err := ds.FindUndeliveredOutboxRows(ctx, 10)
	require.NoError(t, err)

	deliveredIDs := make(map[string]bool)
	for _, r := range undelivered2 {
		deliveredIDs[r.ID] = true
	}
	assert.False(t, deliveredIDs["outbox-1"], "outbox-1 は配信済み")
}
