//go:build integration

package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// postJSON は testServer に JSON ボディで POST リクエストを送る。
func postJSON(t *testing.T, path string, body any) *http.Response {
	t.Helper()
	b, err := json.Marshal(body)
	require.NoError(t, err)
	resp, err := http.Post(testServer.URL+path, "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	return resp
}

// decodeResponse はレスポンスボディを successResponse の data フィールドとしてデコードする。
func decodeResponse(t *testing.T, resp *http.Response, target any) {
	t.Helper()
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var wrapper struct {
		Data json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(raw, &wrapper))
	require.NoError(t, json.Unmarshal(wrapper.Data, target))
}

// parseResponse はレスポンスボディを map[string]interface{} として返す。
func parseResponse(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(raw, &result))
	return result
}

// countRows は指定テーブルの行数を返す。where は "col = ?" 形式、args はその引数。
func countRows(t *testing.T, table, where string, args ...any) int {
	t.Helper()
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table, where)
	var count int
	err := testDB.QueryRowContext(context.Background(), query, args...).Scan(&count)
	require.NoError(t, err)
	return count
}

// cleanupAllTestData はテスト後に全テーブルのデータを削除する。
func cleanupAllTestData(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		ctx := context.Background()
		// kc_kitchens と oc_menus は seed データのため削除しない
		tables := []string{"sc_special_orders", "kc_cooking_tasks", "outbox", "events"}
		for _, table := range tables {
			if _, err := testDB.ExecContext(ctx, "DELETE FROM "+table); err != nil {
				t.Logf("cleanup %s: %v", table, err)
			}
		}
	})
}
