package order

import (
	"mendo/internal/domain"
)

// OrderConfirmed は注文が確定された時のイベント。
type OrderConfirmed struct {
	domain.DomainEvent
	Items  []OrderItem `json:"items"`
	SeatNo SeatNumber  `json:"seat_no"`
}

func NewOrderConfirmed(orderID OrderID, items []OrderItem, seatNo SeatNumber, correlationID string) OrderConfirmed {
	return OrderConfirmed{
		DomainEvent: domain.NewDomainEvent(string(orderID), "Order", "order.confirmed", correlationID),
		Items:       items,
		SeatNo:      seatNo,
	}
}

func (e OrderConfirmed) GetEventType() string    { return e.EventType }
func (e OrderConfirmed) GetAggregateID() string  { return e.AggregateID }
func (e OrderConfirmed) GetCorrelationID() string { return e.CorrelationID }

// OrderCancelled は注文がキャンセルされた時のイベント。
type OrderCancelled struct {
	domain.DomainEvent
	Reason string `json:"reason"`
}

func NewOrderCancelled(orderID OrderID, reason, correlationID string) OrderCancelled {
	return OrderCancelled{
		DomainEvent: domain.NewDomainEvent(string(orderID), "Order", "order.cancelled", correlationID),
		Reason:      reason,
	}
}

func (e OrderCancelled) GetEventType() string    { return e.EventType }
func (e OrderCancelled) GetAggregateID() string  { return e.AggregateID }
func (e OrderCancelled) GetCorrelationID() string { return e.CorrelationID }

// --- 後方互換: 既存コードが DomainEvent interface を参照している場合 ---
// 既存の order.DomainEvent を domain.Event に統一するため、型エイリアスを定義
type DomainEvent = domain.Event
