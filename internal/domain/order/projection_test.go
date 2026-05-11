package order_test

import (
	"testing"
	"time"

	"mendo/internal/domain"
	"mendo/internal/domain/order"

	"github.com/stretchr/testify/assert"
)

// --- OrderStateProjection ---

func TestNewOrderStateProjection_OrderCreated(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	created := order.NewOrderCreated("order-001", 3, "corr-1")
	created.OccurredAt = now

	p := order.NewOrderStateProjection([]domain.Event{created})

	assert.Equal(t, "order-001", p.OrderID)
	assert.Equal(t, 3, p.SeatNo)
	assert.Equal(t, "pending", p.Status)
	assert.Equal(t, now, p.CreatedAt)
	assert.Equal(t, now, p.UpdatedAt)
	assert.Equal(t, 1, p.Version)
	assert.Empty(t, p.Items)
}

func TestNewOrderStateProjection_ItemAdded(t *testing.T) {
	t.Parallel()

	created := order.NewOrderCreated("order-002", 5, "corr-1")
	item1 := order.NewItemAdded("order-002", "ramen-a", []string{"チャーシュー", "メンマ"}, "hard", "corr-1")
	item2 := order.NewItemAdded("order-002", "ramen-b", nil, "normal", "corr-1")

	p := order.NewOrderStateProjection([]domain.Event{created, item1, item2})

	assert.Equal(t, "pending", p.Status)
	assert.Len(t, p.Items, 2)
	assert.Equal(t, "ramen-a", string(p.Items[0].MenuID))
	assert.Equal(t, []string{"チャーシュー", "メンマ"}, p.Items[0].Toppings)
	assert.Equal(t, "hard", p.Items[0].Hardness)
	assert.Equal(t, "ramen-b", string(p.Items[1].MenuID))
	assert.Nil(t, p.Items[1].Toppings)
	assert.Equal(t, "normal", p.Items[1].Hardness)
	assert.Equal(t, 3, p.Version)
}

func TestNewOrderStateProjection_OrderConfirmed(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)
	confirmedAt := createdAt.Add(5 * time.Minute)

	created := order.NewOrderCreated("order-003", 1, "corr-1")
	created.OccurredAt = createdAt

	confirmed := order.NewOrderConfirmed("order-003", nil, 1, "corr-1")
	confirmed.OccurredAt = confirmedAt

	p := order.NewOrderStateProjection([]domain.Event{created, confirmed})

	assert.Equal(t, "confirmed", p.Status)
	assert.Equal(t, createdAt, p.CreatedAt)
	assert.Equal(t, confirmedAt, p.UpdatedAt)
	assert.Equal(t, 2, p.Version)
}

func TestNewOrderStateProjection_OrderCancelled(t *testing.T) {
	t.Parallel()

	cancelledAt := time.Date(2026, 5, 5, 13, 0, 0, 0, time.UTC)

	created := order.NewOrderCreated("order-004", 2, "corr-1")
	cancelled := order.NewOrderCancelled("order-004", "品切れ", "corr-1")
	cancelled.OccurredAt = cancelledAt

	p := order.NewOrderStateProjection([]domain.Event{created, cancelled})

	assert.Equal(t, "canceled", p.Status)
	assert.Equal(t, cancelledAt, p.UpdatedAt)
	assert.Equal(t, 2, p.Version)
}

func TestNewOrderStateProjection_EmptyEvents(t *testing.T) {
	t.Parallel()

	p := order.NewOrderStateProjection([]domain.Event{})

	assert.Equal(t, "", p.OrderID)
	assert.Equal(t, "", p.Status)
	assert.Equal(t, 0, p.Version)
}

func TestOrderStateProjection_Apply_VersionIncrementsOnEachEvent(t *testing.T) {
	t.Parallel()

	created := order.NewOrderCreated("order-005", 7, "corr-1")
	item := order.NewItemAdded("order-005", "ramen-a", nil, "hard", "corr-1")
	confirmed := order.NewOrderConfirmed("order-005", nil, 7, "corr-1")

	p := order.NewOrderStateProjection([]domain.Event{created, item, confirmed})

	assert.Equal(t, 3, p.Version)
}

