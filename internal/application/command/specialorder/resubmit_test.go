package specialorder_test

import (
	"context"
	"testing"

	sodomain "mendo/internal/domain/specialorder"
	"mendo/internal/testutil"

	appso "mendo/internal/application/command/specialorder"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ResubmitSpecialOrder_正常系(t *testing.T) {
	t.Parallel()

	so := sodomain.NewSpecialOrder("so-1", "order-1", "特製つけ麺")
	require.NoError(t, so.Reject("材料切れ", "醤油ラーメン"))

	reader := &testutil.StubSpecialOrderReader{SpecialOrder: so}
	writer := &testutil.SpySpecialOrderWriter{}
	pub := &testutil.SpyEventPublisher{}
	uc := appso.NewResubmitSpecialOrderUsecase(reader, writer, pub)

	err := uc.Execute(context.Background(), "so-1", "塩ラーメン")

	require.NoError(t, err)
	require.NotNil(t, writer.SavedSpecialOrder)
	assert.Equal(t, sodomain.StatusPending, writer.SavedSpecialOrder.Status())
	assert.Equal(t, "塩ラーメン", writer.SavedSpecialOrder.MenuName())
	assert.NotEmpty(t, pub.Published)
}

func Test_ResubmitSpecialOrder_rejected以外はエラー(t *testing.T) {
	t.Parallel()

	// pending 状態 → 再申請不可
	so := sodomain.NewSpecialOrder("so-1", "order-1", "特製つけ麺")
	reader := &testutil.StubSpecialOrderReader{SpecialOrder: so}
	writer := &testutil.SpySpecialOrderWriter{}
	pub := &testutil.SpyEventPublisher{}
	uc := appso.NewResubmitSpecialOrderUsecase(reader, writer, pub)

	err := uc.Execute(context.Background(), "so-1", "塩ラーメン")

	require.Error(t, err)
	assert.Nil(t, writer.SavedSpecialOrder)
}
