package repository_test

import (
	"context"
	"testing"
	"time"

	"mendo/internal/domain/kitchen"
	"mendo/internal/domain/order"
	"mendo/internal/infrastructure/datasource"
	"mendo/internal/infrastructure/repository"
	"mendo/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// ApplyOrderEvent - OrderCreated
// =============================================================================

func Test_OrderBoardRepo_ApplyOrderEvent_OrderCreated_Upsertされる(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubOrderBoardDataSource{}
	repo := repository.NewOrderBoardRepository(ds)

	evt := order.NewOrderCreated("order-1", 4, "corr-1")
	evt.OccurredAt = time.Now()

	err := repo.ApplyOrderEvent(context.Background(), evt)

	require.NoError(t, err)
	require.NotNil(t, ds.UpsertedRow)
	assert.Equal(t, "order-1", ds.UpsertedRow.OrderID)
	assert.Equal(t, 4, ds.UpsertedRow.SeatNo)
	assert.Equal(t, "pending", ds.UpsertedRow.OrderStatus)
	assert.Equal(t, "", ds.UpsertedRow.CookingStatus)
}

// =============================================================================
// ApplyKitchenEvent - CookingCompleted
// =============================================================================

func Test_OrderBoardRepo_ApplyKitchenEvent_CookingCompleted_CookingStatusが更新される(t *testing.T) {
	t.Parallel()

	orderedAt := time.Now().Add(-5 * time.Minute)
	ds := &testutil.StubOrderBoardDataSource{
		Rows: []datasource.OrderBoardRow{
			{
				OrderID:       "order-1",
				SeatNo:        4,
				OrderStatus:   "confirmed",
				CookingStatus: "waiting",
				OrderedAt:     &orderedAt,
			},
		},
	}
	repo := repository.NewOrderBoardRepository(ds)

	evt := kitchen.NewCookingCompleted(kitchen.KitchenID("kitchen-1"), order.OrderID("order-1"), "corr-1")
	evt.OccurredAt = time.Now()

	err := repo.ApplyKitchenEvent(context.Background(), evt)

	require.NoError(t, err)
	require.NotNil(t, ds.UpsertedRow)
	assert.Equal(t, "order-1", ds.UpsertedRow.OrderID)
	assert.Equal(t, "completed", ds.UpsertedRow.CookingStatus)
	assert.NotNil(t, ds.UpsertedRow.CookingAt)
}

// =============================================================================
// FindAll
// =============================================================================

func Test_OrderBoardRepo_FindAll_正常系(t *testing.T) {
	t.Parallel()

	orderedAt := time.Now()
	ds := &testutil.StubOrderBoardDataSource{
		Rows: []datasource.OrderBoardRow{
			{
				OrderID:       "order-1",
				SeatNo:        1,
				OrderStatus:   "pending",
				CookingStatus: "",
				OrderedAt:     &orderedAt,
			},
			{
				OrderID:       "order-2",
				SeatNo:        2,
				OrderStatus:   "confirmed",
				CookingStatus: "waiting",
				OrderedAt:     &orderedAt,
			},
		},
	}
	repo := repository.NewOrderBoardRepository(ds)

	entries, err := repo.FindAll(context.Background())

	require.NoError(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "order-1", entries[0].OrderID)
	assert.Equal(t, "pending", entries[0].OrderStatus)
	assert.Equal(t, "order-2", entries[1].OrderID)
	assert.Equal(t, "confirmed", entries[1].OrderStatus)
	assert.Equal(t, "waiting", entries[1].CookingStatus)
}
