//go:build integration

package e2e_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mendo/internal/domain/kitchen"
	"mendo/internal/domain/order"
)

// Test_OrderFlow_注文から調理完了まで は注文作成 → 確定 → 調理完了の主要フローを通す E2E テスト。
func Test_OrderFlow_注文から調理完了まで(t *testing.T) {
	cleanupAllTestData(t)

	ctx := context.Background()

	// --- Step 1: POST /orders — 注文作成 ---
	createBody := map[string]any{
		"seat_no": 3,
		"items": []map[string]any{
			{
				"menu_id":  "menu-ramen-001",
				"toppings": []string{"nori", "egg"},
				"hardness": "medium",
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

	// DB検証: events に OrderCreated + ItemAdded が記録されているか
	var eventCount int
	err := testDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM events WHERE aggregate_id = ?`, orderID,
	).Scan(&eventCount)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, eventCount, 2, "events should have OrderCreated and ItemAdded")

	// DB検証: outbox に未配信イベントが記録されているか
	var outboxCount int
	err = testDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM outbox WHERE aggregate_id = ? AND delivered = FALSE`, orderID,
	).Scan(&outboxCount)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, outboxCount, 1, "outbox should have undelivered events")

	// --- Step 2: POST /orders/{id}/confirm — 注文確定 ---
	resp2 := postJSON(t, "/orders/"+orderID+"/confirm", map[string]any{})

	assert.Equal(t, http.StatusOK, resp2.StatusCode, "POST /orders/{id}/confirm should return 200")

	var confirmResult struct {
		Status string `json:"status"`
	}
	decodeResponse(t, resp2, &confirmResult)
	assert.Equal(t, "confirmed", confirmResult.Status)

	// DB検証: events に OrderConfirmed が追加されているか
	var confirmedCount int
	err = testDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM events WHERE aggregate_id = ? AND event_type = ?`, orderID, order.EventTypeOrderConfirmed,
	).Scan(&confirmedCount)
	require.NoError(t, err)
	assert.Equal(t, 1, confirmedCount, "events should have exactly one order.confirmed event")

	// DB検証: StartCooking により kc_cooking_tasks にタスクが作成されているか
	var taskCount int
	err = testDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM kc_cooking_tasks WHERE order_id = ?`, orderID,
	).Scan(&taskCount)
	require.NoError(t, err)
	assert.Equal(t, 1, taskCount, "kc_cooking_tasks should have one task for this order")

	// --- Step 3: POST /kitchen/complete — 調理完了 ---
	resp3 := postJSON(t, "/kitchen/complete", map[string]any{
		"order_id": orderID,
	})

	assert.Equal(t, http.StatusOK, resp3.StatusCode, "POST /kitchen/complete should return 200")

	var completeResult struct {
		Status string `json:"status"`
	}
	decodeResponse(t, resp3, &completeResult)
	assert.Equal(t, "cooking_completed", completeResult.Status)

	// DB検証: kc_cooking_tasks のステータスが completed になっているか
	var taskStatus string
	err = testDB.QueryRowContext(ctx,
		`SELECT status FROM kc_cooking_tasks WHERE order_id = ?`, orderID,
	).Scan(&taskStatus)
	require.NoError(t, err)
	assert.Equal(t, "completed", taskStatus, "cooking task status should be completed")

	// DB検証: outbox に CookingCompleted イベントが追加されているか
	var cookingCompletedCount int
	err = testDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM outbox WHERE event_type = ?`, kitchen.EventTypeCookingCompleted,
	).Scan(&cookingCompletedCount)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, cookingCompletedCount, 1, "outbox should have CookingCompleted event")
}
