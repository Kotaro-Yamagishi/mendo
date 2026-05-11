package contract_test

import (
	"encoding/json"
	"testing"

	"mendo/internal/domain/contract"
	kitchendomain "mendo/internal/domain/kitchen"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// OrderConfirmedPublic
// =============================================================================

func Test_OrderConfirmedPublic_スキーマ(t *testing.T) {
	t.Parallel()

	event := contract.OrderConfirmedPublic{
		OrderID: "order-1",
		SeatNo:  3,
		Items: []contract.OrderConfirmedPublicItem{
			{MenuName: "醤油ラーメン", Toppings: []string{"ネギ", "チャーシュー"}, Hardness: "硬め"},
		},
	}

	assert.Equal(t, contract.PublicEventTypeOrderConfirmed, event.GetEventType())
	assert.Equal(t, "order-1", event.GetAggregateID())
	assert.NotEmpty(t, event.OrderID)
	assert.NotZero(t, event.SeatNo)
	require.Len(t, event.Items, 1)
	assert.NotEmpty(t, event.Items[0].MenuName)
	assert.NotEmpty(t, event.Items[0].Toppings)
	assert.NotEmpty(t, event.Items[0].Hardness)
}

func Test_OrderConfirmedPublic_JSONラウンドトリップ(t *testing.T) {
	t.Parallel()

	original := contract.OrderConfirmedPublic{
		OrderID: "order-1",
		SeatNo:  3,
		Items: []contract.OrderConfirmedPublicItem{
			{MenuName: "醤油ラーメン", Toppings: []string{"ネギ"}, Hardness: "硬め"},
			{MenuName: "味噌ラーメン", Toppings: nil, Hardness: "普通"},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored contract.OrderConfirmedPublic
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, original.OrderID, restored.OrderID)
	assert.Equal(t, original.SeatNo, restored.SeatNo)
	require.Len(t, restored.Items, 2)
	assert.Equal(t, "醤油ラーメン", restored.Items[0].MenuName)
	assert.Equal(t, []string{"ネギ"}, restored.Items[0].Toppings)
	assert.Equal(t, "硬め", restored.Items[0].Hardness)
	assert.Equal(t, "味噌ラーメン", restored.Items[1].MenuName)
	assert.Equal(t, "普通", restored.Items[1].Hardness)
}

func Test_OrderConfirmedPublic_Kitchen消費側互換性(t *testing.T) {
	t.Parallel()

	eventJSON := `{
		"order_id": "order-1",
		"seat_no": 3,
		"items": [
			{"menu_name": "醤油ラーメン", "toppings": ["ネギ"], "hardness": "硬め"}
		]
	}`

	var event contract.OrderConfirmedPublic
	require.NoError(t, json.Unmarshal([]byte(eventJSON), &event))

	instructions := make([]kitchendomain.CookingInstruction, len(event.Items))
	for i, item := range event.Items {
		instructions[i] = kitchendomain.CookingInstruction{
			MenuName: item.MenuName,
			Toppings: item.Toppings,
			Hardness: item.Hardness,
		}
	}

	require.Len(t, instructions, 1)
	assert.Equal(t, "醤油ラーメン", instructions[0].MenuName)
	assert.Equal(t, []string{"ネギ"}, instructions[0].Toppings)
	assert.Equal(t, "硬め", instructions[0].Hardness)
}

func Test_OrderConfirmedPublic_後方互換性_未知フィールド無視(t *testing.T) {
	t.Parallel()

	futureJSON := `{
		"order_id": "order-1",
		"seat_no": 3,
		"items": [{"menu_name": "ラーメン", "toppings": [], "hardness": "普通"}],
		"priority": "high",
		"estimated_cooking_time": 300
	}`

	var event contract.OrderConfirmedPublic
	err := json.Unmarshal([]byte(futureJSON), &event)

	require.NoError(t, err)
	assert.Equal(t, "order-1", event.OrderID)
	assert.Len(t, event.Items, 1)
}

func Test_OrderConfirmedPublic_後方互換性_フィールド欠損(t *testing.T) {
	t.Parallel()

	oldJSON := `{
		"order_id": "order-1",
		"items": [{"menu_name": "ラーメン", "toppings": [], "hardness": "普通"}]
	}`

	var event contract.OrderConfirmedPublic
	err := json.Unmarshal([]byte(oldJSON), &event)

	require.NoError(t, err)
	assert.Equal(t, "order-1", event.OrderID)
	assert.Equal(t, 0, event.SeatNo)
}

// =============================================================================
// OrderCanceledPublic
// =============================================================================

func Test_OrderCanceledPublic_スキーマ(t *testing.T) {
	t.Parallel()

	event := contract.OrderCanceledPublic{OrderID: "order-1", Reason: "客都合"}

	assert.Equal(t, contract.PublicEventTypeOrderCanceled, event.GetEventType())
	assert.Equal(t, "order-1", event.GetAggregateID())
	assert.NotEmpty(t, event.Reason)
}

func Test_OrderCanceledPublic_JSONラウンドトリップ(t *testing.T) {
	t.Parallel()

	original := contract.OrderCanceledPublic{OrderID: "order-1", Reason: "客都合"}
	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored contract.OrderCanceledPublic
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, original, restored)
}

// =============================================================================
// OrderCreatedPublic
// =============================================================================

func Test_OrderCreatedPublic_スキーマ(t *testing.T) {
	t.Parallel()

	event := contract.OrderCreatedPublic{OrderID: "order-1", SeatNo: 3}

	assert.Equal(t, contract.PublicEventTypeOrderCreated, event.GetEventType())
	assert.Equal(t, "order-1", event.GetAggregateID())
}
