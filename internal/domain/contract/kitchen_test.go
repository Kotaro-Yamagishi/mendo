package contract_test

import (
	"encoding/json"
	"testing"

	"mendo/internal/domain/contract"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CookingCompletedPublic_スキーマ(t *testing.T) {
	t.Parallel()

	event := contract.CookingCompletedPublic{
		OrderID:   "order-1",
		KitchenID: "kitchen-1",
	}

	assert.Equal(t, contract.PublicEventTypeCookingCompleted, event.GetEventType())
	assert.Equal(t, "kitchen-1", event.GetAggregateID())
	assert.NotEmpty(t, event.OrderID)
}

func Test_CookingCompletedPublic_JSONラウンドトリップ(t *testing.T) {
	t.Parallel()

	original := contract.CookingCompletedPublic{OrderID: "order-1", KitchenID: "kitchen-1"}
	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored contract.CookingCompletedPublic
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, original, restored)
}
