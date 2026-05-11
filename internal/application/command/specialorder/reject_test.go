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

func Test_RejectSpecialOrder_正常系(t *testing.T) {
	t.Parallel()

	so := sodomain.NewSpecialOrder("so-1", "order-1", "特製つけ麺")
	reader := &testutil.StubSpecialOrderReader{SpecialOrder: so}
	writer := &testutil.SpySpecialOrderWriter{}
	pub := &testutil.SpyEventPublisher{}
	uc := appso.NewRejectSpecialOrderUsecase(reader, writer, pub)

	err := uc.Execute(context.Background(), "so-1", "材料切れ", "醤油ラーメン")

	require.NoError(t, err)
	require.NotNil(t, writer.SavedSpecialOrder)
	assert.Equal(t, sodomain.StatusRejected, writer.SavedSpecialOrder.Status())
	assert.NotEmpty(t, pub.Published)
}
