package repository_test

import (
	"context"
	"encoding/json"
	"errors"
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
// FindByID
// =============================================================================

func Test_KitchenRepo_FindByID_正常系(t *testing.T) {
	t.Parallel()

	instrJSON, _ := json.Marshal([]datasource.CookingInstructionDTO{
		{MenuName: "醤油ラーメン", Toppings: []string{"ネギ"}, Hardness: "普通"},
	})
	ds := &testutil.StubKitchenDataSource{
		KitchenRow: &datasource.KitchenRow{KitchenID: "kitchen-1", MaxCapacity: 10},
		TaskRows: []datasource.CookingTaskRow{
			{
				TaskID:       "task-1",
				KitchenID:    "kitchen-1",
				OrderID:      "order-1",
				Status:       "cooking",
				Instructions: string(instrJSON),
				StartedAt:    time.Now(),
			},
		},
	}
	repo := repository.NewKitchenRepository(ds)

	k, err := repo.FindByID(context.Background(), kitchen.KitchenID("kitchen-1"))

	require.NoError(t, err)
	require.NotNil(t, k)
	assert.Equal(t, kitchen.KitchenID("kitchen-1"), k.ID())
	// タスク1つ → キャパは9
	assert.Equal(t, 9, k.CookingCapacity())
}

func Test_KitchenRepo_FindByID_異常系(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		ds              *testutil.StubKitchenDataSource
		wantErrContains string
	}{
		{
			name:            "厨房が見つからない",
			ds:              &testutil.StubKitchenDataSource{KitchenRow: nil},
			wantErrContains: "not found",
		},
		{
			name:            "DataSourceエラー",
			ds:              &testutil.StubKitchenDataSource{FindErr: errors.New("db error")},
			wantErrContains: "FindByID",
		},
		{
			name: "タスク取得エラー",
			ds: &testutil.StubKitchenDataSource{
				KitchenRow:   &datasource.KitchenRow{KitchenID: "kitchen-1"},
				FindTasksErr: errors.New("task db error"),
			},
			wantErrContains: "tasks",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := repository.NewKitchenRepository(tc.ds)

			_, err := repo.FindByID(context.Background(), kitchen.KitchenID("kitchen-1"))

			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErrContains)
		})
	}
}

// =============================================================================
// Save
// =============================================================================

func Test_KitchenRepo_Save_正常系(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubKitchenDataSource{}
	repo := repository.NewKitchenRepository(ds)

	k := kitchen.NewKitchen("kitchen-1")

	err := repo.Save(context.Background(), k)

	require.NoError(t, err)
	require.NotNil(t, ds.UpsertedKitchen)
	assert.Equal(t, "kitchen-1", ds.UpsertedKitchen.KitchenID)
}

func Test_KitchenRepo_Save_CookingCompleted_タスクステータス更新(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubKitchenDataSource{}
	repo := repository.NewKitchenRepository(ds)

	// タスクを持つ Kitchen を作って完了させる
	k := kitchen.NewKitchen("kitchen-1")
	require.NoError(t, k.AddCookingTask(order.OrderID("order-1"), []kitchen.CookingInstruction{{MenuName: "ラーメン"}}))
	require.NoError(t, k.CompleteCookingTask(order.OrderID("order-1")))

	err := repo.Save(context.Background(), k)

	require.NoError(t, err)
	// CookingCompleted イベントでタスクステータスが更新される
	assert.Equal(t, "kitchen-1", ds.UpdatedKitchenID)
	assert.Equal(t, "order-1", ds.UpdatedOrderID)
	assert.Equal(t, "completed", ds.UpdatedStatus)
}

func Test_KitchenRepo_Save_UpsertエラーStopsPropagation(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubKitchenDataSource{UpsertErr: errors.New("upsert failed")}
	repo := repository.NewKitchenRepository(ds)

	k := kitchen.NewKitchen("kitchen-1")

	err := repo.Save(context.Background(), k)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "UpsertKitchen")
}
