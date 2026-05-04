package specialorder

import commondomain "mendo/internal/domain"

const (
	AggregateTypeSpecialOrder = "SpecialOrder"

	EventTypeSpecialOrderRequested = "special_order.requested"
	EventTypeSpecialOrderApproved  = "special_order.approved"
	EventTypeSpecialOrderRejected  = "special_order.rejected"
	EventTypeMenuResubmitted       = "special_order.menu_resubmitted"
	EventTypeCookingDispatched     = "special_order.cooking_dispatched"
)

type SpecialOrderRequested struct {
	commondomain.DomainEvent
	OrderID  string `json:"order_id"`
	MenuName string `json:"menu_name"`
}

func NewSpecialOrderRequested(id SpecialOrderID, orderID, menuName, correlationID string) SpecialOrderRequested {
	return SpecialOrderRequested{
		DomainEvent: commondomain.NewDomainEvent(string(id), AggregateTypeSpecialOrder, EventTypeSpecialOrderRequested, correlationID),
		OrderID:     orderID,
		MenuName:    menuName,
	}
}

func (e SpecialOrderRequested) GetEventType() string    { return e.EventType }
func (e SpecialOrderRequested) GetAggregateID() string  { return e.AggregateID }
func (e SpecialOrderRequested) GetCorrelationID() string { return e.CorrelationID }

type SpecialOrderApproved struct {
	commondomain.DomainEvent
}

func NewSpecialOrderApproved(id SpecialOrderID, correlationID string) SpecialOrderApproved {
	return SpecialOrderApproved{
		DomainEvent: commondomain.NewDomainEvent(string(id), AggregateTypeSpecialOrder, EventTypeSpecialOrderApproved, correlationID),
	}
}

func (e SpecialOrderApproved) GetEventType() string    { return e.EventType }
func (e SpecialOrderApproved) GetAggregateID() string  { return e.AggregateID }
func (e SpecialOrderApproved) GetCorrelationID() string { return e.CorrelationID }

type SpecialOrderRejected struct {
	commondomain.DomainEvent
	Reason        string `json:"reason"`
	SuggestedMenu string `json:"suggested_menu"`
}

func NewSpecialOrderRejected(id SpecialOrderID, reason, suggestedMenu, correlationID string) SpecialOrderRejected {
	return SpecialOrderRejected{
		DomainEvent:   commondomain.NewDomainEvent(string(id), AggregateTypeSpecialOrder, EventTypeSpecialOrderRejected, correlationID),
		Reason:        reason,
		SuggestedMenu: suggestedMenu,
	}
}

func (e SpecialOrderRejected) GetEventType() string    { return e.EventType }
func (e SpecialOrderRejected) GetAggregateID() string  { return e.AggregateID }
func (e SpecialOrderRejected) GetCorrelationID() string { return e.CorrelationID }

type MenuResubmitted struct {
	commondomain.DomainEvent
	NewMenuName string `json:"new_menu_name"`
}

func NewMenuResubmitted(id SpecialOrderID, newMenuName, correlationID string) MenuResubmitted {
	return MenuResubmitted{
		DomainEvent: commondomain.NewDomainEvent(string(id), AggregateTypeSpecialOrder, EventTypeMenuResubmitted, correlationID),
		NewMenuName: newMenuName,
	}
}

func (e MenuResubmitted) GetEventType() string    { return e.EventType }
func (e MenuResubmitted) GetAggregateID() string  { return e.AggregateID }
func (e MenuResubmitted) GetCorrelationID() string { return e.CorrelationID }

type CookingDispatched struct {
	commondomain.DomainEvent
	OrderID string `json:"order_id"`
}

func NewCookingDispatched(id SpecialOrderID, orderID, correlationID string) CookingDispatched {
	return CookingDispatched{
		DomainEvent: commondomain.NewDomainEvent(string(id), AggregateTypeSpecialOrder, EventTypeCookingDispatched, correlationID),
		OrderID:     orderID,
	}
}

func (e CookingDispatched) GetEventType() string    { return e.EventType }
func (e CookingDispatched) GetAggregateID() string  { return e.AggregateID }
func (e CookingDispatched) GetCorrelationID() string { return e.CorrelationID }
