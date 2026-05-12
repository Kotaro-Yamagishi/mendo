package dlq

import (
	"context"

	"mendo/internal/apperrors"
	"mendo/internal/domain"
)

type RetryDLQUsecase struct {
	dlq       domain.DeadLetterQueue
	publisher domain.EventPublisher
}

func NewRetryDLQUsecase(dlq domain.DeadLetterQueue, pub domain.EventPublisher) *RetryDLQUsecase {
	return &RetryDLQUsecase{dlq: dlq, publisher: pub}
}

func (uc *RetryDLQUsecase) Execute(ctx context.Context, id string) error {
	letter, err := uc.dlq.FindByID(ctx, id)
	if err != nil {
		return apperrors.NotFound("dead_letter", id)
	}
	if err := uc.publisher.Publish(ctx, letter.Event); err != nil {
		return err
	}
	if err := uc.dlq.Remove(ctx, id); err != nil {
		return err
	}
	return nil
}
