package order_test

import (
	"testing"

	"mendo/internal/domain/contract"
	"mendo/internal/domain/order"

	"github.com/stretchr/testify/assert"
)

func TestToPublicConfirmed(t *testing.T) {
	t.Parallel()

	items := []order.ConfirmedItem{
		{MenuID: "ramen-a", Toppings: []string{"チャーシュー", "メンマ"}, Hardness: "hard"},
		{MenuID: "ramen-b", Toppings: nil, Hardness: "normal"},
	}
	internal := order.NewOrderConfirmed("order-100", items, 3, "corr-1")

	pub := order.ToPublicConfirmed(internal)

	assert.Equal(t, "order-100", pub.OrderID)
	assert.Equal(t, 3, pub.SeatNo)
	assert.Len(t, pub.Items, 2)

	// 1件目: MenuID が MenuName にマッピングされ、Toppings/Hardness がそのまま引き継がれる
	assert.Equal(t, "ramen-a", pub.Items[0].MenuName)
	assert.Equal(t, []string{"チャーシュー", "メンマ"}, pub.Items[0].Toppings)
	assert.Equal(t, "hard", pub.Items[0].Hardness)

	// 2件目: Toppings が nil のケース
	assert.Equal(t, "ramen-b", pub.Items[1].MenuName)
	assert.Nil(t, pub.Items[1].Toppings)
	assert.Equal(t, "normal", pub.Items[1].Hardness)
}

func TestToPublicConfirmed_EventType(t *testing.T) {
	t.Parallel()

	internal := order.NewOrderConfirmed("order-101", nil, 1, "corr-1")
	pub := order.ToPublicConfirmed(internal)

	assert.Equal(t, contract.PublicEventTypeOrderConfirmed, pub.GetEventType())
	assert.Equal(t, "order-101", pub.GetAggregateID())
}

func TestToPublicConfirmed_EmptyItems(t *testing.T) {
	t.Parallel()

	internal := order.NewOrderConfirmed("order-102", []order.ConfirmedItem{}, 5, "corr-1")
	pub := order.ToPublicConfirmed(internal)

	assert.Equal(t, "order-102", pub.OrderID)
	assert.Equal(t, 5, pub.SeatNo)
	assert.Empty(t, pub.Items)
}

func TestToPublicCanceled(t *testing.T) {
	t.Parallel()

	internal := order.NewOrderCancelled("order-200", "品切れのため", "corr-2")

	pub := order.ToPublicCanceled(internal)

	assert.Equal(t, "order-200", pub.OrderID)
	assert.Equal(t, "品切れのため", pub.Reason)
}

func TestToPublicCanceled_EventType(t *testing.T) {
	t.Parallel()

	internal := order.NewOrderCancelled("order-201", "理由", "corr-2")
	pub := order.ToPublicCanceled(internal)

	assert.Equal(t, contract.PublicEventTypeOrderCanceled, pub.GetEventType())
	assert.Equal(t, "order-201", pub.GetAggregateID())
}

func TestToPublicCanceled_EmptyReason(t *testing.T) {
	t.Parallel()

	internal := order.NewOrderCancelled("order-202", "", "corr-2")
	pub := order.ToPublicCanceled(internal)

	assert.Equal(t, "order-202", pub.OrderID)
	assert.Equal(t, "", pub.Reason)
}

func TestToPublicCreated(t *testing.T) {
	t.Parallel()

	internal := order.NewOrderCreated("order-300", 7, "corr-3")

	pub := order.ToPublicCreated(internal)

	assert.Equal(t, "order-300", pub.OrderID)
	assert.Equal(t, 7, pub.SeatNo)
}

func TestToPublicCreated_EventType(t *testing.T) {
	t.Parallel()

	internal := order.NewOrderCreated("order-301", 2, "corr-3")
	pub := order.ToPublicCreated(internal)

	assert.Equal(t, contract.PublicEventTypeOrderCreated, pub.GetEventType())
	assert.Equal(t, "order-301", pub.GetAggregateID())
}

func TestToPublicCreated_SeatNoPreserved(t *testing.T) {
	t.Parallel()

	for _, seatNo := range []int{1, 5, 10, 99} {
		seatNo := seatNo
		t.Run("seat"+string(rune('0'+seatNo%10)), func(t *testing.T) {
			t.Parallel()
			internal := order.NewOrderCreated("order-302", seatNo, "corr-3")
			pub := order.ToPublicCreated(internal)
			assert.Equal(t, seatNo, pub.SeatNo)
		})
	}
}

// --- 型アサーション: 公開イベントが contract.Event インタフェースを満たすか ---

func TestToPublicConfirmed_ImplementsContractEventInterface(t *testing.T) {
	t.Parallel()

	internal := order.NewOrderConfirmed("order-400", nil, 1, "corr-4")
	pub := order.ToPublicConfirmed(internal)

	// domain.Event インタフェースを満たすことを確認
	var _ interface {
		GetEventType() string
		GetAggregateID() string
		GetCorrelationID() string
	} = pub
}

func TestToPublicCanceled_ImplementsContractEventInterface(t *testing.T) {
	t.Parallel()

	internal := order.NewOrderCancelled("order-401", "理由", "corr-4")
	pub := order.ToPublicCanceled(internal)

	var _ interface {
		GetEventType() string
		GetAggregateID() string
		GetCorrelationID() string
	} = pub
}

func TestToPublicCreated_ImplementsContractEventInterface(t *testing.T) {
	t.Parallel()

	internal := order.NewOrderCreated("order-402", 3, "corr-4")
	pub := order.ToPublicCreated(internal)

	var _ interface {
		GetEventType() string
		GetAggregateID() string
		GetCorrelationID() string
	} = pub
}
