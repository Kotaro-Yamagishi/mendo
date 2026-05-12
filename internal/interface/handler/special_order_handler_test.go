package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appso "mendo/internal/application/command/specialorder"
	sodomain "mendo/internal/domain/specialorder"
	"mendo/internal/interface/handler"
	"mendo/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildSpecialOrderHandler は全ユースケースをまとめてハンドラを構築するヘルパー。
func buildSpecialOrderHandler(
	reader *testutil.StubSpecialOrderReader,
	writer *testutil.SpySpecialOrderWriter,
	pub *testutil.SpyEventPublisher,
) *handler.SpecialOrderHandler {
	createUC := appso.NewCreateSpecialOrderUsecase(writer, pub)
	approveUC := appso.NewApproveSpecialOrderUsecase(reader, writer, pub)
	rejectUC := appso.NewRejectSpecialOrderUsecase(reader, writer, pub)
	resubmitUC := appso.NewResubmitSpecialOrderUsecase(reader, writer, pub)
	return handler.NewSpecialOrderHandler(createUC, approveUC, rejectUC, resubmitUC)
}

// =============================================================================
// HandleCreate
// =============================================================================

func Test_SpecialOrderHandler_HandleCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		body           string
		wantStatus     int
		wantRespStatus string
		wantIDNotEmpty bool
	}{
		{
			name:           "正常系",
			body:           `{"order_id":"order-1","menu_name":"特製つけ麺"}`,
			wantStatus:     http.StatusCreated,
			wantRespStatus: "pending",
			wantIDNotEmpty: true,
		},
		{
			name:       "不正JSON",
			body:       "invalid json",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reader := &testutil.StubSpecialOrderReader{}
			writer := &testutil.SpySpecialOrderWriter{}
			pub := &testutil.SpyEventPublisher{}
			h := buildSpecialOrderHandler(reader, writer, pub)

			req := httptest.NewRequest(http.MethodPost, "/special-orders", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			wrap(h.HandleCreate)(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantRespStatus == "" {
				return
			}

			var resp map[string]interface{}
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
			data := resp["data"].(map[string]interface{})
			if tt.wantIDNotEmpty {
				assert.NotEmpty(t, data["special_order_id"])
			}
			assert.Equal(t, tt.wantRespStatus, data["status"])
		})
	}
}

// =============================================================================
// HandleApprove
// =============================================================================

func Test_SpecialOrderHandler_HandleApprove_正常系(t *testing.T) {
	t.Parallel()

	so := sodomain.NewSpecialOrder("so-1", "order-1", "特製つけ麺")
	reader := &testutil.StubSpecialOrderReader{SpecialOrder: so}
	writer := &testutil.SpySpecialOrderWriter{}
	pub := &testutil.SpyEventPublisher{}
	h := buildSpecialOrderHandler(reader, writer, pub)

	req := httptest.NewRequest(http.MethodPost, "/special-orders/so-1/approve", nil)
	req.SetPathValue("id", "so-1")
	rec := httptest.NewRecorder()

	wrap(h.HandleApprove)(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "approved_and_cooking", data["status"])
}

// =============================================================================
// HandleReject
// =============================================================================

func Test_SpecialOrderHandler_HandleReject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		body              string
		wantStatus        int
		wantRespStatus    string
		wantSuggestedMenu string
	}{
		{
			name:              "正常系",
			body:              `{"reason":"材料切れ","suggested_menu":"醤油ラーメン"}`,
			wantStatus:        http.StatusOK,
			wantRespStatus:    "rejected",
			wantSuggestedMenu: "醤油ラーメン",
		},
		{
			name:       "不正JSON",
			body:       "bad json",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			so := sodomain.NewSpecialOrder("so-1", "order-1", "特製つけ麺")
			reader := &testutil.StubSpecialOrderReader{SpecialOrder: so}
			writer := &testutil.SpySpecialOrderWriter{}
			pub := &testutil.SpyEventPublisher{}
			h := buildSpecialOrderHandler(reader, writer, pub)

			req := httptest.NewRequest(http.MethodPost, "/special-orders/so-1/reject", strings.NewReader(tt.body))
			req.SetPathValue("id", "so-1")
			rec := httptest.NewRecorder()

			wrap(h.HandleReject)(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantRespStatus == "" {
				return
			}

			var resp map[string]interface{}
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
			data := resp["data"].(map[string]interface{})
			assert.Equal(t, tt.wantRespStatus, data["status"])
			assert.Equal(t, tt.wantSuggestedMenu, data["suggested_menu"])
		})
	}
}

// =============================================================================
// HandleResubmit
// =============================================================================

func Test_SpecialOrderHandler_HandleResubmit(t *testing.T) {
	t.Parallel()

	rejectedSO := func() *sodomain.SpecialOrder {
		so := sodomain.NewSpecialOrder("so-1", "order-1", "特製つけ麺")
		require.NoError(t, so.Reject("材料切れ", "醤油ラーメン"))
		return so
	}

	tests := []struct {
		name           string
		so             func() *sodomain.SpecialOrder
		body           string
		wantStatus     int
		wantRespStatus string
	}{
		{
			name:           "正常系",
			so:             rejectedSO,
			body:           `{"menu_name":"塩ラーメン"}`,
			wantStatus:     http.StatusOK,
			wantRespStatus: "pending",
		},
		{
			name:       "不正JSON",
			so:         func() *sodomain.SpecialOrder { return sodomain.NewSpecialOrder("so-1", "order-1", "特製つけ麺") },
			body:       "bad json",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reader := &testutil.StubSpecialOrderReader{SpecialOrder: tt.so()}
			writer := &testutil.SpySpecialOrderWriter{}
			pub := &testutil.SpyEventPublisher{}
			h := buildSpecialOrderHandler(reader, writer, pub)

			req := httptest.NewRequest(http.MethodPost, "/special-orders/so-1/resubmit", strings.NewReader(tt.body))
			req.SetPathValue("id", "so-1")
			rec := httptest.NewRecorder()

			wrap(h.HandleResubmit)(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantRespStatus == "" {
				return
			}

			var resp map[string]interface{}
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
			data := resp["data"].(map[string]interface{})
			assert.Equal(t, tt.wantRespStatus, data["status"])
		})
	}
}
