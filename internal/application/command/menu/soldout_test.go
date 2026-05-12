package menu_test

import (
	"context"
	"testing"

	appmenu "mendo/internal/application/command/menu"
	"mendo/internal/apperrors"
	"mendo/internal/domain/menu"
	"mendo/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SoldOutMenu_正常系(t *testing.T) {
	t.Parallel()

	name, _ := menu.NewMenuName("醤油ラーメン")
	price, _ := menu.NewPrice(800)
	m := menu.NewMenu(menu.MenuID("menu-1"), name, price)

	reader := &testutil.StubMenuReader{Menu: m}
	writer := &testutil.SpyMenuWriter{}
	uc := appmenu.NewSoldOutMenuUsecase(reader, writer)

	err := uc.Execute(context.Background(), menu.MenuID("menu-1"))

	require.NoError(t, err)
	require.NotNil(t, writer.SavedMenu)
	assert.False(t, writer.SavedMenu.IsAvailable(), "SoldOut 後は品切れ状態")
}

func Test_SoldOutMenu_異常系(t *testing.T) {
	t.Parallel()

	name, _ := menu.NewMenuName("醤油ラーメン")
	price, _ := menu.NewPrice(800)
	m := menu.NewMenu(menu.MenuID("menu-1"), name, price)

	tests := []struct {
		name     string
		setup    func() (*testutil.StubMenuReader, *testutil.SpyMenuWriter)
		wantCode string
	}{
		{
			name: "メニュー見つからない",
			setup: func() (*testutil.StubMenuReader, *testutil.SpyMenuWriter) {
				return &testutil.StubMenuReader{FindErr: apperrors.Infrastructure("menu not found", nil)},
					&testutil.SpyMenuWriter{}
			},
			wantCode: "INTERNAL_ERROR",
		},
		{
			name: "Save失敗",
			setup: func() (*testutil.StubMenuReader, *testutil.SpyMenuWriter) {
				return &testutil.StubMenuReader{Menu: m},
					&testutil.SpyMenuWriter{SaveErr: apperrors.Infrastructure("save failed", nil)}
			},
			wantCode: "INTERNAL_ERROR",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reader, writer := tt.setup()
			uc := appmenu.NewSoldOutMenuUsecase(reader, writer)

			err := uc.Execute(context.Background(), menu.MenuID("menu-1"))

			require.Error(t, err)
			assert.True(t, apperrors.IsCode(err, tt.wantCode), "got: %v", err)
		})
	}
}
