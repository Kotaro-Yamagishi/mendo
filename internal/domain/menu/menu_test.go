package menu_test

import (
	"testing"

	"mendo/internal/domain/menu"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// NewMenu — 集約の生成
// =============================================================================

func Test_NewMenu(t *testing.T) {
	t.Parallel()

	// Given: 有効な値オブジェクト
	name := newTestMenuName(t, "醤油ラーメン")
	price := newTestPrice(t, 800)
	id := menu.MenuID("menu-001")

	// When: Menu を生成する
	m := menu.NewMenu(id, name, price)

	// Then: 各フィールドが正しく設定され、Available 状態で作成される
	assert.Equal(t, "menu-001", m.ID().String())
	assert.Equal(t, "醤油ラーメン", m.Name().String())
	assert.Equal(t, 800, m.Price().Yen())
	assert.True(t, m.IsAvailable(), "新規作成時は Available であるべき")
}

// =============================================================================
// ReconstructMenu — DB からの復元
// =============================================================================

func Test_ReconstructMenu(t *testing.T) {
	t.Parallel()

	t.Run("Available状態で復元", func(t *testing.T) {
		t.Parallel()
		name := newTestMenuName(t, "味噌ラーメン")
		price := newTestPrice(t, 900)

		m := menu.ReconstructMenu(menu.MenuID("menu-002"), name, price, true)

		assert.True(t, m.IsAvailable())
	})

	t.Run("SoldOut状態で復元", func(t *testing.T) {
		t.Parallel()
		name := newTestMenuName(t, "限定ラーメン")
		price := newTestPrice(t, 1200)

		m := menu.ReconstructMenu(menu.MenuID("menu-003"), name, price, false)

		assert.False(t, m.IsAvailable())
	})
}

// =============================================================================
// SoldOut — 品切れコマンド
// =============================================================================

func Test_Menu_SoldOut(t *testing.T) {
	t.Parallel()

	// Given: Available 状態のメニュー
	m := newTestMenu(t)
	require.True(t, m.IsAvailable(), "テストの前提: Available 状態")

	// When: 品切れにする
	m.SoldOut()

	// Then: IsAvailable が false になる
	assert.False(t, m.IsAvailable())
}

// =============================================================================
// DomainEvents — Menu はイベントを発行しない
// =============================================================================

func Test_Menu_DomainEvents_is_nil(t *testing.T) {
	t.Parallel()

	m := newTestMenu(t)

	assert.Nil(t, m.DomainEvents(), "Menu 集約はイベントを発行しない設計")
}

// =============================================================================
// テストヘルパー
// =============================================================================

// newTestMenuName は有効な MenuName を生成する。
// テストの前提条件構築用。生成自体のテストではないため、失敗したら即停止する。
func newTestMenuName(t *testing.T, s string) menu.MenuName {
	t.Helper()
	name, err := menu.NewMenuName(s)
	require.NoError(t, err)
	return name
}

// newTestPrice は有効な Price を生成する。
func newTestPrice(t *testing.T, yen int) menu.Price {
	t.Helper()
	price, err := menu.NewPrice(yen)
	require.NoError(t, err)
	return price
}

// newTestMenu は標準的なテスト用 Menu を生成する。
// Available 状態で返す。
func newTestMenu(t *testing.T) *menu.Menu {
	t.Helper()
	name := newTestMenuName(t, "醤油ラーメン")
	price := newTestPrice(t, 800)
	return menu.NewMenu(menu.MenuID("menu-001"), name, price)
}
