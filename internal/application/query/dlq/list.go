package dlq

import (
	"context"

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
		return nil, err
	}
	return letters, nil
}
