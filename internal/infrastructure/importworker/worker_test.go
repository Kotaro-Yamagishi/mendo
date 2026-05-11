package importworker_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	importjob "mendo/internal/domain/import"
	menudomain "mendo/internal/domain/menu"
	"mendo/internal/infrastructure/importworker"
	"mendo/internal/testutil"
)

// enqueueAndWait はジョブをエンキューして完了を待つ。
// Job の Status/SucceededRows 等のフィールドは直接読まず、安全なメソッド経由でアクセスする。
func enqueueAndWait(t *testing.T, w *importworker.Worker, job *importjob.Job) {
	t.Helper()
	w.Enqueue(job)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		st := job.GetStatus()
		if st == importjob.StatusCompleted || st == importjob.StatusFailed {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("job did not complete in time: %s", job.Summary())
}

// uniqueName は i を使って重複しないメニュー名を生成する。
func uniqueName(i int) string {
	return fmt.Sprintf("メニュー%03d", i)
}

// TestWorker_AllSuccess は全行成功した場合に Job.Status が completed になることを確認する。
func TestWorker_AllSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobStore := importworker.NewInMemoryJobStore()
	menuReader := &testutil.StubMenuReader{Menus: nil}
	menuWriter := &testutil.SpyMenuWriter{}

	w := importworker.NewWorker(jobStore, menuReader, menuWriter)
	w.Start(ctx)

	rows := []importjob.ImportRow{
		{MenuName: "カレー", Price: 800},
		{MenuName: "ラーメン", Price: 700},
	}
	job := importjob.NewJob("job-1", rows)

	enqueueAndWait(t, w, job)

	if got := job.GetStatus(); got != importjob.StatusCompleted {
		t.Errorf("Status = %s, want %s", got, importjob.StatusCompleted)
	}
	if got := job.GetSucceededRows(); got != 2 {
		t.Errorf("SucceededRows = %d, want 2", got)
	}
	if got := job.GetFailedRows(); got != 0 {
		t.Errorf("FailedRows = %d, want 0", got)
	}
}

// TestWorker_DuplicateMenuName は重複メニュー名の行が RecordFailure されることを確認する。
func TestWorker_DuplicateMenuName(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobStore := importworker.NewInMemoryJobStore()

	// 既存メニューに "カレー" が存在する
	name, _ := menudomain.NewMenuName("カレー")
	price, _ := menudomain.NewPrice(800)
	existingMenu := menudomain.NewMenu("menu-existing", name, price)
	menuReader := &testutil.StubMenuReader{Menus: []*menudomain.Menu{existingMenu}}
	menuWriter := &testutil.SpyMenuWriter{}

	w := importworker.NewWorker(jobStore, menuReader, menuWriter)
	w.Start(ctx)

	rows := []importjob.ImportRow{
		{MenuName: "カレー", Price: 900},  // 重複 → 失敗
		{MenuName: "うどん", Price: 500}, // 新規 → 成功
	}
	job := importjob.NewJob("job-2", rows)

	enqueueAndWait(t, w, job)

	if got := job.GetFailedRows(); got != 1 {
		t.Errorf("FailedRows = %d, want 1", got)
	}
	if got := job.GetSucceededRows(); got != 1 {
		t.Errorf("SucceededRows = %d, want 1", got)
	}
	errs := job.GetErrors()
	if len(errs) != 1 {
		t.Fatalf("Errors len = %d, want 1", len(errs))
	}
	if errs[0].Name != "カレー" {
		t.Errorf("Errors[0].Name = %q, want %q", errs[0].Name, "カレー")
	}
}

// TestWorker_ChunkSplit はチャンクサイズを超える行数が正しく処理されることを確認する。
// chunkSize=10 なので 25 行は 3 チャンク（10+10+5）に分割される。
func TestWorker_ChunkSplit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobStore := importworker.NewInMemoryJobStore()
	menuReader := &testutil.StubMenuReader{Menus: nil}
	menuWriter := &testutil.SpyMenuWriter{}

	w := importworker.NewWorker(jobStore, menuReader, menuWriter)
	w.Start(ctx)

	rows := make([]importjob.ImportRow, 25)
	for i := range rows {
		rows[i] = importjob.ImportRow{MenuName: uniqueName(i), Price: 100 + i}
	}
	job := importjob.NewJob("job-3", rows)

	enqueueAndWait(t, w, job)

	if got := job.GetStatus(); got != importjob.StatusCompleted {
		t.Errorf("Status = %s, want completed", got)
	}
	if got := job.GetSucceededRows(); got != 25 {
		t.Errorf("SucceededRows = %d, want 25", got)
	}
}

// TestWorker_EmptyRows は空の rows が即完了することを確認する。
func TestWorker_EmptyRows(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobStore := importworker.NewInMemoryJobStore()
	menuReader := &testutil.StubMenuReader{Menus: nil}
	menuWriter := &testutil.SpyMenuWriter{}

	w := importworker.NewWorker(jobStore, menuReader, menuWriter)
	w.Start(ctx)

	job := importjob.NewJob("job-4", []importjob.ImportRow{})

	enqueueAndWait(t, w, job)

	if got := job.GetStatus(); got != importjob.StatusCompleted {
		t.Errorf("Status = %s, want completed", got)
	}
	if got := job.GetSucceededRows(); got != 0 {
		t.Errorf("SucceededRows = %d, want 0", got)
	}
}

// TestWorker_WriterFail は Writer.Save が失敗した行が RecordFailure されることを確認する。
func TestWorker_WriterFail(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobStore := importworker.NewInMemoryJobStore()
	menuReader := &testutil.StubMenuReader{Menus: nil}
	menuWriter := &testutil.SpyMenuWriter{SaveErr: errors.New("db error")}

	w := importworker.NewWorker(jobStore, menuReader, menuWriter)
	w.Start(ctx)

	rows := []importjob.ImportRow{
		{MenuName: "パスタ", Price: 900},
	}
	job := importjob.NewJob("job-5", rows)

	enqueueAndWait(t, w, job)

	if got := job.GetFailedRows(); got != 1 {
		t.Errorf("FailedRows = %d, want 1", got)
	}
}
