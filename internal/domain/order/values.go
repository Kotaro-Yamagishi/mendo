package order

import "github.com/google/uuid"

// --- OrderID ---

type OrderID string

func NewOrderID() OrderID         { return OrderID(uuid.New().String()) }
func (id OrderID) String() string { return string(id) }

// --- SeatNumber ---

type SeatNumber int

// --- Status ---

const (
	StatusPending   = "pending"
	StatusConfirmed = "confirmed"
	StatusCanceled  = "canceled"
)
