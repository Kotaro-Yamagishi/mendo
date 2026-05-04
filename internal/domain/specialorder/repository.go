package specialorder

import "context"

type Reader interface {
	FindByID(ctx context.Context, id SpecialOrderID) (*SpecialOrder, error)
}

type Writer interface {
	Save(ctx context.Context, so *SpecialOrder) error
}
