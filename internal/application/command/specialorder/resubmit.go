package specialorder

import (
	"context"

	"mendo/internal/apperrors"
	"mendo/internal/domain"
	"mendo/internal/domain/specialorder"
)

type ResubmitSpecialOrderUsecase struct {
	reader    specialorder.Reader
	writer    specialorder.Writer
	publisher domain.EventPublisher
}

func NewResubmitSpecialOrderUsecase(r specialorder.Reader, w specialorder.Writer, pub domain.EventPublisher) *ResubmitSpecialOrderUsecase {
	return &ResubmitSpecialOrderUsecase{reader: r, writer: w, publisher: pub}
}

func (uc *ResubmitSpecialOrderUsecase) Execute(ctx context.Context, id, newMenuName string) error {
	so, err := uc.reader.FindByID(ctx, specialorder.SpecialOrderID(id))
	if err != nil {
		return apperrors.NotFound("special_order", id)
	}
	if err := so.ResubmitWithMenu(newMenuName); err != nil {
		return err
	}
	if err := uc.writer.Save(ctx, so); err != nil {
		return err
	}
	if err := uc.publisher.Publish(ctx, so.DomainEvents()...); err != nil {
		return err
	}
	return nil
}
