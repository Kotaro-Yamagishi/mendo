package kitchen

import (
	"mendo/internal/domain"
	"mendo/internal/domain/order"
)

const (
	AggregateTypeKitchen      = "Kitchen"
	EventTypeCookingCompleted = "cooking.completed"
	EventTypeCookingRejected  = "cooking.rejected"
)

// CookingCompleted は調理が完了した時のイベント。
type CookingCompleted struct {
	domain.DomainEvent
	OrderID order.OrderID `json:"order_id"`
}

func NewCookingCompleted(kitchenID KitchenID, orderID order.OrderID, correlationID string) CookingCompleted {
	return CookingCompleted{
		DomainEvent: domain.NewDomainEvent(string(kitchenID), AggregateTypeKitchen, EventTypeCookingCompleted, correlationID),
		OrderID:     orderID,
	}
}

func (e CookingCompleted) GetEventType() string     { return e.EventType }
func (e CookingCompleted) GetAggregateID() string   { return e.AggregateID }
func (e CookingCompleted) GetCorrelationID() string { return e.CorrelationID }

// CookingRejected は厨房がフル稼働で調理タスクを受け付けられなかった時のイベント。
type CookingRejected struct {
	domain.DomainEvent
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}

func NewCookingRejected(kitchenID KitchenID, orderID, reason, correlationID string) CookingRejected {
	return CookingRejected{
		DomainEvent: domain.NewDomainEvent(string(kitchenID), AggregateTypeKitchen, EventTypeCookingRejected, correlationID),
		OrderID:     orderID,
		Reason:      reason,
	}
}

func (e CookingRejected) GetEventType() string     { return e.EventType }
func (e CookingRejected) GetAggregateID() string   { return e.AggregateID }
func (e CookingRejected) GetCorrelationID() string { return e.CorrelationID }

type DomainEvent = domain.Event
