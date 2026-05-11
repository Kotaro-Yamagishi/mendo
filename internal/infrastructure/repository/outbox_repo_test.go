package repository_test

import (
	"context"
	"testing"

	"mendo/internal/domain"
	"mendo/internal/domain/order"
	"mendo/internal/infrastructure/repository"
	"mendo/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_OutboxRepo_Store(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubOutboxDataSource{}
	repo := repository.NewOutboxRepository(ds)

	events := []domain.Event{
		order.NewOrderCreated("order-1", 3, "corr-1"),
	}

	err := repo.Store(context.Background(), events)

	require.NoError(t, err)
	require.Len(t, ds.InsertedRows, 1)
	assert.Equal(t, "order.created", ds.InsertedRows[0].EventType)
	assert.Equal(t, "order-1", ds.InsertedRows[0].AggregateID)
	assert.NotEmpty(t, ds.InsertedRows[0].Payload)
	assert.False(t, ds.InsertedRows[0].Delivered)
}

func Test_OutboxRepo_MarkDelivered(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubOutboxDataSource{}
	repo := repository.NewOutboxRepository(ds)

	err := repo.MarkDelivered(context.Background(), []string{"evt-1", "evt-2"})

	require.NoError(t, err)
	assert.Equal(t, []string{"evt-1", "evt-2"}, ds.MarkedIDs)
}
