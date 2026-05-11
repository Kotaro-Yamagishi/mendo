//go:build integration

package e2e_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_SpecialOrderFlow_申請から承認まで は特別注文作成 → 承認のフローを通す E2E テスト。
func Test_SpecialOrderFlow_申請から承認まで(t *testing.T) {
	cleanupAllTestData(t)

	// --- Step 1: POST /special-orders — 特別注文作成 ---
	createBody := map[string]any{
		"order_id":  "order-001",
		"menu_name": "特製トリュフラーメン",
	}
	resp1 := postJSON(t, "/special-orders", createBody)
	assert.Equal(t, http.StatusCreated, resp1.StatusCode, "POST /special-orders should return 201")

	var createResult struct {
		SpecialOrderID string `json:"special_order_id"`
	}
	decodeResponse(t, resp1, &createResult)
	require.NotEmpty(t, createResult.SpecialOrderID, "special_order_id should not be empty")
	soID := createResult.SpecialOrderID
	t.Logf("created special_order_id: %s", soID)

	// --- Step 2: POST /special-orders/{id}/approve — 承認 ---
	resp2 := postJSON(t, "/special-orders/"+soID+"/approve", map[string]any{})
	assert.Equal(t, http.StatusOK, resp2.StatusCode, "POST /special-orders/{id}/approve should return 200")

	// DB検証: sc_special_orders の status が cooking（4）になっていること
	var status int
	err := testDB.QueryRowContext(context.Background(),
		`SELECT status FROM sc_special_orders WHERE id = ?`, soID,
	).Scan(&status)
	require.NoError(t, err)
	// StatusCooking = 4 (0:Requested, 1:Pending, 2:Approved, 3:Rejected, 4:Cooking)
	assert.Equal(t, 4, status, "special order status should be cooking (4)")
}

// Test_SpecialOrderFlow_却下から再申請 は特別注文作成 → 却下 → 再申請 → 承認のフローを通す E2E テスト。
func Test_SpecialOrderFlow_却下から再申請(t *testing.T) {
	cleanupAllTestData(t)

	// --- Step 1: POST /special-orders — 特別注文作成 ---
	resp1 := postJSON(t, "/special-orders", map[string]any{
		"order_id":  "order-002",
		"menu_name": "高級フォアグララーメン",
	})
	assert.Equal(t, http.StatusCreated, resp1.StatusCode, "POST /special-orders should return 201")

	var createResult struct {
		SpecialOrderID string `json:"special_order_id"`
	}
	decodeResponse(t, resp1, &createResult)
	require.NotEmpty(t, createResult.SpecialOrderID, "special_order_id should not be empty")
	soID := createResult.SpecialOrderID
	t.Logf("created special_order_id: %s", soID)

	// --- Step 2: POST /special-orders/{id}/reject — 却下 ---
	resp2 := postJSON(t, "/special-orders/"+soID+"/reject", map[string]any{
		"reason":         "食材の在庫なし",
		"suggested_menu": "特製味噌ラーメン",
	})
	assert.Equal(t, http.StatusOK, resp2.StatusCode, "POST /special-orders/{id}/reject should return 200")

	// --- Step 3: POST /special-orders/{id}/resubmit — 再申請 ---
	resp3 := postJSON(t, "/special-orders/"+soID+"/resubmit", map[string]any{
		"menu_name": "特製味噌ラーメン",
	})
	assert.Equal(t, http.StatusOK, resp3.StatusCode, "POST /special-orders/{id}/resubmit should return 200")

	// --- Step 4: POST /special-orders/{id}/approve — 承認 ---
	resp4 := postJSON(t, "/special-orders/"+soID+"/approve", map[string]any{})
	assert.Equal(t, http.StatusOK, resp4.StatusCode, "POST /special-orders/{id}/approve should return 200")

	// DB検証: sc_special_orders の status が cooking（4）になっていること
	var status int
	err := testDB.QueryRowContext(context.Background(),
		`SELECT status FROM sc_special_orders WHERE id = ?`, soID,
	).Scan(&status)
	require.NoError(t, err)
	// StatusCooking = 4
	assert.Equal(t, 4, status, "special order status should be cooking (4) after approve")
}
