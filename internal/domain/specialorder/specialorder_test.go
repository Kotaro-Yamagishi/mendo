package specialorder_test

import (
	"testing"

	"mendo/internal/domain/specialorder"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// NewSpecialOrder — 特別注文の作成
// =============================================================================

func Test_NewSpecialOrder(t *testing.T) {
	t.Parallel()

	// When: 特別注文を作成
	so := specialorder.NewSpecialOrder("so-1", "order-1", "特製つけ麺")

	// Then: pending 状態で作成され、SpecialOrderRequested イベントが発行される
	assert.Equal(t, specialorder.SpecialOrderID("so-1"), so.ID())
	assert.Equal(t, "order-1", so.OrderID())
	assert.Equal(t, "特製つけ麺", so.MenuName())
	assert.Equal(t, specialorder.StatusPending, so.Status())

	events := so.DomainEvents()
	require.Len(t, events, 1)
	requested, ok := events[0].(specialorder.SpecialOrderRequested)
	require.True(t, ok)
	assert.Equal(t, "order-1", requested.OrderID)
	assert.Equal(t, "特製つけ麺", requested.MenuName)
}

// =============================================================================
// Approve — 承認（プロセスマネージャーの自動遷移を含む）
// =============================================================================

func Test_Approve_正常系(t *testing.T) {
	t.Parallel()

	// Given: pending 状態
	so := newPendingSpecialOrder(t)
	eventsBefore := len(so.DomainEvents())

	// When: 承認
	err := so.Approve()

	// Then: StatusCooking まで自動遷移する
	// なぜ StatusApproved ではなく StatusCooking か:
	// プロセスマネージャーが「承認されたら自動的に調理を開始する」という
	// ビジネスルールを実装している。Approve() 内で2段階の状態遷移が起きる。
	require.NoError(t, err)
	assert.Equal(t, specialorder.StatusCooking, so.Status())

	// イベントが2つ発行される: SpecialOrderApproved + CookingDispatched
	// なぜ2つか: 「承認」と「調理開始」は別の事実。1つのイベントにまとめると、
	// 下流のサービスが「承認だけに反応したい」ケースに対応できなくなる。
	newEvents := so.DomainEvents()[eventsBefore:]
	require.Len(t, newEvents, 2, "Approve で2つのイベントが発行される")

	approved, isApproved := newEvents[0].(specialorder.SpecialOrderApproved)
	assert.True(t, isApproved, "1つ目は SpecialOrderApproved")
	assert.Equal(t, specialorder.EventTypeSpecialOrderApproved, approved.GetEventType())

	dispatched, isDispatched := newEvents[1].(specialorder.CookingDispatched)
	assert.True(t, isDispatched, "2つ目は CookingDispatched")
	assert.Equal(t, "order-1", dispatched.OrderID)
}

// ルール: pending 以外からは承認できない
func Test_Approve_異常系(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		setup   func(t *testing.T) *specialorder.SpecialOrder
		wantErr string
	}{
		{
			// cooking 状態（承認済み→調理中）から再承認は不可
			name:    "cooking状態_承認不可",
			setup:   newCookingSpecialOrder,
			wantErr: "承認待ち状態のみ",
		},
		{
			// rejected 状態から承認は不可（再申請が必要）
			name:    "rejected状態_承認不可",
			setup:   newRejectedSpecialOrder,
			wantErr: "承認待ち状態のみ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			so := tt.setup(t)
			eventsBefore := len(so.DomainEvents())

			err := so.Approve()

			require.ErrorContains(t, err, tt.wantErr)
			assert.Len(t, so.DomainEvents(), eventsBefore,
				"エラー時にイベントが発行されていないこと")
		})
	}
}

// =============================================================================
// Reject — 却下
// =============================================================================

func Test_Reject_正常系(t *testing.T) {
	t.Parallel()

	// Given: pending 状態
	so := newPendingSpecialOrder(t)
	eventsBefore := len(so.DomainEvents())

	// When: 却下（代替メニューを提案）
	err := so.Reject("材料切れ", "醤油ラーメン")

	// Then: rejected 状態になり、代替メニューが記録される
	require.NoError(t, err)
	assert.Equal(t, specialorder.StatusRejected, so.Status())
	assert.Equal(t, "醤油ラーメン", so.SuggestedMenu())

	newEvents := so.DomainEvents()[eventsBefore:]
	require.Len(t, newEvents, 1)
	rejected, ok := newEvents[0].(specialorder.SpecialOrderRejected)
	require.True(t, ok)
	assert.Equal(t, "材料切れ", rejected.Reason)
	assert.Equal(t, "醤油ラーメン", rejected.SuggestedMenu)
	assert.Equal(t, specialorder.EventTypeSpecialOrderRejected, rejected.GetEventType())
}

func Test_Reject_異常系(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		setup   func(t *testing.T) *specialorder.SpecialOrder
		wantErr string
	}{
		{
			name:    "cooking状態_却下不可",
			setup:   newCookingSpecialOrder,
			wantErr: "承認待ち状態のみ",
		},
		{
			name:    "rejected状態_二重却下不可",
			setup:   newRejectedSpecialOrder,
			wantErr: "承認待ち状態のみ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			so := tt.setup(t)
			eventsBefore := len(so.DomainEvents())

			err := so.Reject("理由", "代替")

			require.ErrorContains(t, err, tt.wantErr)
			assert.Len(t, so.DomainEvents(), eventsBefore)
		})
	}
}

// =============================================================================
// ResubmitWithMenu — 却下後の再申請
// =============================================================================

