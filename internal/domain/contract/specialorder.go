package contract

// --- SpecialOrder BC の公開イベント ---

const (
	PublicEventTypeSpecialOrderRejected = "public.special_order.rejected"
	PublicEventTypeCookingDispatched    = "public.special_order.cooking_dispatched"
)

type SpecialOrderRejectedPublic struct {
	SpecialOrderID string `json:"special_order_id"`
	Reason         string `json:"reason"`
	SuggestedMenu  string `json:"suggested_menu"`
}

func (e SpecialOrderRejectedPublic) GetEventType() string    { return PublicEventTypeSpecialOrderRejected }
func (e SpecialOrderRejectedPublic) GetAggregateID() string   { return e.SpecialOrderID }
func (e SpecialOrderRejectedPublic) GetCorrelationID() string { return "" }

type CookingDispatchedPublic struct {
	SpecialOrderID string `json:"special_order_id"`
	OrderID        string `json:"order_id"`
}

func (e CookingDispatchedPublic) GetEventType() string    { return PublicEventTypeCookingDispatched }
func (e CookingDispatchedPublic) GetAggregateID() string   { return e.SpecialOrderID }
func (e CookingDispatchedPublic) GetCorrelationID() string { return "" }
