package dlq_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"mendo/internal/application/query/dlq"
	"mendo/internal/domain"
	"mendo/internal/testutil"
)

func TestListDLQHandler_Handle_Success(t *testing.T) {
	t.Parallel()
	letters := []domain.DeadLetter{
		{ID: "l-1", Error: "err1", FailCount: 1, LastFailAt: time.Now(), HandlerName: "HandlerA"},
		{ID: "l-2", Error: "err2", FailCount: 2, LastFailAt: time.Now(), HandlerName: "HandlerB"},
	}
	stubDLQ := &testutil.StubDeadLetterQueue{Letters: letters}
	h := dlq.NewListDLQHandler(stubDLQ)

	got, err := h.Handle(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 letters, got: %d", len(got))
	}
	if got[0].ID != "l-1" || got[1].ID != "l-2" {
		t.Errorf("unexpected letters: %v", got)
	}
}

func TestListDLQHandler_Handle_EmptyList(t *testing.T) {
	t.Parallel()
	stubDLQ := &testutil.StubDeadLetterQueue{Letters: []domain.DeadLetter{}}
	h := dlq.NewListDLQHandler(stubDLQ)

	got, err := h.Handle(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty list, got: %d items", len(got))
	}
}

func TestListDLQHandler_Handle_Error(t *testing.T) {
	t.Parallel()
	listErr := errors.New("storage unavailable")
	stubDLQ := &testutil.StubDeadLetterQueue{ListErr: listErr}
	h := dlq.NewListDLQHandler(stubDLQ)

	_, err := h.Handle(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, listErr) {
		t.Errorf("expected wrapped list error, got: %v", err)
	}
}
