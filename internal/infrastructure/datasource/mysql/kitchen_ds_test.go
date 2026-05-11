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

func newKitchenDS() *dsmysql.MySQLKitchenDataSource {
	return dsmysql.NewMySQLKitchenDataSource(testDB)
}

func cleanupKitchen(t *testing.T, id string) {
	t.Helper()
	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(), "DELETE FROM kc_cooking_tasks WHERE kitchen_id = ?", id)
		_, _ = testDB.ExecContext(context.Background(), "DELETE FROM kc_kitchens WHERE kitchen_id = ?", id)
	})
}

func Test_KitchenDS_UpsertAndFind(t *testing.T) {
	ctx := context.Background()
	ds := newKitchenDS()
	cleanupKitchen(t, "kitchen-test-1")

	row := &datasource.KitchenRow{
		KitchenID:   "kitchen-test-1",
		MaxCapacity: 10,
		CreatedAt:   time.Now().UTC(),
	}
	require.NoError(t, ds.UpsertKitchen(ctx, row))

	found, err := ds.FindKitchenByID(ctx, "kitchen-test-1")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "kitchen-test-1", found.KitchenID)
	assert.Equal(t, 10, found.MaxCapacity)
}

func Test_KitchenDS_CookingTask_InsertAndFind(t *testing.T) {
	ctx := context.Background()
	ds := newKitchenDS()
	cleanupKitchen(t, "kitchen-test-2")

	// Kitchen 作成
	require.NoError(t, ds.UpsertKitchen(ctx, &datasource.KitchenRow{
		KitchenID: "kitchen-test-2", MaxCapacity: 10, CreatedAt: time.Now().UTC(),
	}))

	// CookingTask 作成
	taskRow := &datasource.CookingTaskRow{
		TaskID:       "task-1",
		KitchenID:    "kitchen-test-2",
		OrderID:      "order-1",
		Status:       "pending",
		Instructions: `[{"menu_name":"醤油ラーメン","toppings":["ネギ"],"hardness":"硬め"}]`,
		StartedAt:    time.Now().UTC(),
	}
	require.NoError(t, ds.InsertCookingTask(ctx, taskRow))

	// Find
	tasks, err := ds.FindCookingTasksByKitchenID(ctx, "kitchen-test-2")
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "task-1", tasks[0].TaskID)
	assert.Equal(t, "order-1", tasks[0].OrderID)
	assert.Equal(t, "pending", tasks[0].Status)
}

func Test_KitchenDS_UpdateTaskStatus(t *testing.T) {
	ctx := context.Background()
	ds := newKitchenDS()
	cleanupKitchen(t, "kitchen-test-3")

	require.NoError(t, ds.UpsertKitchen(ctx, &datasource.KitchenRow{
		KitchenID: "kitchen-test-3", MaxCapacity: 10, CreatedAt: time.Now().UTC(),
	}))
	require.NoError(t, ds.InsertCookingTask(ctx, &datasource.CookingTaskRow{
		TaskID: "task-2", KitchenID: "kitchen-test-3", OrderID: "order-2",
		Status: "pending", Instructions: `[]`, StartedAt: time.Now().UTC(),
	}))

	// Update status
	require.NoError(t, ds.UpdateCookingTaskStatus(ctx, "kitchen-test-3", "order-2", "completed"))

	tasks, err := ds.FindCookingTasksByKitchenID(ctx, "kitchen-test-3")
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "completed", tasks[0].Status)
}
