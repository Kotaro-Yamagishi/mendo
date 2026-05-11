package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	menucommand "mendo/internal/application/command/menu"
	menudomain "mendo/internal/domain/menu"
	"mendo/internal/interface/handler"
	"mendo/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_MenuHandler_HandleSoldOut(t *testing.T) {
	t.Parallel()

	name, _ := menudomain.NewMenuName("醤油ラーメン")
	price, _ := menudomain.NewPrice(800)
	m := menudomain.NewMenu(menudomain.MenuID("menu-1"), name, price)

	tests := []struct {
		name         string
		reader       *testutil.StubMenuReader
		wantStatus   int
		wantContains string
	}{
		{
			name:         "正常系",
			reader:       &testutil.StubMenuReader{Menu: m},
			wantStatus:   http.StatusOK,
			wantContains: "sold_out",
		},
		{
			name:         "メニュー見つからない",
			reader:       &testutil.StubMenuReader{FindErr: errors.New("not found")},
			wantStatus:   http.StatusUnprocessableEntity,
			wantContains: "find menu",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			writer := &testutil.SpyMenuWriter{}
			uc := menucommand.NewSoldOutMenuUsecase(tt.reader, writer)
			h := handler.NewMenuHandler(uc)

			req := httptest.NewRequest(http.MethodPost, "/menus/menu-1/soldout", nil)
			req.SetPathValue("id", "menu-1")
			rec := httptest.NewRecorder()

			h.HandleSoldOut(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var body map[string]interface{}
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
			if tt.wantStatus == http.StatusOK {
				data := body["data"].(map[string]interface{})
				assert.Equal(t, tt.wantContains, data["status"])
			} else {
				assert.Contains(t, body["error"], tt.wantContains)
			}
		})
	}
}
