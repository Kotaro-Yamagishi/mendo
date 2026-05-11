package order_test

import (
	"testing"

	"mendo/internal/domain"
	"mendo/internal/domain/menu"
	"mendo/internal/domain/order"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Create — 注文の作成
// =============================================================================

func Test_Create(t *testing.T) {
	t.Parallel()

	// When: 注文を作成する
	o := order.Create("order-1", 3)

	// Then: pending 状態で作成され、OrderCreated イベントが1つ発行される
	assert.Equal(t, "order-1", o.ID())
	assert.Equal(t, order.StatusPending, o.Status())

	// イベント検証: イベントソーシングでは「何が起きたか」の記録が本体。
	// 状態は apply() の副産物であり、イベントが正しければ状態も正しくなる。
	events := o.UncommittedEvents()
	require.Len(t, events, 1, "Create で発行されるイベントは1つ")
	created, ok := events[0].(order.OrderCreated)
	require.True(t, ok, "イベントの型は OrderCreated であるべき")
	assert.Equal(t, "order-1", created.AggregateID)
	assert.Equal(t, 3, created.SeatNo)
	assert.Equal(t, order.EventTypeOrderCreated, created.EventType)
}

// =============================================================================
// AddItem — 商品の追加
// =============================================================================

func Test_AddItem_正常系(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		toppings []string
		hardness string
	}{
		// 境界値: トッピング0個。最小値。
		{"トッピングなし", nil, "普通"},
		// 境界値: トッピング3個。上限ちょうど。
		// ルールが「3つまで」なので、3はOK、4がNGの境界。
		{"トッピング3個_上限", []string{"ネギ", "チャーシュー", "味玉"}, "硬め"},
		// 代表値: トッピング1個。典型的な入力。
		{"トッピング1個", []string{"ネギ"}, "柔らかめ"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			o := order.Create("order-1", 1)
			initialEventCount := len(o.UncommittedEvents())

			err := o.AddItem(menu.MenuID("menu-1"), tt.toppings, tt.hardness)

			require.NoError(t, err)
			// Create のイベント + AddItem のイベント = initialEventCount + 1
			events := o.UncommittedEvents()
			assert.Len(t, events, initialEventCount+1)
			added, ok := events[len(events)-1].(order.ItemAdded)
			require.True(t, ok)
			assert.Equal(t, menu.MenuID("menu-1"), added.MenuID)
		})
	}
}

func Test_AddItem_異常系(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		toppings []string
		wantErr  string
	}{
		// 境界値: トッピング4個。上限+1。
		{"トッピング4個_上限超過", []string{"ネギ", "チャーシュー", "味玉", "海苔"}, "3つまで"},
		// 大きな違反値。
		{"トッピング大量", []string{"a", "b", "c", "d", "e"}, "3つまで"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			o := order.Create("order-1", 1)
			initialEventCount := len(o.UncommittedEvents())

			err := o.AddItem(menu.MenuID("menu-1"), tt.toppings, "普通")

			require.ErrorContains(t, err, tt.wantErr)
			// エラー時はイベントが増えないことを検証。
			// なぜ: イベントソーシングでは「イベント = 事実の記録」。
			// 失敗した操作はイベントとして記録されてはいけない。
			assert.Len(t, o.UncommittedEvents(), initialEventCount,
				"エラー時にイベントが発行されていないこと")
		})
	}
}

// =============================================================================
// Confirm — 注文の確定
// =============================================================================

