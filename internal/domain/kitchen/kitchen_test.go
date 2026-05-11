package kitchen_test

import (
	"testing"
	"time"

	"mendo/internal/domain/kitchen"
	"mendo/internal/domain/order"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// AddCookingTask — 調理タスクの追加
// =============================================================================

// ルール: 同時調理タスク数は MaxConcurrentTasks（10）まで
func Test_AddCookingTask_キャパシティ(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		activeTasks    int
		completedTasks int
		wantErr        bool
		wantRejected   bool
	}{
		{
			// 空の厨房 → 追加可能。最も基本的なケース。
			name:        "タスク0個_追加成功",
			activeTasks: 0,
			wantErr:     false,
		},
		{
			// 境界値: 上限-1。あと1つ追加できる。
			name:        "アクティブ9個_追加成功",
			activeTasks: 9,
			wantErr:     false,
		},
		{
			// 境界値: 上限ちょうど。これ以上追加できない。
			// CookingRejected イベントが発行される。
			// なぜイベントを発行するか: 拒否されたことを他のサービス（注文BC等）に
			// 通知して、リトライや代替フローを起動するため。
			name:         "アクティブ10個_追加拒否",
			activeTasks:  10,
			wantErr:      true,
			wantRejected: true,
		},
		{
			// 完了タスクはカウントしない。
			// なぜこのケースが重要: activeTasks のカウントロジックが
			// 「全タスク数」ではなく「未完了タスク数」で正しく計算されているか検証する。
			name:           "合計10個だが5個完了_追加成功",
			activeTasks:    5,
			completedTasks: 5,
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			k := newKitchenWithTasks(t, tt.activeTasks, tt.completedTasks)

			err := k.AddCookingTask(
				order.OrderID("new-order"),
				[]kitchen.CookingInstruction{{MenuName: "ラーメン", Hardness: "普通"}},
			)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "フル稼働")

				if tt.wantRejected {
					events := k.DomainEvents()
					require.NotEmpty(t, events, "拒否時は CookingRejected イベントが発行される")
					_, ok := events[len(events)-1].(kitchen.CookingRejected)
					assert.True(t, ok, "最後のイベントは CookingRejected であるべき")
				}
				return
			}
			require.NoError(t, err)
		})
	}
}

// =============================================================================
// CompleteCookingTask — 調理完了
// =============================================================================

func Test_CompleteCookingTask_正常系(t *testing.T) {
	t.Parallel()

	// Given: タスクが1つある厨房
	k := kitchen.NewKitchen("kitchen-1")
	err := k.AddCookingTask(
		order.OrderID("order-1"),
		[]kitchen.CookingInstruction{{MenuName: "味噌ラーメン"}},
	)
	require.NoError(t, err)

	// When: そのタスクを完了にする
	err = k.CompleteCookingTask(order.OrderID("order-1"))

	// Then: 成功し、CookingCompleted イベントが発行される
	require.NoError(t, err)
	events := k.DomainEvents()
	require.NotEmpty(t, events)
	completed, ok := events[len(events)-1].(kitchen.CookingCompleted)
	require.True(t, ok)
	assert.Equal(t, order.OrderID("order-1"), completed.OrderID)
	// event getter のカバレッジ
	assert.Equal(t, kitchen.EventTypeCookingCompleted, completed.GetEventType())
	assert.Equal(t, "kitchen-1", completed.GetAggregateID())
}

