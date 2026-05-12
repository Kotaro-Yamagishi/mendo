package specialorder_test

import (
	"context"
	"errors"
	"testing"

	"mendo/internal/apperrors"
	sodomain "mendo/internal/domain/specialorder"
	"mendo/internal/testutil"

	appso "mendo/internal/application/command/specialorder"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ApproveSpecialOrder_正常系(t *testing.T) {
	t.Parallel()

	so := sodomain.NewSpecialOrder("so-1", "order-1", "特製つけ麺")
	reader := &testutil.StubSpecialOrderReader{SpecialOrder: so}
	writer := &testutil.SpySpecialOrderWriter{}
	pub := &testutil.SpyEventPublisher{}
	uc := appso.NewApproveSpecialOrderUsecase(reader, writer, pub)

	err := uc.Execute(context.Background(), "so-1")

	require.NoError(t, err)
	require.NotNil(t, writer.SavedSpecialOrder)
	assert.Equal(t, sodomain.StatusCooking, writer.SavedSpecialOrder.Status())
	assert.NotEmpty(t, pub.Published, "承認時にイベントが Publish される")
}

func Test_ApproveSpecialOrder_異常系(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(t *testing.T) *testutil.StubSpecialOrderReader
		wantCode string
	}{
		{
			name: "見つからない",
			setup: func(t *testing.T) *testutil.StubSpecialOrderReader {
				t.Helper()
				return &testutil.StubSpecialOrderReader{FindErr: errors.New("not found")}
			},
			wantCode: "NOT_FOUND",
		},
		{
			name: "業務ルール違反",
			setup: func(t *testing.T) *testutil.StubSpecialOrderReader {
				t.Helper()
				// rejected 状態 → 承認不可
				so := sodomain.NewSpecialOrder("so-1", "order-1", "特製つけ麺")
				require.NoError(t, so.Reject("材料切れ", "醤油ラーメン"))
				return &testutil.StubSpecialOrderReader{SpecialOrder: so}
			},
			wantCode: sodomain.ErrCodeNotPending,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reader := tt.setup(t)
			writer := &testutil.SpySpecialOrderWriter{}
			pub := &testutil.SpyEventPublisher{}
			uc := appso.NewApproveSpecialOrderUsecase(reader, writer, pub)

			err := uc.Execute(context.Background(), "so-1")

			require.Error(t, err)
			assert.True(t, apperrors.IsCode(err, tt.wantCode), "expected code %s, got: %s", tt.wantCode, err.Error())
		})
	}
}