// ルール1: pending 状態でのみ確定できる
func Test_Confirm_ステータスチェック(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		setup   func(t *testing.T) *order.Order
		wantErr string
	}{
		{
			// 正常系: pending + アイテムあり → 成功
			name:    "pending_アイテムあり_成功",
			setup:   newPendingWithItems,
			wantErr: "",
		},
		{
			// 異常系: confirmed 状態 → エラー
			// なぜテスト: 二重確定は業務上ありえない。防止しないと重複イベントが発行される
			name:    "confirmed_二重確定_エラー",
			setup:   newConfirmedOrder,
			wantErr: "確定待ち以外",
		},
		{
			// 異常系: canceled 状態 → エラー
			// なぜテスト: キャンセル済み注文の復活は許されない
			name:    "canceled_確定不可_エラー",
			setup:   newCanceledOrder,
			wantErr: "確定待ち以外",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			o := tt.setup(t)
			eventsBefore := len(o.UncommittedEvents())

			err := o.Confirm()

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				assert.Len(t, o.UncommittedEvents(), eventsBefore,
					"エラー時にイベントが発行されていないこと")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, order.StatusConfirmed, o.Status())

			// Confirm 成功時のイベント検証
			events := o.UncommittedEvents()
			lastEvent := events[len(events)-1]
			confirmed, ok := lastEvent.(order.OrderConfirmed)
			require.True(t, ok, "最後のイベントは OrderConfirmed であるべき")
			assert.Equal(t, o.ID(), confirmed.AggregateID)
			assert.NotEmpty(t, confirmed.Items, "確定した注文明細がイベントに含まれるべき")
		})
	}
}

// ルール2: アイテムが1つ以上必要
func Test_Confirm_アイテム必須(t *testing.T) {
	t.Parallel()

	// Given: pending だがアイテムなし
	o := order.Create("order-1", 1)

	// When
	err := o.Confirm()

	// Then: アイテムがないので確定できない
	require.ErrorContains(t, err, "注文が空")
}

// =============================================================================
// Cancel — 注文のキャンセル
// =============================================================================

// ルール: confirmed 状態でのみキャンセルできる
func Test_Cancel_ステータスチェック(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		setup   func(t *testing.T) *order.Order
		wantErr string
	}{
		{
			name:    "confirmed_キャンセル成功",
			setup:   newConfirmedOrder,
			wantErr: "",
		},
		{
			// pending からのキャンセルは不可。
			// なぜ: この設計では「まず確定してからキャンセル」という業務フロー。
			// pending は「まだ注文が固まっていない状態」なのでキャンセルという概念がない。
			name:    "pending_キャンセル不可_エラー",
			setup:   newPendingWithItems,
			wantErr: "確定済みのみ",
		},
		{
			// 二重キャンセル防止。
			name:    "canceled_二重キャンセル_エラー",
			setup:   newCanceledOrder,
			wantErr: "確定済みのみ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			o := tt.setup(t)
			eventsBefore := len(o.UncommittedEvents())

			err := o.Cancel("テスト理由")

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				assert.Len(t, o.UncommittedEvents(), eventsBefore)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, order.StatusCanceled, o.Status())

			events := o.UncommittedEvents()
			lastEvent := events[len(events)-1]
			cancelled, ok := lastEvent.(order.OrderCancelled)
			require.True(t, ok, "最後のイベントは OrderCancelled であるべき")
			assert.Equal(t, "テスト理由", cancelled.Reason)
		})
	}
}

// =============================================================================
// ReconstructFromEvents — イベント列からの状態復元
// =============================================================================

// イベントソーシングの核心テスト。
// 「同じイベント列を apply すれば、必ず同じ状態に復元される」ことを保証する。
// なぜ重要: DB に状態を保存しないので、復元が壊れる = 全データが読めなくなる。

func Test_ReconstructFromEvents_確定済み注文(t *testing.T) {
	t.Parallel()

	// Given: Create → AddItem → Confirm のイベント列
	events := []domain.Event{
		order.NewOrderCreated("order-1", 3, ""),
		order.NewItemAdded("order-1", menu.MenuID("menu-1"), []string{"ネギ"}, "硬め", ""),
		order.NewOrderConfirmed("order-1", []order.ConfirmedItem{
			{MenuID: "menu-1", Toppings: []string{"ネギ"}, Hardness: "硬め"},
		}, 3, ""),
	}

	// When: イベント列から復元
	o := order.ReconstructFromEvents(events)

	// Then: 全状態が正しく復元される
	assert.Equal(t, "order-1", o.ID())
	assert.Equal(t, order.StatusConfirmed, o.Status())
	assert.Equal(t, 3, o.Version(), "イベント3つ → version=3")
	// 復元時は未コミットイベントがないこと。
	// なぜ: これは「既に保存済みのイベント」を再生しているので、新たに保存する必要がない。
	assert.Empty(t, o.UncommittedEvents(), "復元時は未コミットイベントが空であるべき")
}

