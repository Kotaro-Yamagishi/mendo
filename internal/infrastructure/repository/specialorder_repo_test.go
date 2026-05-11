package repository_test

import (
	"context"
	"testing"

	"mendo/internal/domain/specialorder"
	"mendo/internal/infrastructure/datasource"
	"mendo/internal/infrastructure/repository"
	"mendo/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SpecialOrderRepo_FindByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		stubRow      *datasource.SpecialOrderRow
		wantErr      bool
		wantStatus   specialorder.SpecialOrderStatus
		wantMenuName string
		wantSuggested string
	}{
		{
			name: "正常系",
			stubRow: &datasource.SpecialOrderRow{
				ID: "so-1", OrderID: "order-1", MenuName: "特製つけ麺",
				Status: 3, SuggestedMenu: "醤油ラーメン",
			},
			wantErr:       false,
			wantStatus:    specialorder.StatusRejected,
			wantMenuName:  "特製つけ麺",
			wantSuggested: "醤油ラーメン",
		},
		{
			name:    "見つからない",
			stubRow: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ds := &testutil.StubSpecialOrderDataSource{SpecialOrder: tt.stubRow}
			repo := repository.NewSpecialOrderRepository(ds)

			so, err := repo.FindByID(context.Background(), specialorder.SpecialOrderID("so-1"))

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, so.Status())
			assert.Equal(t, tt.wantMenuName, so.MenuName())
			assert.Equal(t, tt.wantSuggested, so.SuggestedMenu())
		})
	}
}

func Test_SpecialOrderRepo_Save(t *testing.T) {
	t.Parallel()

	ds := &testutil.StubSpecialOrderDataSource{}
	repo := repository.NewSpecialOrderRepository(ds)

	so := specialorder.NewSpecialOrder(specialorder.SpecialOrderID("so-2"), "order-2", "塩ラーメン")

	err := repo.Save(context.Background(), so)

	require.NoError(t, err)
	require.NotNil(t, ds.UpsertedRow)
	assert.Equal(t, "so-2", ds.UpsertedRow.ID)
	assert.Equal(t, "塩ラーメン", ds.UpsertedRow.MenuName)
	assert.Equal(t, int(specialorder.StatusPending), ds.UpsertedRow.Status)
}
