package importstatus_test

import (
	"context"
	"errors"
	"testing"

	"mendo/internal/application/query/importstatus"
	importjob "mendo/internal/domain/import"
	"mendo/internal/testutil"
)

func TestImportStatusHandler_Handle_Success(t *testing.T) {
	t.Parallel()
	job := importjob.NewJob("job-abc", []importjob.ImportRow{
		{MenuName: "カレー", Price: 800},
		{MenuName: "ラーメン", Price: 700},
	})
	job.Start()
	job.RecordSuccess()

	stubReader := &testutil.StubJobReader{Job: job}
	h := importstatus.NewImportStatusHandler(stubReader)

	resp, err := h.Handle(context.Background(), "job-abc")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.ID != "job-abc" {
		t.Errorf("expected ID=job-abc, got: %s", resp.ID)
	}
	if resp.Status != string(importjob.StatusProcessing) {
		t.Errorf("expected status=processing, got: %s", resp.Status)
	}
	if resp.TotalRows != 2 {
		t.Errorf("expected TotalRows=2, got: %d", resp.TotalRows)
	}
	if resp.ProcessedRows != 1 {
		t.Errorf("expected ProcessedRows=1, got: %d", resp.ProcessedRows)
	}
	if resp.SucceededRows != 1 {
		t.Errorf("expected SucceededRows=1, got: %d", resp.SucceededRows)
	}
	if resp.FailedRows != 0 {
		t.Errorf("expected FailedRows=0, got: %d", resp.FailedRows)
	}
	if resp.Progress != 50 {
		t.Errorf("expected Progress=50, got: %d", resp.Progress)
	}
}

func TestImportStatusHandler_Handle_NotFound(t *testing.T) {
	t.Parallel()
	findErr := errors.New("job not found")
	stubReader := &testutil.StubJobReader{FindErr: findErr}
	h := importstatus.NewImportStatusHandler(stubReader)

	_, err := h.Handle(context.Background(), "unknown-job")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, findErr) {
		t.Errorf("expected wrapped find error, got: %v", err)
	}
}
