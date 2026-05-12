package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appkitchen "mendo/internal/application/command/kitchen"
	kitchendomain "mendo/internal/domain/kitchen"
	"mendo/internal/domain/order"
	"mendo/internal/interface/handler"
	"mendo/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const kitchenHandlerTestKitchenID = kitchendomain.KitchenID("kitchen-1")

func buildCompleteCookingHandler(k *kitchendomain.Kitchen) *handler.KitchenHandler {
	reader := &testutil.StubKitchenReader{Kitchen: k}
	writer := &testutil.SpyKitchenWriter{}
	outbox := &testutil.SpyOutbox{}
	tx := &testutil.StubTransactionManager{}
	uc := appkitchen.NewCompleteCookingUsecase(reader, writer, outbox, tx, kitchenHandlerTestKitchenID)
	return handler.NewKitchenHandler(uc)
}

func Test_KitchenHandler_HandleCompleteCooking(t *testing.T) {
	t.Parallel()

	kitchenWithTask := func() *kitchendomain.Kitchen {
		k := kitchendomain.NewKitchen(kitchenHandlerTestKitchenID)
		require.NoError(t, k.AddCookingTask(order.OrderID("order-1"), []kitchendomain.CookingInstruction{{MenuName: "醤油ラーメン"}}))
		return k
	}

	tests := []struct {
		name           string
		kitchen        func() *kitchendomain.Kitchen
		body           string
		wantStatus     int
		wantRespStatus string
	}{
		{
			name:           "正常系",
			kitchen:        kitchenWithTask,
			body:           `{"order_id":"order-1"}`,
			wantStatus:     http.StatusOK,
			wantRespStatus: "cooking_completed",
		},
		{
			name:       "不正JSON",
			kitchen:    func() *kitchendomain.Kitchen { return kitchendomain.NewKitchen(kitchenHandlerTestKitchenID) },
			body:       "invalid json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "業務エラー",
			kitchen:    func() *kitchendomain.Kitchen { return kitchendomain.NewKitchen(kitchenHandlerTestKitchenID) },
			body:       `{"order_id":"order-not-exist"}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := buildCompleteCookingHandler(tt.kitchen())

			req := httptest.NewRequest(http.MethodPost, "/kitchen/complete", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			wrap(h.HandleCompleteCooking)(rec, req)

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
