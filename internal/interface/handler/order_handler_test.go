package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ordercommand "mendo/internal/application/command/order"
	"mendo/internal/domain"
	"mendo/internal/domain/menu"
	"mendo/internal/domain/order"
	"mendo/internal/interface/handler"
	"mendo/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newOrderHandler(es *testutil.StubEventStore, outbox *testutil.SpyOutbox) *handler.OrderHandler {
	pub := &testutil.SpyEventPublisher{}
	createUC := ordercommand.NewCreateOrderUsecase(es, outbox, pub)
	confirmUC := ordercommand.NewConfirmOrderESUsecase(es, outbox, pub)
	cancelUC := ordercommand.NewCancelOrderUsecase(es, outbox)
	return handler.NewOrderHandler(createUC, confirmUC, cancelUC, nil, nil)
}

func Test_OrderHandler_HandleCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "正常系",
			body:       `{"SeatNo":3,"Items":[{"MenuID":"menu-1","Toppings":["ネギ"],"Hardness":"普通"}]}`,
			wantStatus: http.StatusCreated,
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

			es := &testutil.StubEventStore{}
			outbox := &testutil.SpyOutbox{}
			h := newOrderHandler(es, outbox)

			req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			h.HandleCreate(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantStatus == http.StatusCreated {
				var resp map[string]interface{}
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
				data := resp["data"].(map[string]interface{})
				assert.NotEmpty(t, data["order_id"])
			}
		})
	}
}

func Test_OrderHandler_HandleConfirm(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		events     []domain.Event
		wantStatus int
	}{
		{
			name: "正常系",
			events: []domain.Event{
				order.NewOrderCreated("order-1", 3, ""),
				order.NewItemAdded("order-1", menu.MenuID("menu-1"), nil, "普通", ""),
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "業務エラー_confirmed状態から二重確定",
			events: []domain.Event{
				order.NewOrderCreated("order-1", 3, ""),
				order.NewItemAdded("order-1", menu.MenuID("menu-1"), nil, "普通", ""),
				order.NewOrderConfirmed("order-1", []order.ConfirmedItem{{MenuID: "menu-1"}}, 3, ""),
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			es := &testutil.StubEventStore{Events: tt.events}
			outbox := &testutil.SpyOutbox{}
			h := newOrderHandler(es, outbox)

			req := httptest.NewRequest(http.MethodPost, "/orders/order-1/confirm", nil)
			req.SetPathValue("id", "order-1")
			rec := httptest.NewRecorder()

			h.HandleConfirm(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantStatus == http.StatusOK {
				var resp map[string]interface{}
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
				data := resp["data"].(map[string]interface{})
				assert.Equal(t, "confirmed", data["status"])
			}
		})
	}
}

func Test_OrderHandler_HandleCancel_正常系(t *testing.T) {
	t.Parallel()

	es := &testutil.StubEventStore{
		Events: []domain.Event{
			order.NewOrderCreated("order-1", 3, ""),
			order.NewItemAdded("order-1", menu.MenuID("menu-1"), nil, "普通", ""),
			order.NewOrderConfirmed("order-1", []order.ConfirmedItem{{MenuID: "menu-1"}}, 3, ""),
		},
	}
	outbox := &testutil.SpyOutbox{}
	h := newOrderHandler(es, outbox)

	body := `{"reason":"客都合"}`
	req := httptest.NewRequest(http.MethodPost, "/orders/order-1/cancel", strings.NewReader(body))
	req.SetPathValue("id", "order-1")
	rec := httptest.NewRecorder()

	h.HandleCancel(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
