package dlq

import (
	"context"
	"fmt"

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
		return fmt.Errorf("failed to find dead letter: %w", err)
	}
	if err := uc.publisher.Publish(ctx, letter.Event); err != nil {
		return fmt.Errorf("failed to retry event: %w", err)
	}
	if err := uc.dlq.Remove(ctx, id); err != nil {
		return fmt.Errorf("failed to remove dead letter: %w", err)
	}
	return nil
}
