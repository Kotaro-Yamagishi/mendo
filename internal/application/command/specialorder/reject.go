package specialorder

import (
	"context"
	"fmt"

	"mendo/internal/domain"
	"mendo/internal/domain/specialorder"
)

type RejectSpecialOrderUsecase struct {
	reader    specialorder.Reader
	writer    specialorder.Writer
	publisher domain.EventPublisher
}

func NewRejectSpecialOrderUsecase(r specialorder.Reader, w specialorder.Writer, pub domain.EventPublisher) *RejectSpecialOrderUsecase {
	return &RejectSpecialOrderUsecase{reader: r, writer: w, publisher: pub}
}

func (uc *RejectSpecialOrderUsecase) Execute(ctx context.Context, id, reason, suggestedMenu string) error {
	so, err := uc.reader.FindByID(ctx, specialorder.SpecialOrderID(id))
	if err != nil {
		return fmt.Errorf("failed to find special order: %w", err)
	}
	if err := so.Reject(reason, suggestedMenu); err != nil {
		return fmt.Errorf("failed to reject: %w", err)
	}
	if err := uc.writer.Save(ctx, so); err != nil {
		return fmt.Errorf("failed to save special order: %w", err)
	}
	if err := uc.publisher.Publish(ctx, so.DomainEvents()...); err != nil {
		return fmt.Errorf("failed to publish events: %w", err)
	}
	return nil
}
