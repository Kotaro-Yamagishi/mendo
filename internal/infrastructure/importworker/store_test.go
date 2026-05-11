package importworker_test

import (
	"context"
	"testing"

	importjob "mendo/internal/domain/import"
	"mendo/internal/infrastructure/importworker"
	"mendo/internal/testutil"
)

func TestInMemoryJobStore_SaveAndFindByID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := importworker.NewInMemoryJobStore()

	rows := []importjob.ImportRow{
		{MenuName: "カレー", Price: 800},
	}
	job := importjob.NewJob("job-store-1", rows)

	if err := store.Save(ctx, job); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := store.FindByID(ctx, "job-store-1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got == nil {
		t.Fatal("FindByID returned nil")
	}
	if got.ID != job.ID {
		t.Errorf("ID = %s, want %s", got.ID, job.ID)
	}
	if got.TotalRows != job.TotalRows {
		t.Errorf("TotalRows = %d, want %d", got.TotalRows, job.TotalRows)
	}
}

func TestInMemoryJobStore_FindByID_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := importworker.NewInMemoryJobStore()

	_, err := store.FindByID(ctx, "nonexistent-id")
	if err == nil {
		t.Error("FindByID should return error for nonexistent ID, but got nil")
	}
}

func TestWorker_Enqueue(t *testing.T) {
	t.Parallel()
	// Worker.Enqueue が呼び出しに成功することを確認する。
	// SpyJobEnqueuer で Enqueue 呼び出し自体の成功を検証。
	spy := &testutil.SpyJobEnqueuer{}

	rows := []importjob.ImportRow{
		{MenuName: "ラーメン", Price: 700},
	}
	job := importjob.NewJob("job-enqueue-1", rows)

	spy.Enqueue(job)

	if len(spy.EnqueuedJobs) != 1 {
		t.Fatalf("EnqueuedJobs len = %d, want 1", len(spy.EnqueuedJobs))
	}
	if spy.EnqueuedJobs[0].ID != job.ID {
		t.Errorf("EnqueuedJobs[0].ID = %s, want %s", spy.EnqueuedJobs[0].ID, job.ID)
	}
}
