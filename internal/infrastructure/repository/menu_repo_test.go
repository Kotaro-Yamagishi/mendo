package repository_test

import (
	"context"
	"errors"
	"testing"

	"mendo/internal/domain/menu"
	"mendo/internal/infrastructure/datasource"
	"mendo/internal/infrastructure/repository"
	"mendo/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_MenuRepo_FindByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		stubMenu *datasource.MenuRow
		wantErr  bool
		wantName string
	}{
		{
			name: "正常系",
			stubMenu: &datasource.MenuRow{
				MenuID: "menu-1", Name: "醤油ラーメン", Price: 800, Available: true,
			},
			wantErr:  false,
			wantName: "醤油ラーメン",
		},
		{
			name:     "見つからない",
			stubMenu: nil,
			wantErr:  true,
			wantName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ds := &testutil.StubMenuDataSource{Menu: tt.stubMenu}
			repo := repository.NewMenuRepository(ds)

			m, err := repo.FindByID(context.Background(), menu.MenuID("menu-1"))

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, m.Name().String())
			assert.Equal(t, 800, m.Price().Yen())
			assert.True(t, m.IsAvailable())
		})
	}
}

func Test_MenuRepo_Save(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		insertErr error
		wantErr   bool
	}{
		{
			name:      "正常系",
			insertErr: nil,
			wantErr:   false,
		},
		{
			name:      "DBエラー",
			insertErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ds := &testutil.StubMenuDataSource{InsertErr: tt.insertErr}
			repo := repository.NewMenuRepository(ds)

			name, _ := menu.NewMenuName("味噌ラーメン")
			price, _ := menu.NewPrice(900)
			m := menu.NewMenu(menu.MenuID("menu-2"), name, price)

			err := repo.Save(context.Background(), m)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, ds.InsertedMenu)
			assert.Equal(t, "menu-2", ds.InsertedMenu.MenuID)
			assert.Equal(t, "味噌ラーメン", ds.InsertedMenu.Name)
			assert.Equal(t, 900, ds.InsertedMenu.Price)
			assert.True(t, ds.InsertedMenu.Available)
		})
	}
}