// --- OrderAnalyticsProjection ---

func TestNewOrderAnalyticsProjection_ItemCount(t *testing.T) {
	t.Parallel()

	created := order.NewOrderCreated("order-010", 3, "corr-1")
	item1 := order.NewItemAdded("order-010", "ramen-a", []string{"チャーシュー"}, "hard", "corr-1")
	item2 := order.NewItemAdded("order-010", "ramen-b", []string{"メンマ", "煮卵"}, "normal", "corr-1")
	item3 := order.NewItemAdded("order-010", "ramen-c", nil, "soft", "corr-1")

	p := order.NewOrderAnalyticsProjection([]domain.Event{created, item1, item2, item3})

	assert.Equal(t, 3, p.ItemCount)
	assert.Equal(t, 3, p.ToppingCount) // 1 + 2 + 0
}

func TestNewOrderAnalyticsProjection_ToppingCount_MultiplePerItem(t *testing.T) {
	t.Parallel()

	created := order.NewOrderCreated("order-011", 4, "corr-1")
	item := order.NewItemAdded("order-011", "ramen-a", []string{"a", "b", "c", "d"}, "hard", "corr-1")

	p := order.NewOrderAnalyticsProjection([]domain.Event{created, item})

	assert.Equal(t, 1, p.ItemCount)
	assert.Equal(t, 4, p.ToppingCount)
}

func TestNewOrderAnalyticsProjection_ToppingCount_NoToppings(t *testing.T) {
	t.Parallel()

	created := order.NewOrderCreated("order-015", 2, "corr-1")
	item := order.NewItemAdded("order-015", "ramen-a", nil, "normal", "corr-1")

	p := order.NewOrderAnalyticsProjection([]domain.Event{created, item})

	assert.Equal(t, 1, p.ItemCount)
	assert.Equal(t, 0, p.ToppingCount)
}

func TestNewOrderAnalyticsProjection_TimeToConfirm(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC)
	confirmedAt := createdAt.Add(3 * time.Minute)

	created := order.NewOrderCreated("order-012", 6, "corr-1")
	created.OccurredAt = createdAt

	confirmed := order.NewOrderConfirmed("order-012", nil, 6, "corr-1")
	confirmed.OccurredAt = confirmedAt

	p := order.NewOrderAnalyticsProjection([]domain.Event{created, confirmed})

	assert.Equal(t, "confirmed", p.Status)
	assert.Equal(t, 3*time.Minute, p.TimeToConfirm)
}

func TestNewOrderAnalyticsProjection_TimeToConfirm_ZeroWhenCancelled(t *testing.T) {
	t.Parallel()

	created := order.NewOrderCreated("order-013", 8, "corr-1")
	cancelled := order.NewOrderCancelled("order-013", "在庫なし", "corr-1")

	p := order.NewOrderAnalyticsProjection([]domain.Event{created, cancelled})

	assert.Equal(t, "canceled", p.Status)
	assert.Equal(t, time.Duration(0), p.TimeToConfirm)
}

func TestNewOrderAnalyticsProjection_MenuIDs(t *testing.T) {
	t.Parallel()

	created := order.NewOrderCreated("order-014", 2, "corr-1")
	item1 := order.NewItemAdded("order-014", "ramen-x", nil, "normal", "corr-1")
	item2 := order.NewItemAdded("order-014", "ramen-y", nil, "hard", "corr-1")

	p := order.NewOrderAnalyticsProjection([]domain.Event{created, item1, item2})

	assert.Len(t, p.MenuIDs, 2)
	assert.Equal(t, "ramen-x", string(p.MenuIDs[0]))
	assert.Equal(t, "ramen-y", string(p.MenuIDs[1]))
}

func TestNewOrderAnalyticsProjection_StatusPending(t *testing.T) {
	t.Parallel()

	created := order.NewOrderCreated("order-016", 9, "corr-1")

	p := order.NewOrderAnalyticsProjection([]domain.Event{created})

	assert.Equal(t, "pending", p.Status)
	assert.Equal(t, "order-016", p.OrderID)
}
