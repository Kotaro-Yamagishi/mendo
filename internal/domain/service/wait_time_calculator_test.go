package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"mendo/internal/domain/kitchen"
	"mendo/internal/domain/service"
	"mendo/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_EstimateWaitTime_正常系(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		pendingOrders int
		capacity      int
		want          time.Duration
	}{
		{
			// pending 10件、capacity 5 → 10/5*5 = 10分
			name:          "通常計算",
			pendingOrders: 10,
			capacity:      5,
			want:          10 * time.Minute,
		},
		{
			// pending 0件 → 0分
			name:          "注文なし",
			pendingOrders: 0,
			capacity:      5,
			want:          0,
		},
		{
			// pending 3件、capacity 5 → 3/5=0 (整数除算) → 0分
			name:          "capacity以下_整数除算で0",
			pendingOrders: 3,
			capacity:      5,
			want:          0,
		},
		{
			// pending 20件、capacity 2 → 20/2*5 = 50分
			name:          "大量注文",
			pendingOrders: 20,
			capacity:      2,
			want:          50 * time.Minute,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			orderReader := &testutil.StubOrderReader{PendingCount: tt.pendingOrders}
			// capacity を作るには MaxConcurrentTasks - activeTasks = tt.capacity
			// activeTasks = MaxConcurrentTasks - tt.capacity = 10 - tt.capacity
			activeTasks := kitchen.MaxConcurrentTasks - tt.capacity
			kitchenReader := &testutil.StubKitchenReader{
				Kitchen: newKitchenWithCapacity(t, activeTasks),
			}
			calc := service.NewWaitTimeCalculator(orderReader, kitchenReader)

			got, err := calc.EstimateWaitTime(context.Background(), "kitchen-1")

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_EstimateWaitTime_フル稼働(t *testing.T) {
	t.Parallel()

	// capacity 0（フル稼働）→ 一律30分
	orderReader := &testutil.StubOrderReader{PendingCount: 5}
	kitchenReader := &testutil.StubKitchenReader{
		Kitchen: newKitchenWithCapacity(t, kitchen.MaxConcurrentTasks),
	}
	calc := service.NewWaitTimeCalculator(orderReader, kitchenReader)

	got, err := calc.EstimateWaitTime(context.Background(), "kitchen-1")

	require.NoError(t, err)
	assert.Equal(t, 30*time.Minute, got, "フル稼働時は一律30分")
}

func Test_EstimateWaitTime_OrderReader_エラー(t *testing.T) {
	t.Parallel()

	orderReader := &testutil.StubOrderReader{CountErr: errors.New("db error")}
	kitchenReader := &testutil.StubKitchenReader{
		Kitchen: kitchen.NewKitchen("kitchen-1"),
	}
	calc := service.NewWaitTimeCalculator(orderReader, kitchenReader)

	_, err := calc.EstimateWaitTime(context.Background(), "kitchen-1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "count pending")
}

func Test_EstimateWaitTime_KitchenReader_エラー(t *testing.T) {
	t.Parallel()

	orderReader := &testutil.StubOrderReader{PendingCount: 5}
	kitchenReader := &testutil.StubKitchenReader{FindErr: errors.New("not found")}
	calc := service.NewWaitTimeCalculator(orderReader, kitchenReader)

	_, err := calc.EstimateWaitTime(context.Background(), "kitchen-1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "find kitchen")
}

// newKitchenWithCapacity は指定した activeTasks 数の Kitchen を作る。
func newKitchenWithCapacity(t *testing.T, activeTasks int) *kitchen.Kitchen {
	t.Helper()
	var tasks []kitchen.CookingTask
	for i := 0; i < activeTasks; i++ {
		tasks = append(tasks, kitchen.ReconstructCookingTask(
			kitchen.TaskID("task-"+string(rune('a'+i))),
			"order-dummy",
			nil,
			kitchen.TaskPending,
			time.Now(),
		))
	}
	return kitchen.ReconstructKitchen("kitchen-1", tasks)
}
