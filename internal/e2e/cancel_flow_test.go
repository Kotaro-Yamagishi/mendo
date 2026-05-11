//go:build integration

package e2e_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mendo/internal/domain/order"
)

// Test_CancelFlow_注文確定後キャンセル は注文作成 → 確定 → キャンセルのフローを通す E2E テスト。
func Test_CancelFlow_注文確定後キャンセル(t *testing.T) {
	cleanupAllTestData(t)

	// --- Step 1: POST /orders — 注文作成 ---
	createBody := map[string]any{
		"seat_no": 5,
		"items": []map[string]any{
			{
				"menu_id":  "menu-ramen-001",
				"toppings": []string{"nori"},
				"hardness": "hard",
			},
		},
	}
	resp1 := postJSON(t, "/orders", createBody)
	assert.Equal(t, http.StatusCreated, resp1.StatusCode, "POST /orders should return 201")

	var createResult struct {
		OrderID string `json:"order_id"`
	}
	decodeResponse(t, resp1, &createResult)
	require.NotEmpty(t, createResult.OrderID, "order_id should not be empty")
	orderID := createResult.OrderID
	t.Logf("created order_id: %s", orderID)

	// --- Step 2: POST /orders/{id}/confirm — 注文確定 ---
	resp2 := postJSON(t, "/orders/"+orderID+"/confirm", map[string]any{})
	assert.Equal(t, http.StatusOK, resp2.StatusCode, "POST /orders/{id}/confirm should return 200")

	// --- Step 3: POST /orders/{id}/cancel — キャンセル ---
	resp3 := postJSON(t, "/orders/"+orderID+"/cancel", map[string]any{
		"reason": "客都合",
	})
	assert.Equal(t, http.StatusOK, resp3.StatusCode, "POST /orders/{id}/cancel should return 200")

	// DB検証: events テーブルに order.canceled イベントがあること
	canceledCount := countRows(t, "events", "aggregate_id = ? AND event_type = ?", orderID, order.EventTypeOrderCanceled)
	assert.Equal(t, 1, canceledCount, "events should have exactly one order.canceled event")
}
