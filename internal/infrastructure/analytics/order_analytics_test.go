package analytics_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"mendo/internal/domain/order"
	"mendo/internal/infrastructure/analytics"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newConfirmedEvent は OrderConfirmed を生成してポインタで返す。
// applyOrderConfirmed 内で event.(*order.OrderConfirmed) とポインタアサーションするため、
// HandleOrderEvent には *order.OrderConfirmed を渡す必要がある。
func newConfirmedEvent(orderID string, items []order.ConfirmedItem, seatNo int) *order.OrderConfirmed {
	e := order.NewOrderConfirmed(orderID, items, seatNo, "corr-001")
	return &e
}

func TestHandleOrderEvent_SingleItem_GeneratesFact(t *testing.T) {
	t.Parallel()
	proj := analytics.NewOrderAnalyticsProjection()
	ctx := context.Background()

	items := []order.ConfirmedItem{
		{MenuID: "menu-ramen", Toppings: []string{"egg"}, Hardness: "normal"},
	}
	event := newConfirmedEvent("order-001", items, 3)

	err := proj.HandleOrderEvent(ctx, event)
	require.NoError(t, err)

	facts := proj.GetFacts()
	require.Len(t, facts, 1)

	fact := facts[0]
	assert.Equal(t, "order-001", fact.OrderID)
	assert.Equal(t, "menu-ramen", fact.MenuKey)
	assert.Equal(t, 3, fact.SeatKey)
	assert.Equal(t, 1, fact.Quantity)
}

func TestHandleOrderEvent_MultipleItems_GeneratesFactPerItem(t *testing.T) {
	t.Parallel()
	proj := analytics.NewOrderAnalyticsProjection()
	ctx := context.Background()

	items := []order.ConfirmedItem{
		{MenuID: "menu-ramen", Toppings: []string{"egg"}, Hardness: "normal"},
		{MenuID: "menu-gyoza", Toppings: nil, Hardness: ""},
		{MenuID: "menu-ramen", Toppings: nil, Hardness: "hard"}, // 同一メニューが複数
	}
	event := newConfirmedEvent("order-002", items, 5)

	err := proj.HandleOrderEvent(ctx, event)
	require.NoError(t, err)

	facts := proj.GetFacts()
	assert.Len(t, facts, 3, "アイテム数分の fact が生成されること")
}

func TestHandleOrderEvent_FactIDIsUnique(t *testing.T) {
	t.Parallel()
	proj := analytics.NewOrderAnalyticsProjection()
	ctx := context.Background()

	items := []order.ConfirmedItem{
		{MenuID: "menu-ramen"},
		{MenuID: "menu-ramen"}, // 同一メニューを2個
		{MenuID: "menu-gyoza"},
	}
	event := newConfirmedEvent("order-003", items, 1)

	err := proj.HandleOrderEvent(ctx, event)
	require.NoError(t, err)

	facts := proj.GetFacts()
	require.Len(t, facts, 3)

	// FactID が全て異なることを確認
	ids := make(map[string]struct{}, len(facts))
	for _, f := range facts {
		_, dup := ids[f.FactID]
		assert.False(t, dup, "FactID が重複している: %s", f.FactID)
		ids[f.FactID] = struct{}{}
	}

	// FactID のフォーマットが "OrderID-MenuID-index" 形式であることを確認
	assert.Equal(t, "order-003-menu-ramen-0", facts[0].FactID)
	assert.Equal(t, "order-003-menu-ramen-1", facts[1].FactID)
	assert.Equal(t, "order-003-menu-gyoza-2", facts[2].FactID)
}

func TestHandleOrderEvent_SkipsNonConfirmedEvents(t *testing.T) {
	t.Parallel()
	proj := analytics.NewOrderAnalyticsProjection()
	ctx := context.Background()

	// OrderCreated は分析不要なのでスキップされる
	created := order.NewOrderCreated("order-004", 2, "corr-004")
	err := proj.HandleOrderEvent(ctx, &created)
	require.NoError(t, err, "不要なイベントはエラーなくスキップされること")

	facts := proj.GetFacts()
	assert.Empty(t, facts, "OrderCreated では fact が生成されないこと")
}

func TestHandleOrderEvent_MultipleEvents_AccumulatesFacts(t *testing.T) {
	t.Parallel()
	proj := analytics.NewOrderAnalyticsProjection()
	ctx := context.Background()

	// 1件目の注文
	e1 := newConfirmedEvent("order-010", []order.ConfirmedItem{
		{MenuID: "menu-ramen"},
	}, 1)
	require.NoError(t, proj.HandleOrderEvent(ctx, e1))

	// 2件目の注文
	e2 := newConfirmedEvent("order-011", []order.ConfirmedItem{
		{MenuID: "menu-gyoza"},
		{MenuID: "menu-cola"},
	}, 2)
	require.NoError(t, proj.HandleOrderEvent(ctx, e2))

	facts := proj.GetFacts()
	assert.Len(t, facts, 3, "2つのイベント合計3アイテム分の fact が蓄積されること")
}

func TestToDateKey_Format(t *testing.T) {
	t.Parallel()
	// toDateKey は非公開関数なので HandleOrderEvent 経由で間接検証する。
	// テスト実行日の日付キーが YYYYMMDD 形式になっていることを確認する。
	proj := analytics.NewOrderAnalyticsProjection()
	ctx := context.Background()

	event := newConfirmedEvent("order-100", []order.ConfirmedItem{
		{MenuID: "menu-ramen"},
	}, 1)

	err := proj.HandleOrderEvent(ctx, event)
	require.NoError(t, err)

	facts := proj.GetFacts()
	require.Len(t, facts, 1)

	now := time.Now()
	expectedDateKey := now.Year()*10000 + int(now.Month())*100 + now.Day()
	assert.Equal(t, expectedDateKey, facts[0].DateKey,
		fmt.Sprintf("DateKey は YYYYMMDD 形式であること。期待値: %d", expectedDateKey))
}
