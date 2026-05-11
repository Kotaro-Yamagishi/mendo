package contract_test

import (
	"encoding/json"
	"testing"

	"mendo/internal/domain/contract"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SpecialOrderRejectedPublic_スキーマ(t *testing.T) {
	t.Parallel()

	event := contract.SpecialOrderRejectedPublic{
		SpecialOrderID: "so-1",
		Reason:         "材料切れ",
		SuggestedMenu:  "醤油ラーメン",
	}

	assert.Equal(t, contract.PublicEventTypeSpecialOrderRejected, event.GetEventType())
	assert.NotEmpty(t, event.Reason)
	assert.NotEmpty(t, event.SuggestedMenu)
}

func Test_SpecialOrderRejectedPublic_JSONラウンドトリップ(t *testing.T) {
	t.Parallel()

	original := contract.SpecialOrderRejectedPublic{
		SpecialOrderID: "so-1", Reason: "材料切れ", SuggestedMenu: "醤油ラーメン",
	}
	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored contract.SpecialOrderRejectedPublic
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, original, restored)
}

func Test_CookingDispatchedPublic_スキーマ(t *testing.T) {
	t.Parallel()

	event := contract.CookingDispatchedPublic{
		SpecialOrderID: "so-1",
		OrderID:        "order-1",
	}

	assert.Equal(t, contract.PublicEventTypeCookingDispatched, event.GetEventType())
	assert.NotEmpty(t, event.OrderID)
}

func Test_CookingDispatchedPublic_JSONラウンドトリップ(t *testing.T) {
	t.Parallel()

	original := contract.CookingDispatchedPublic{SpecialOrderID: "so-1", OrderID: "order-1"}
	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored contract.CookingDispatchedPublic
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, original, restored)
}
