package order_test

import (
	"context"
	"errors"
	"testing"

	apporder "mendo/internal/application/query/order"
	domorder "mendo/internal/domain/order"
	"mendo/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ListOrdersUsecase_Execute_正常系(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		stubRows  []domorder.OrderStateRow
		wantLen   int
		wantFirst *domorder.OrderStateRow // nil なら先頭要素の検証スキップ
	}{
		{
			name: "複数件",
			stubRows: []domorder.OrderStateRow{
				{OrderID: "order-1", SeatNo: 3, Status: "pending", ItemCount: 1},
				{OrderID: "order-2", SeatNo: 5, Status: "confirmed", ItemCount: 2},
			},
			wantLen:   2,
			wantFirst: &domorder.OrderStateRow{OrderID: "order-1", Status: "pending"},
		},
		{
			name:      "空",
			stubRows:  nil,
			wantLen:   0,
			wantFirst: nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			pr := &testutil.StubProjectionReader{Rows: tc.stubRows}
			uc := apporder.NewListOrdersUsecase(pr)

			rows, err := uc.Execute(context.Background())

			require.NoError(t, err)
			assert.Len(t, rows, tc.wantLen)
			if tc.wantFirst != nil {
				require.NotEmpty(t, rows)
				assert.Equal(t, tc.wantFirst.OrderID, rows[0].OrderID)
				assert.Equal(t, tc.wantFirst.Status, rows[0].Status)
			}
		})
	}
}

func Test_ListOrdersUsecase_Execute_エラー(t *testing.T) {
	t.Parallel()

	pr := &testutil.StubProjectionReader{AllErr: errors.New("db error")}
	uc := apporder.NewListOrdersUsecase(pr)

	_, err := uc.Execute(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "list orders")
}

func Test_ListOrdersUsecase_FindByID_正常系(t *testing.T) {
	t.Parallel()

	pr := &testutil.StubProjectionReader{
		Row: &domorder.OrderStateRow{OrderID: "order-1", SeatNo: 3, Status: "confirmed", ItemCount: 1},
	}
	uc := apporder.NewListOrdersUsecase(pr)

	row, err := uc.FindByID(context.Background(), "order-1")

	require.NoError(t, err)
	require.NotNil(t, row)
	assert.Equal(t, "confirmed", row.Status)
}

func Test_ListOrdersUsecase_FindByID_見つからない(t *testing.T) {
	t.Parallel()

	pr := &testutil.StubProjectionReader{FindErr: errors.New("not found")}
	uc := apporder.NewListOrdersUsecase(pr)

	_, err := uc.FindByID(context.Background(), "nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "find order")
}
