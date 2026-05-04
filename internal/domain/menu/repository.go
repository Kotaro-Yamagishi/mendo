package menu

import "context"

type Reader interface {
	FindByID(ctx context.Context, id MenuID) (*Menu, error)
	FindAll(ctx context.Context) ([]*Menu, error)
}

type Writer interface {
	Save(ctx context.Context, m *Menu) error
}
