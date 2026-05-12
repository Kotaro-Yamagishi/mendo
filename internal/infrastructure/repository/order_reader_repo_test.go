package repository_test

import (
	"context"
	"errors"
	"testing"

	"mendo/internal/infrastructure/datasource"
	"mendo/internal/infrastructure/repository"
	"mendo/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// CountPending
// =============================================================================

func Test_OrderReaderRepo_CountPending_正常系(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		rows      []datasource.OrderProjectionRow
		wantCount int
	}{
		{
			name: "pending が混在する場合は pending のみカウント",
			rows: []datasource.OrderProjectionRow{
				{OrderID: "order-1", Status: "pending"},
				{OrderID: "order-2", Status: "confirmed"},
				{OrderID: "order-3", Status: "pending"},
			},
			wantCount: 2,
		},
		{
			name: "全て確定済みの場合は 0",
			rows: []datasource.OrderProjectionRow{
				{OrderID: "order-1", Status: "confirmed"},
				{OrderID: "order-2", Status: "canceled"},
			},
			wantCount: 0,
		},
		{
			name:      "件数なしの場合は 0",
			rows:      nil,
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ds := &testutil.StubOrderProjectionDataSource{Rows: tc.rows}
			repo := repository.NewOrderReaderRepository(ds)

			count, err := repo.CountPending(context.Background())

			require.NoError(t, err)
			assert.Equal(t, tc.wantCount, count)
		})
	}
}

func Test_OrderReaderRepo_CountPending_DataSourceエラー(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubOrderProjectionDataSource{AllErr: errors.New("db error")}
	repo := repository.NewOrderReaderRepository(ds)

	_, err := repo.CountPending(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "注文一覧の取得に失敗")
}
