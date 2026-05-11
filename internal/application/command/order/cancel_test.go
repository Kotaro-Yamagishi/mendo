package order_test

import (
	"context"
	"errors"
	"testing"

	"mendo/internal/domain"
	"mendo/internal/domain/menu"
	"mendo/internal/domain/order"
	"mendo/internal/testutil"

	apporder "mendo/internal/application/command/order"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CancelOrder_正常系(t *testing.T) {
	t.Parallel()

	// Given: confirmed 状態の注文
	es := &testutil.StubEventStore{
		Events: []domain.Event{
			order.NewOrderCreated("order-1", 3, ""),
			order.NewItemAdded("order-1", menu.MenuID("menu-1"), nil, "普通", ""),
			order.NewOrderConfirmed("order-1", []order.ConfirmedItem{
				{MenuID: "menu-1", Toppings: nil, Hardness: "普通"},
			}, 3, ""),
		},
	}
	outbox := &testutil.SpyOutbox{}
	uc := apporder.NewCancelOrderUsecase(es, outbox)

	err := uc.Execute(context.Background(), "order-1", "客都合")

	require.NoError(t, err)
	require.Len(t, es.Saved, 1)
	cancelled, ok := es.Saved[0].(order.OrderCancelled)
	assert.True(t, ok)
	assert.Equal(t, "客都合", cancelled.Reason)
	require.Len(t, outbox.Stored, 1)
}

func Test_CancelOrder_異常系(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func() *testutil.StubEventStore
		wantErr string
	}{
		{
			name: "業務ルール違反_pending",
			setup: func() *testutil.StubEventStore {
				return &testutil.StubEventStore{
					Events: []domain.Event{
						order.NewOrderCreated("order-1", 3, ""),
						order.NewItemAdded("order-1", menu.MenuID("menu-1"), nil, "普通", ""),
					},
				}
			},
			wantErr: "",
		},
		{
			name: "ロード失敗",
			setup: func() *testutil.StubEventStore {
				return &testutil.StubEventStore{LoadErr: errors.New("not found")}
			},
			wantErr: "load events",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			es := tt.setup()
			outbox := &testutil.SpyOutbox{}
			uc := apporder.NewCancelOrderUsecase(es, outbox)

			err := uc.Execute(context.Background(), "order-1", "客都合")

			require.Error(t, err)
			if tt.wantErr != "" {
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}
