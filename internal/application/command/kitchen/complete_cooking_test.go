package kitchen_test

import (
	"context"
	"testing"

	appkitchen "mendo/internal/application/command/kitchen"
	"mendo/internal/apperrors"
	kitchendomain "mendo/internal/domain/kitchen"
	"mendo/internal/domain/order"
	"mendo/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CompleteCooking_正常系(t *testing.T) {
	t.Parallel()

	k := kitchendomain.NewKitchen("kitchen-1")
	require.NoError(t, k.AddCookingTask(
		order.OrderID("order-1"),
		[]kitchendomain.CookingInstruction{{MenuName: "ラーメン"}},
	))

	reader := &testutil.StubKitchenReader{Kitchen: k}
	writer := &testutil.SpyKitchenWriter{}
	outbox := &testutil.SpyOutbox{}
	tx := &testutil.StubTransactionManager{}
	uc := appkitchen.NewCompleteCookingUsecase(reader, writer, outbox, tx, "kitchen-1")

	err := uc.Execute(context.Background(), order.OrderID("order-1"))

	require.NoError(t, err)
	require.NotNil(t, writer.SavedKitchen)
	assert.NotEmpty(t, outbox.Stored, "CookingCompleted イベントが Outbox に保存される")
}

func Test_CompleteCooking_異常系(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func() (*testutil.StubKitchenReader, order.OrderID)
		wantCode string
	}{
		{
			name: "厨房見つからない",
			setup: func() (*testutil.StubKitchenReader, order.OrderID) {
				return &testutil.StubKitchenReader{FindErr: apperrors.Infrastructure("kitchen not found", nil)},
					order.OrderID("order-1")
			},
			wantCode: "INTERNAL_ERROR",
		},
		{
			name: "タスク見つからない",
			setup: func() (*testutil.StubKitchenReader, order.OrderID) {
				k := kitchendomain.NewKitchen("kitchen-1")
				return &testutil.StubKitchenReader{Kitchen: k},
					order.OrderID("nonexistent")
			},
			wantCode: kitchendomain.ErrCodeTaskNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reader, orderID := tt.setup()
			writer := &testutil.SpyKitchenWriter{}
			outbox := &testutil.SpyOutbox{}
			tx := &testutil.StubTransactionManager{}
			uc := appkitchen.NewCompleteCookingUsecase(reader, writer, outbox, tx, "kitchen-1")

			err := uc.Execute(context.Background(), orderID)

			require.Error(t, err)
			assert.True(t, apperrors.IsCode(err, tt.wantCode), "got: %v", err)
		})
	}
}
