package dlq

import (
	"context"
	"fmt"

	"mendo/internal/domain"
)

type ListDLQHandler struct {
	dlq domain.DeadLetterQueue
}

func NewListDLQHandler(dlq domain.DeadLetterQueue) *ListDLQHandler {
	return &ListDLQHandler{dlq: dlq}
}

func (h *ListDLQHandler) Handle(ctx context.Context) ([]domain.DeadLetter, error) {
	letters, err := h.dlq.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list dead letters: %w", err)
	}
	return letters, nil
}
