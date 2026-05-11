package menu_test

import (
	"testing"

	"mendo/internal/domain/menu"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// MenuName
// =============================================================================

func Test_NewMenuName_正常系(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"1文字_境界値最小", "あ", "あ"},
		{"通常のメニュー名", "醤油ラーメン", "醤油ラーメン"},
		{"英数字混在", "Ramen #1", "Ramen #1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := menu.NewMenuName(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.String())
		})
	}
}

func Test_NewMenuName_異常系(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{"空文字_境界値", "", "空にできません"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := menu.NewMenuName(tt.input)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

// =============================================================================
// Price
// =============================================================================

func Test_NewPrice_正常系(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input int
		want  int
	}{
		{"0円_境界値最小", 0, 0},
		{"通常価格", 800, 800},
		{"高額", 10000, 10000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := menu.NewPrice(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Yen())
		})
	}
}

func Test_NewPrice_異常系(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   int
		wantErr string
	}{
		{"負数_境界値", -1, "価格は0以上"},
		{"大きな負数", -500, "価格は0以上"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := menu.NewPrice(tt.input)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

// =============================================================================
// MenuID
// =============================================================================

func Test_MenuID_String(t *testing.T) {
	t.Parallel()
	id := menu.MenuID("menu-001")
	assert.Equal(t, "menu-001", id.String())
}
