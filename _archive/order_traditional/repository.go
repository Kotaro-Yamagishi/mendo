package order

import "context"

type Reader interface {
	FindByID(ctx context.Context, id OrderID) (*Order, error)
	CountPending(ctx context.Context) (int, error)
}

type Writer interface {
	Save(ctx context.Context, o *Order) error
}
