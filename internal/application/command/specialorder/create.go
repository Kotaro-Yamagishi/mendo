package specialorder

import (
	"context"
	"fmt"

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
		return "", fmt.Errorf("failed to save special order: %w", err)
	}
	if err := uc.publisher.Publish(ctx, so.DomainEvents()...); err != nil {
		return "", fmt.Errorf("failed to publish events: %w", err)
	}
	return string(id), nil
}
