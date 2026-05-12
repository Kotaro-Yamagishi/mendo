package dlq_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"mendo/internal/apperrors"
	"mendo/internal/application/command/dlq"
	"mendo/internal/domain"
	"mendo/internal/testutil"
)

func newTestDeadLetter(id string) *domain.DeadLetter {
	return &domain.DeadLetter{
		ID:          id,
		Error:       "some error",
		FailCount:   1,
		LastFailAt:  time.Now(),
		HandlerName: "TestHandler",
	}
}

func TestRetryDLQUsecase_Execute_Success(t *testing.T) {
	letter := newTestDeadLetter("letter-1")
	stubDLQ := &testutil.StubDeadLetterQueue{
		SingleLetter: letter,
	}
	spyPub := &testutil.SpyEventPublisher{}

	uc := dlq.NewRetryDLQUsecase(stubDLQ, spyPub)
	err := uc.Execute(context.Background(), "letter-1")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	// Remove が呼ばれたこと
	if len(stubDLQ.RemovedIDs) != 1 || stubDLQ.RemovedIDs[0] != "letter-1" {
		t.Errorf("expected RemovedIDs=[letter-1], got: %v", stubDLQ.RemovedIDs)
	}
}

func Test_RetryDLQ_異常系(t *testing.T) {
	t.Parallel()

	findErr := apperrors.Infrastructure("db error", nil)
	publishErr := apperrors.Infrastructure("publish failed", nil)
	removeErr := apperrors.Infrastructure("remove failed", nil)

	tests := []struct {
		name           string
		stubDLQ        *testutil.StubDeadLetterQueue
		spyPub         *testutil.SpyEventPublisher
		letterID       string
		wantCode       string  // apperrors コードで検証
		wantCause      error   // Cause による検証（WrappedError の場合）
		wantPublished  int
		wantRemovedIDs int
	}{
		{
			name:           "FindByID失敗_NotFoundコード",
			stubDLQ:        &testutil.StubDeadLetterQueue{FindErr: findErr},
			spyPub:         &testutil.SpyEventPublisher{},
			letterID:       "letter-1",
			wantCode:       "NOT_FOUND",
			wantPublished:  0,
			wantRemovedIDs: 0,
		},
		{
			name:           "Publish失敗",
			stubDLQ:        &testutil.StubDeadLetterQueue{SingleLetter: newTestDeadLetter("letter-2")},
			spyPub:         &testutil.SpyEventPublisher{PublishErr: publishErr},
			letterID:       "letter-2",
			wantCode:       "INTERNAL_ERROR",
			wantCause:      publishErr,
			wantPublished:  0,
			wantRemovedIDs: 0,
		},
		{
			name:           "Remove失敗",
			stubDLQ:        &testutil.StubDeadLetterQueue{SingleLetter: newTestDeadLetter("letter-3"), RemoveErr: removeErr},
			spyPub:         &testutil.SpyEventPublisher{},
			letterID:       "letter-3",
			wantCode:       "INTERNAL_ERROR",
			wantCause:      removeErr,
			wantPublished:  1, // Publish は成功した後に Remove が失敗する
			wantRemovedIDs: 0,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			uc := dlq.NewRetryDLQUsecase(tc.stubDLQ, tc.spyPub)
			err := uc.Execute(context.Background(), tc.letterID)

			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !apperrors.IsCode(err, tc.wantCode) {
				t.Errorf("expected code %s, got: %v", tc.wantCode, err)
			}
			if tc.wantCause != nil && !errors.Is(err, tc.wantCause) {
				t.Errorf("expected cause %v, got: %v", tc.wantCause, err)
			}
			if len(tc.spyPub.Published) != tc.wantPublished {
				t.Errorf("expected %d published events, got: %d", tc.wantPublished, len(tc.spyPub.Published))
			}
			if len(tc.stubDLQ.RemovedIDs) != tc.wantRemovedIDs {
				t.Errorf("expected %d removed IDs, got: %v", tc.wantRemovedIDs, tc.stubDLQ.RemovedIDs)
			}
		})
	}
}
