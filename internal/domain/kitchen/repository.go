package kitchen

import "context"

type Reader interface {
	FindByID(ctx context.Context, id KitchenID) (*Kitchen, error)
}

type Writer interface {
	Save(ctx context.Context, k *Kitchen) error
}
