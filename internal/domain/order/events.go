package order

import (
	"mendo/internal/domain"
	"mendo/internal/domain/menu"
)

// イベント種別の定数。Subscribe 時にハードコーディングしないために使う。
const (
	AggregateTypeOrder      = "Order"
	EventTypeOrderCreated   = "order.created"
	EventTypeOrderConfirmed = "order.confirmed"
	EventTypeOrderCanceled  = "order.canceled"
	EventTypeItemAdded      = "order.item_added"
)

// --- OrderCreated ---

type OrderCreated struct {
	domain.DomainEvent
	SeatNo int `json:"seat_no"`
}

func NewOrderCreated(orderID string, seatNo int, correlationID string) OrderCreated {
	return OrderCreated{
		DomainEvent: domain.NewDomainEvent(orderID, AggregateTypeOrder, EventTypeOrderCreated, correlationID),
		SeatNo:      seatNo,
	}
}

func (e OrderCreated) GetEventType() string    { return e.EventType }
func (e OrderCreated) GetAggregateID() string  { return e.AggregateID }
func (e OrderCreated) GetCorrelationID() string { return e.CorrelationID }

// --- ItemAdded ---

type ItemAdded struct {
	domain.DomainEvent
	MenuID   menu.MenuID `json:"menu_id"`
	Toppings []string    `json:"toppings"`
	Hardness string      `json:"hardness"`
}

func NewItemAdded(orderID string, menuID menu.MenuID, toppings []string, hardness, correlationID string) ItemAdded {
	return ItemAdded{
		DomainEvent: domain.NewDomainEvent(orderID, AggregateTypeOrder, EventTypeItemAdded, correlationID),
		MenuID:      menuID,
		Toppings:    toppings,
		Hardness:    hardness,
	}
}

func (e ItemAdded) GetEventType() string    { return e.EventType }
func (e ItemAdded) GetAggregateID() string  { return e.AggregateID }
func (e ItemAdded) GetCorrelationID() string { return e.CorrelationID }

// --- OrderConfirmed ---

// ConfirmedItem は OrderConfirmed 内部イベントに載せる注文明細。
type ConfirmedItem struct {
	MenuID   string   `json:"menu_id"`
	Toppings []string `json:"toppings"`
	Hardness string   `json:"hardness"`
}

type OrderConfirmed struct {
	domain.DomainEvent
	Items  []ConfirmedItem `json:"items"`
	SeatNo int             `json:"seat_no"`
}

func NewOrderConfirmed(orderID string, items []ConfirmedItem, seatNo int, correlationID string) OrderConfirmed {
	return OrderConfirmed{
		DomainEvent: domain.NewDomainEvent(orderID, AggregateTypeOrder, EventTypeOrderConfirmed, correlationID),
		Items:       items,
		SeatNo:      seatNo,
	}
}

func (e OrderConfirmed) GetEventType() string    { return e.EventType }
func (e OrderConfirmed) GetAggregateID() string  { return e.AggregateID }
func (e OrderConfirmed) GetCorrelationID() string { return e.CorrelationID }

// --- OrderCancelled ---

type OrderCancelled struct {
	domain.DomainEvent
	Reason string `json:"reason"`
}

func NewOrderCancelled(orderID, reason, correlationID string) OrderCancelled {
	return OrderCancelled{
		DomainEvent: domain.NewDomainEvent(orderID, AggregateTypeOrder, EventTypeOrderCanceled, correlationID),
		Reason:      reason,
	}
}

func (e OrderCancelled) GetEventType() string    { return e.EventType }
func (e OrderCancelled) GetAggregateID() string  { return e.AggregateID }
func (e OrderCancelled) GetCorrelationID() string { return e.CorrelationID }
