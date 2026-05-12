package order_test

import (
	"context"
	"testing"

	apporder "mendo/internal/application/command/order"
	"mendo/internal/apperrors"
	"mendo/internal/domain"
	"mendo/internal/domain/menu"
	"mendo/internal/domain/order"
	"mendo/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ConfirmOrderES_正常系(t *testing.T) {
	t.Parallel()

	es := &testutil.StubEventStore{
		Events: []domain.Event{
			order.NewOrderCreated("order-1", 3, ""),
			order.NewItemAdded("order-1", menu.MenuID("menu-1"), []string{"ネギ"}, "普通", ""),
		},
	}
	outbox := &testutil.SpyOutbox{}
	uc := apporder.NewConfirmOrderESUsecase(es, outbox, &testutil.SpyEventPublisher{})

	err := uc.Execute(context.Background(), "order-1")

	require.NoError(t, err)
	require.Len(t, es.Saved, 1)
	_, ok := es.Saved[0].(order.OrderConfirmed)
	assert.True(t, ok)
	require.Len(t, outbox.Stored, 1)
}

func Test_ConfirmOrderES_異常系(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func() *testutil.StubEventStore
		wantCode string
	}{
		{
			name: "EventStoreロード失敗",
			setup: func() *testutil.StubEventStore {
				return &testutil.StubEventStore{LoadErr: apperrors.Infrastructure("load failed", nil)}
			},
			wantCode: "INTERNAL_ERROR",
		},
		{
			name: "業務ルール違反",
			setup: func() *testutil.StubEventStore {
				return &testutil.StubEventStore{
					Events: []domain.Event{
						order.NewOrderCreated("order-1", 3, ""),
						order.NewItemAdded("order-1", menu.MenuID("menu-1"), nil, "普通", ""),
						order.NewOrderConfirmed("order-1", []order.ConfirmedItem{
							{MenuID: "menu-1", Toppings: nil, Hardness: "普通"},
						}, 3, ""),
					},
				}
			},
			wantCode: order.ErrCodeNotPending,
		},
		{
			name: "EventStoreSave失敗",
			setup: func() *testutil.StubEventStore {
				return &testutil.StubEventStore{
					Events: []domain.Event{
						order.NewOrderCreated("order-1", 3, ""),
						order.NewItemAdded("order-1", menu.MenuID("menu-1"), nil, "普通", ""),
					},
					SaveErr: apperrors.Infrastructure("save failed", nil),
				}
			},
			wantCode: "INTERNAL_ERROR",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			es := tt.setup()
			outbox := &testutil.SpyOutbox{}
			uc := apporder.NewConfirmOrderESUsecase(es, outbox, &testutil.SpyEventPublisher{})

			err := uc.Execute(context.Background(), "order-1")

			require.Error(t, err)
			assert.True(t, apperrors.IsCode(err, tt.wantCode), "got: %v", err)
			assert.Empty(t, outbox.Stored)
		})
	}
}
