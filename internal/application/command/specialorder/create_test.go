package specialorder_test

import (
	"context"
	"errors"
	"testing"

	appso "mendo/internal/application/command/specialorder"
	"mendo/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CreateSpecialOrder_正常系(t *testing.T) {
	t.Parallel()

	writer := &testutil.SpySpecialOrderWriter{}
	pub := &testutil.SpyEventPublisher{}
	uc := appso.NewCreateSpecialOrderUsecase(writer, pub)

	id, err := uc.Execute(context.Background(), "order-1", "特製つけ麺")

	require.NoError(t, err)
	assert.NotEmpty(t, id, "生成された SpecialOrderID が返される")
	require.NotNil(t, writer.SavedSpecialOrder, "Save が呼ばれる")
	assert.Equal(t, "order-1", writer.SavedSpecialOrder.OrderID())
	assert.Equal(t, "特製つけ麺", writer.SavedSpecialOrder.MenuName())
	assert.NotEmpty(t, pub.Published, "SpecialOrderRequested が Publish される")
}

func Test_CreateSpecialOrder_Save失敗(t *testing.T) {
	t.Parallel()

	writer := &testutil.SpySpecialOrderWriter{SaveErr: errors.New("db error")}
	pub := &testutil.SpyEventPublisher{}
	uc := appso.NewCreateSpecialOrderUsecase(writer, pub)

	id, err := uc.Execute(context.Background(), "order-1", "特製つけ麺")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
	assert.Empty(t, id)
	assert.Empty(t, pub.Published, "Save 失敗時は Publish されない")
}
