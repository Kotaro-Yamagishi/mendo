package repository_test

import (
	"context"
	"testing"
	"time"

	"mendo/internal/domain/order"
	"mendo/internal/infrastructure/datasource"
	"mendo/internal/infrastructure/repository"
	"mendo/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// HandleEvent - OrderCreated
// =============================================================================

func Test_OrderProjectionRepo_HandleEvent_OrderCreated_Upsertされる(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubOrderProjectionDataSource{}
	repo := repository.NewOrderProjectionRepository(ds)

	evt := order.NewOrderCreated("order-1", 3, "corr-1")
	evt.OccurredAt = time.Now()

	err := repo.HandleEvent(context.Background(), evt)

	require.NoError(t, err)
	require.NotNil(t, ds.UpsertedRow)
	assert.Equal(t, "order-1", ds.UpsertedRow.OrderID)
	assert.Equal(t, 3, ds.UpsertedRow.SeatNo)
	assert.Equal(t, "pending", ds.UpsertedRow.Status)
}

// =============================================================================
// HandleEvent - OrderConfirmed
// =============================================================================

func Test_OrderProjectionRepo_HandleEvent_OrderConfirmed_StatusがUpdatedされる(t *testing.T) {
	t.Parallel()

	existing := &datasource.OrderProjectionRow{
		OrderID:   "order-1",
		SeatNo:    3,
		Items:     "[]",
		Total:     0,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	ds := &testutil.StubOrderProjectionDataSource{
		Row: existing,
	}
	repo := repository.NewOrderProjectionRepository(ds)

	evt := order.NewOrderConfirmed("order-1", nil, 3, "corr-1")
	evt.OccurredAt = time.Now()

	err := repo.HandleEvent(context.Background(), evt)

	require.NoError(t, err)
	require.NotNil(t, ds.UpsertedRow)
	assert.Equal(t, "confirmed", ds.UpsertedRow.Status)
	assert.Equal(t, "order-1", ds.UpsertedRow.OrderID)
}

// =============================================================================
// FindByID
// =============================================================================

func Test_OrderProjectionRepo_FindByID_正常系(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubOrderProjectionDataSource{
		Row: &datasource.OrderProjectionRow{
			OrderID:   "order-1",
			SeatNo:    5,
			Items:     `[{"menu_id":"menu-A","toppings":[],"hardness":""}]`,
			Total:     500,
			Status:    "confirmed",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	repo := repository.NewOrderProjectionRepository(ds)

	result, err := repo.FindByID(context.Background(), "order-1")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "order-1", result.OrderID)
	assert.Equal(t, 5, result.SeatNo)
	assert.Equal(t, "confirmed", result.Status)
	assert.Equal(t, 1, result.ItemCount)
}

func Test_OrderProjectionRepo_FindByID_見つからない場合エラー(t *testing.T) {
	t.Parallel()

	// Row も Rows も nil → FindOrderProjectionByID は "order not found" エラーを返す
	ds := &testutil.StubOrderProjectionDataSource{}
	repo := repository.NewOrderProjectionRepository(ds)

	_, err := repo.FindByID(context.Background(), "nonexistent")

	require.Error(t, err)
}

// =============================================================================
// FindAll
// =============================================================================

func Test_OrderProjectionRepo_FindAll_正常系(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubOrderProjectionDataSource{
		Rows: []datasource.OrderProjectionRow{
			{OrderID: "order-1", SeatNo: 1, Items: "[]", Status: "pending"},
			{OrderID: "order-2", SeatNo: 2, Items: `[{"menu_id":"m1","toppings":[],"hardness":""}]`, Status: "confirmed"},
		},
	}
	repo := repository.NewOrderProjectionRepository(ds)

	results, err := repo.FindAll(context.Background())

	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "order-1", results[0].OrderID)
	assert.Equal(t, 0, results[0].ItemCount)
	assert.Equal(t, "order-2", results[1].OrderID)
	assert.Equal(t, 1, results[1].ItemCount)
}