func Test_CompleteCookingTask_異常系(t *testing.T) {
	t.Parallel()

	t.Run("存在しないタスク", func(t *testing.T) {
		t.Parallel()
		k := kitchen.NewKitchen("kitchen-1")

		err := k.CompleteCookingTask(order.OrderID("nonexistent"))

		require.ErrorContains(t, err, "見つかりません")
	})

	t.Run("二重完了", func(t *testing.T) {
		t.Parallel()
		// Given: 既に完了したタスクがある
		k := kitchen.NewKitchen("kitchen-1")
		require.NoError(t, k.AddCookingTask(
			order.OrderID("order-1"),
			[]kitchen.CookingInstruction{{MenuName: "ラーメン"}},
		))
		require.NoError(t, k.CompleteCookingTask(order.OrderID("order-1")))

		// When: もう一度完了しようとする
		err := k.CompleteCookingTask(order.OrderID("order-1"))

		// Then: 二重完了はエラー。
		// なぜ防止するか: CookingCompleted イベントが重複発行されると、
		// 下流のサービスが「2回完了した」と誤解する。
		require.ErrorContains(t, err, "すでに調理完了")
	})
}

// =============================================================================
// CookingCapacity — 調理可能数の確認
// =============================================================================

func Test_CookingCapacity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		activeTasks    int
		completedTasks int
		want           int
	}{
		{"空の厨房", 0, 0, 10},
		{"アクティブ3個", 3, 0, 7},
		{"アクティブ10個", 10, 0, 0},
		// 完了タスクはキャパシティに影響しない。
		{"アクティブ3個_完了5個", 3, 5, 7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			k := newKitchenWithTasks(t, tt.activeTasks, tt.completedTasks)

			got := k.CookingCapacity()

			assert.Equal(t, tt.want, got)
		})
	}
}

// =============================================================================
// ReconstructKitchen — DB からの復元
// =============================================================================

func Test_ReconstructKitchen(t *testing.T) {
	t.Parallel()

	// Given: DB から読み込んだタスク一覧
	tasks := []kitchen.CookingTask{
		kitchen.ReconstructCookingTask(
			kitchen.TaskID("task-1"),
			order.OrderID("order-1"),
			[]kitchen.CookingInstruction{{MenuName: "醤油ラーメン"}},
			kitchen.TaskPending,
			time.Now(),
		),
	}

	// When: 復元
	k := kitchen.ReconstructKitchen("kitchen-1", tasks)

	// Then: 復元された Kitchen は正しい状態
	assert.Equal(t, kitchen.KitchenID("kitchen-1"), k.ID())
	// 復元時はイベントを発行しない。
	// なぜ: 復元は「既に永続化された状態の読み戻し」であり、新たな事実の発生ではないから。
	assert.Empty(t, k.DomainEvents())
	assert.Equal(t, 9, k.CookingCapacity(), "タスク1つなのでキャパは9")
}

// =============================================================================
// テストヘルパー
// =============================================================================

// newKitchenWithTasks は指定数のアクティブタスクと完了タスクを持つ Kitchen を作る。
//
// ReconstructKitchen + ReconstructCookingTask を使って構築する。
// なぜ AddCookingTask を10回呼ばないか:
//   - 10回呼ぶと10個のドメインイベントは発行されないが、テストの準備が冗長になる
//   - ReconstructKitchen は公開されたファクトリ関数なので、ブラックボックステストの範囲内
//   - テストの Given（前提条件）を効率的に構築するための正当な手段
func newKitchenWithTasks(t *testing.T, active, completed int) *kitchen.Kitchen {
	t.Helper()
	var tasks []kitchen.CookingTask
	for i := 0; i < active; i++ {
		tasks = append(tasks, kitchen.ReconstructCookingTask(
			kitchen.TaskID("task-active-"+string(rune('0'+i))),
			order.OrderID("order-active-"+string(rune('0'+i))),
			[]kitchen.CookingInstruction{{MenuName: "ラーメン"}},
			kitchen.TaskPending,
			time.Now(),
		))
	}
	for i := 0; i < completed; i++ {
		tasks = append(tasks, kitchen.ReconstructCookingTask(
			kitchen.TaskID("task-done-"+string(rune('0'+i))),
			order.OrderID("order-done-"+string(rune('0'+i))),
			[]kitchen.CookingInstruction{{MenuName: "ラーメン"}},
			kitchen.TaskCompleted,
			time.Now(),
		))
	}
	return kitchen.ReconstructKitchen("kitchen-1", tasks)
}
