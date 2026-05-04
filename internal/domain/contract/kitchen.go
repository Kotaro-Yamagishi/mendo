package contract

// --- Kitchen BC の公開イベント ---

const (
	PublicEventTypeCookingCompleted = "public.cooking.completed"
)

type CookingCompletedPublic struct {
	OrderID   string `json:"order_id"`
	KitchenID string `json:"kitchen_id"`
}

func (e CookingCompletedPublic) GetEventType() string    { return PublicEventTypeCookingCompleted }
func (e CookingCompletedPublic) GetAggregateID() string  { return e.KitchenID }
func (e CookingCompletedPublic) GetCorrelationID() string { return "" }
