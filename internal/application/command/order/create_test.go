package order_test

import (
	"context"
	"testing"

	apporder "mendo/internal/application/command/order"
	"mendo/internal/apperrors"
	"mendo/internal/domain/order"
	"mendo/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CreateOrder_正常系(t *testing.T) {
	t.Parallel()

	es := &testutil.StubEventStore{}
	outbox := &testutil.SpyOutbox{}
	uc := apporder.NewCreateOrderUsecase(es, outbox, &testutil.SpyEventPublisher{})

	input := apporder.CreateOrderInput{
		SeatNo: 3,
		Items: []apporder.CreateOrderItemInput{
			{MenuID: "menu-1", Toppings: []string{"ネギ"}, Hardness: "ふつう"},
		},
	}

	orderID, err := uc.Execute(context.Background(), input)

	require.NoError(t, err)
	assert.NotEmpty(t, orderID)

	// EventStore に OrderCreated + ItemAdded が保存される
	require.Len(t, es.Saved, 2)
	_, isCreated := es.Saved[0].(order.OrderCreated)
	assert.True(t, isCreated)
	_, isAdded := es.Saved[1].(order.ItemAdded)
	assert.True(t, isAdded)

	// Outbox にも同数保存される
	assert.Len(t, outbox.Stored, 2)
}

func Test_CreateOrder_トッピング上限超過(t *testing.T) {
	t.Parallel()

	es := &testutil.StubEventStore{}
	outbox := &testutil.SpyOutbox{}
	uc := apporder.NewCreateOrderUsecase(es, outbox, &testutil.SpyEventPublisher{})

	input := apporder.CreateOrderInput{
		SeatNo: 1,
		Items: []apporder.CreateOrderItemInput{
			{MenuID: "menu-1", Toppings: []string{"a", "b", "c", "d"}, Hardness: "ふつう"},
		},
	}

	_, err := uc.Execute(context.Background(), input)

	require.Error(t, err)
	assert.True(t, apperrors.IsCode(err, order.ErrCodeTooManyToppings), "got: %v", err)
}