func Test_ReconstructFromEvents_キャンセル済み注文(t *testing.T) {
	t.Parallel()

	events := []domain.Event{
		order.NewOrderCreated("order-2", 1, ""),
		order.NewItemAdded("order-2", menu.MenuID("menu-1"), nil, "普通", ""),
		order.NewOrderConfirmed("order-2", []order.ConfirmedItem{
			{MenuID: "menu-1", Toppings: nil, Hardness: "普通"},
		}, 1, ""),
		order.NewOrderCancelled("order-2", "客都合", ""),
	}

	o := order.ReconstructFromEvents(events)

	assert.Equal(t, order.StatusCanceled, o.Status())
	assert.Equal(t, 4, o.Version(), "イベント4つ → version=4")
}

func Test_ReconstructFromEvents_複数アイテム(t *testing.T) {
	t.Parallel()

	// Given: アイテムを3つ追加した注文
	events := []domain.Event{
		order.NewOrderCreated("order-3", 5, ""),
		order.NewItemAdded("order-3", menu.MenuID("menu-1"), nil, "普通", ""),
		order.NewItemAdded("order-3", menu.MenuID("menu-2"), []string{"ネギ"}, "硬め", ""),
		order.NewItemAdded("order-3", menu.MenuID("menu-3"), []string{"味玉", "チャーシュー"}, "柔らかめ", ""),
	}

	o := order.ReconstructFromEvents(events)

	assert.Equal(t, order.StatusPending, o.Status())
	assert.Equal(t, 4, o.Version(), "イベント4つ → version=4")
}

// =============================================================================
// ClearEvents — 未コミットイベントのクリア
// =============================================================================

func Test_ClearEvents(t *testing.T) {
	t.Parallel()

	// Given: イベントが溜まっている注文
	o := order.Create("order-1", 1)
	require.NotEmpty(t, o.UncommittedEvents())

	// When: イベントをクリア（Save 後に呼ばれる想定）
	o.ClearEvents()

	// Then
	assert.Empty(t, o.UncommittedEvents())
}

// =============================================================================
// テストヘルパー — 積み上げ式
//
// 各ヘルパーは前段の状態を作って次のコマンドを実行する。
// 状態遷移の順序がコードに表現される:
//   newPendingWithItems → newConfirmedOrder → newCanceledOrder
//
// なぜ積み上げ式にするか:
//   1. 各状態の作り方が一箇所に集約される（DRY）
//   2. 状態遷移のルールが変わったとき、修正箇所が最小になる
//   3. 不正な状態のテストデータが作れない（公開 API を経由するので）
// =============================================================================

// newPendingWithItems は pending 状態でアイテムが1つある注文を作る。
func newPendingWithItems(t *testing.T) *order.Order {
	t.Helper()
	o := order.Create("order-1", 1)
	err := o.AddItem(menu.MenuID("menu-1"), []string{"ネギ"}, "普通")
	require.NoError(t, err)
	return o
}

// newConfirmedOrder は confirmed 状態の注文を作る。
func newConfirmedOrder(t *testing.T) *order.Order {
	t.Helper()
	o := newPendingWithItems(t)
	err := o.Confirm()
	require.NoError(t, err)
	return o
}

// newCanceledOrder は canceled 状態の注文を作る。
func newCanceledOrder(t *testing.T) *order.Order {
	t.Helper()
	o := newConfirmedOrder(t)
	err := o.Cancel("テストキャンセル")
	require.NoError(t, err)
	return o
}