func Test_ResubmitWithMenu_正常系(t *testing.T) {
	t.Parallel()

	// Given: rejected 状態
	so := newRejectedSpecialOrder(t)
	eventsBefore := len(so.DomainEvents())

	// When: 別メニューで再申請
	err := so.ResubmitWithMenu("塩ラーメン")

	// Then: pending に戻り、メニュー名が更新される
	// なぜ pending に戻るか: 再申請は「新しい申請」として扱われ、
	// 再度店長の承認が必要になる。自動承認しない。
	require.NoError(t, err)
	assert.Equal(t, specialorder.StatusPending, so.Status())
	assert.Equal(t, "塩ラーメン", so.MenuName())

	newEvents := so.DomainEvents()[eventsBefore:]
	require.Len(t, newEvents, 1)
	resubmitted, ok := newEvents[0].(specialorder.MenuResubmitted)
	require.True(t, ok)
	assert.Equal(t, "塩ラーメン", resubmitted.NewMenuName)
}

func Test_ResubmitWithMenu_異常系(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		setup   func(t *testing.T) *specialorder.SpecialOrder
		wantErr string
	}{
		{
			// pending から再申請は意味がない（まだ却下されていない）
			name:    "pending状態_再申請不可",
			setup:   newPendingSpecialOrder,
			wantErr: "却下済みのみ",
		},
		{
			// cooking 中の再申請は不可（調理が始まっている）
			name:    "cooking状態_再申請不可",
			setup:   newCookingSpecialOrder,
			wantErr: "却下済みのみ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			so := tt.setup(t)
			eventsBefore := len(so.DomainEvents())

			err := so.ResubmitWithMenu("代替メニュー")

			require.ErrorContains(t, err, tt.wantErr)
			assert.Len(t, so.DomainEvents(), eventsBefore)
		})
	}
}

// =============================================================================
// Reject → ResubmitWithMenu → Approve の一連フロー
// =============================================================================

// 状態遷移のフルサイクルテスト。
// なぜ個別テストとは別にこれが必要か:
// 個別テストは各コマンドが正しく動くことを保証する。
// このテストは「却下→再申請→承認」という業務フロー全体が通ることを保証する。
// 個別テストだけでは、例えば「再申請後の pending は最初の pending と同じ振る舞いか？」
// という疑問に答えられない。
func Test_SpecialOrder_フルサイクル(t *testing.T) {
	t.Parallel()

	// 1. 作成（pending）
	so := specialorder.NewSpecialOrder("so-1", "order-1", "特製つけ麺")
	assert.Equal(t, specialorder.StatusPending, so.Status())

	// 2. 却下
	require.NoError(t, so.Reject("材料切れ", "醤油ラーメン"))
	assert.Equal(t, specialorder.StatusRejected, so.Status())

	// 3. 再申請（pending に戻る）
	require.NoError(t, so.ResubmitWithMenu("塩ラーメン"))
	assert.Equal(t, specialorder.StatusPending, so.Status())
	assert.Equal(t, "塩ラーメン", so.MenuName(), "再申請でメニュー名が更新される")

	// 4. 承認（cooking まで自動遷移）
	require.NoError(t, so.Approve())
	assert.Equal(t, specialorder.StatusCooking, so.Status())

	// イベント総数: Requested(1) + Rejected(1) + Resubmitted(1) + Approved(1) + Dispatched(1) = 5
	assert.Len(t, so.DomainEvents(), 5)
}

// =============================================================================
// ReconstructSpecialOrder — DB からの復元
// =============================================================================

func Test_ReconstructSpecialOrder(t *testing.T) {
	t.Parallel()

	so := specialorder.ReconstructSpecialOrder(
		"so-1", "order-1", "特製つけ麺",
		specialorder.StatusRejected, "醤油ラーメン",
	)

	assert.Equal(t, specialorder.SpecialOrderID("so-1"), so.ID())
	assert.Equal(t, specialorder.StatusRejected, so.Status())
	assert.Equal(t, "醤油ラーメン", so.SuggestedMenu())
	assert.Empty(t, so.DomainEvents(), "復元時はイベントを発行しない")
}

// =============================================================================
// テストヘルパー — 積み上げ式
// =============================================================================

func newPendingSpecialOrder(t *testing.T) *specialorder.SpecialOrder {
	t.Helper()
	return specialorder.NewSpecialOrder("so-1", "order-1", "特製つけ麺")
}

func newCookingSpecialOrder(t *testing.T) *specialorder.SpecialOrder {
	t.Helper()
	so := newPendingSpecialOrder(t)
	require.NoError(t, so.Approve())
	return so
}

func newRejectedSpecialOrder(t *testing.T) *specialorder.SpecialOrder {
	t.Helper()
	so := newPendingSpecialOrder(t)
	require.NoError(t, so.Reject("材料切れ", "醤油ラーメン"))
	return so
}

// =============================================================================
// SpecialOrderStatus.String() — ステータス文字列変換
// =============================================================================

func Test_SpecialOrderStatus_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		status specialorder.SpecialOrderStatus
		want   string
	}{
		{specialorder.StatusRequested, "requested"},
		{specialorder.StatusPending, "pending"},
		{specialorder.StatusApproved, "approved"},
		{specialorder.StatusRejected, "rejected"},
		{specialorder.StatusCooking, "cooking"},
		{specialorder.StatusCompleted, "completed"},
		{specialorder.SpecialOrderStatus(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.status.String())
		})
	}
}
