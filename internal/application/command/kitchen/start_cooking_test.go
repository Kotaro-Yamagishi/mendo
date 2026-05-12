package kitchen_test

import (
	"context"
	"fmt"
	"testing"

	appkitchen "mendo/internal/application/command/kitchen"
	"mendo/internal/apperrors"
	"mendo/internal/domain/contract"
	kitchendomain "mendo/internal/domain/kitchen"
	"mendo/internal/domain/order"
	"mendo/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testKitchenID = kitchendomain.KitchenID("kitchen-1")

func newStartCookingUsecase(
	reader *testutil.StubKitchenReader,
	writer *testutil.SpyKitchenWriter,
	pub *testutil.SpyEventPublisher,
) *appkitchen.StartCookingUsecase {
	return appkitchen.NewStartCookingUsecase(reader, writer, pub, testKitchenID)
}

func Test_StartCooking_正常系(t *testing.T) {
	t.Parallel()

	k := kitchendomain.NewKitchen(testKitchenID)
	reader := &testutil.StubKitchenReader{Kitchen: k}
	writer := &testutil.SpyKitchenWriter{}
	pub := &testutil.SpyEventPublisher{}
	uc := newStartCookingUsecase(reader, writer, pub)

	event := contract.OrderConfirmedPublic{
		OrderID: "order-1",
		SeatNo:  3,
		Items: []contract.OrderConfirmedPublicItem{
			{MenuName: "醤油ラーメン", Toppings: []string{"ネギ"}, Hardness: "普通"},
		},
	}

	err := uc.HandleOrderConfirmedPublic(context.Background(), event)

	require.NoError(t, err)
	require.NotNil(t, writer.SavedKitchen, "調理タスク追加後に Save される")
	assert.Empty(t, pub.Published, "正常系ではイベントは Publish されない")
}

func Test_StartCooking_フル稼働時はCookingRejectedをPublish(t *testing.T) {
	t.Parallel()

	// MaxConcurrentTasks 分のタスクを積んだ Kitchen を用意する
	k := kitchendomain.NewKitchen(testKitchenID)
	for i := 0; i < kitchendomain.MaxConcurrentTasks; i++ {
		_ = k.AddCookingTask(
			order.OrderID(fmt.Sprintf("order-dummy-%d", i)),
			[]kitchendomain.CookingInstruction{{MenuName: "dummy"}},
		)
	}

	reader := &testutil.StubKitchenReader{Kitchen: k}
	writer := &testutil.SpyKitchenWriter{}
	pub := &testutil.SpyEventPublisher{}
	uc := newStartCookingUsecase(reader, writer, pub)

	event := contract.OrderConfirmedPublic{
		OrderID: "order-overflow",
		SeatNo:  1,
		Items: []contract.OrderConfirmedPublicItem{
			{MenuName: "つけ麺", Toppings: nil, Hardness: "硬め"},
		},
	}

	err := uc.HandleOrderConfirmedPublic(context.Background(), event)

	// フル稼働時は補償アクションに委ねるため、ユースケース自体はエラーを返さない
	require.NoError(t, err)
	assert.Nil(t, writer.SavedKitchen, "フル稼働時は Save されない")
	assert.NotEmpty(t, pub.Published, "CookingRejected が Publish される")
}

func Test_StartCooking_Kitchen見つからない(t *testing.T) {
	t.Parallel()

	reader := &testutil.StubKitchenReader{FindErr: apperrors.Infrastructure("kitchen not found", nil)}
	writer := &testutil.SpyKitchenWriter{}
	pub := &testutil.SpyEventPublisher{}
	uc := newStartCookingUsecase(reader, writer, pub)

	event := contract.OrderConfirmedPublic{
		OrderID: "order-1",
		SeatNo:  1,
		Items:   []contract.OrderConfirmedPublicItem{{MenuName: "塩ラーメン"}},
	}

	err := uc.HandleOrderConfirmedPublic(context.Background(), event)

	require.Error(t, err)
	assert.True(t, apperrors.IsCode(err, "INTERNAL_ERROR"), "インフラエラーは INTERNAL_ERROR コード")
	assert.Nil(t, writer.SavedKitchen)
}
