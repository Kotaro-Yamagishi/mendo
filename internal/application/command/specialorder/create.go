package specialorder

import (
	"context"
	"log/slog"

	"mendo/internal/domain"
	"mendo/internal/domain/specialorder"
)

type CreateSpecialOrderUsecase struct {
	writer    specialorder.Writer
	publisher domain.EventPublisher
}

func NewCreateSpecialOrderUsecase(w specialorder.Writer, pub domain.EventPublisher) *CreateSpecialOrderUsecase {
	return &CreateSpecialOrderUsecase{writer: w, publisher: pub}
}

func (uc *CreateSpecialOrderUsecase) Execute(ctx context.Context, orderID, menuName string) (string, error) {
	id := specialorder.NewSpecialOrderID()
	so := specialorder.NewSpecialOrder(id, orderID, menuName)

	if err := uc.writer.Save(ctx, so); err != nil {
		return "", err
	}
	if err := uc.publisher.Publish(ctx, so.DomainEvents()...); err != nil {
		return "", err
	}

	slog.InfoContext(ctx, "special order created", slog.String("special_order_id", string(id)), slog.String("order_id", orderID), slog.String("menu_name", menuName))
	return string(id), nil
}
