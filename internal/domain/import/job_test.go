package importjob_test

import (
	"testing"

	importjob "mendo/internal/domain/import"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewJob(t *testing.T) {
	t.Parallel()

	rows := []importjob.ImportRow{
		{MenuName: "ラーメン", Price: 800},
		{MenuName: "つけ麺", Price: 900},
	}

	job := importjob.NewJob("job-1", rows)

	assert.Equal(t, importjob.JobID("job-1"), job.ID)
	assert.Equal(t, importjob.StatusQueued, job.Status)
	assert.Equal(t, 2, job.TotalRows)
	assert.Equal(t, 0, job.ProcessedRows)
	assert.Equal(t, 0, job.SucceededRows)
	assert.Equal(t, 0, job.FailedRows)
	assert.Empty(t, job.Errors)
	assert.Nil(t, job.CompletedAt)
}

func Test_Job_Start(t *testing.T) {
	t.Parallel()

	job := importjob.NewJob("job-1", []importjob.ImportRow{{MenuName: "ラーメン", Price: 800}})

	job.Start()

	assert.Equal(t, importjob.StatusProcessing, job.Status)
}

func Test_Job_RecordSuccess(t *testing.T) {
	t.Parallel()

	job := importjob.NewJob("job-1", []importjob.ImportRow{{MenuName: "ラーメン", Price: 800}})
	job.Start()

	job.RecordSuccess()

	assert.Equal(t, 1, job.ProcessedRows)
	assert.Equal(t, 1, job.SucceededRows)
	assert.Equal(t, 0, job.FailedRows)
}

func Test_Job_RecordFailure(t *testing.T) {
	t.Parallel()

	job := importjob.NewJob("job-1", []importjob.ImportRow{{MenuName: "ラーメン", Price: 800}})
	job.Start()

	job.RecordFailure(1, "ラーメン", "価格が不正")

	assert.Equal(t, 1, job.ProcessedRows)
	assert.Equal(t, 0, job.SucceededRows)
	assert.Equal(t, 1, job.FailedRows)
	require.Len(t, job.Errors, 1)
	assert.Equal(t, 1, job.Errors[0].Row)
	assert.Equal(t, "ラーメン", job.Errors[0].Name)
	assert.Equal(t, "価格が不正", job.Errors[0].Reason)
}

func Test_Job_Complete_全成功(t *testing.T) {
	t.Parallel()

	job := importjob.NewJob("job-1", []importjob.ImportRow{
		{MenuName: "ラーメン", Price: 800},
		{MenuName: "つけ麺", Price: 900},
	})
	job.Start()
	job.RecordSuccess()
	job.RecordSuccess()

	job.Complete()

	assert.Equal(t, importjob.StatusCompleted, job.Status)
	assert.NotNil(t, job.CompletedAt)
}

func Test_Job_Complete_全失敗(t *testing.T) {
	t.Parallel()

	job := importjob.NewJob("job-1", []importjob.ImportRow{
		{MenuName: "ラーメン", Price: 800},
	})
	job.Start()
	job.RecordFailure(1, "ラーメン", "エラー")

	job.Complete()

	// 全失敗 → StatusFailed
	assert.Equal(t, importjob.StatusFailed, job.Status)
}

func Test_Job_Complete_一部失敗(t *testing.T) {
	t.Parallel()

	job := importjob.NewJob("job-1", []importjob.ImportRow{
		{MenuName: "ラーメン", Price: 800},
		{MenuName: "つけ麺", Price: 900},
	})
	job.Start()
	job.RecordSuccess()
	job.RecordFailure(2, "つけ麺", "エラー")

	job.Complete()

	// 一部成功あり → StatusCompleted（StatusFailed にはならない）
	assert.Equal(t, importjob.StatusCompleted, job.Status)
}

func Test_Job_Progress(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		total     int
		processed int
		want      int
	}{
		{"0/2", 2, 0, 0},
		{"1/2", 2, 1, 50},
		{"2/2", 2, 2, 100},
		{"1/3", 3, 1, 33},
		{"空ジョブ", 0, 0, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rows := make([]importjob.ImportRow, tt.total)
			job := importjob.NewJob("job-1", rows)
			for i := 0; i < tt.processed; i++ {
				job.RecordSuccess()
			}
			assert.Equal(t, tt.want, job.Progress())
		})
	}
}

func Test_Job_Summary(t *testing.T) {
	t.Parallel()

	job := importjob.NewJob("job-1", []importjob.ImportRow{
		{MenuName: "ラーメン", Price: 800},
		{MenuName: "つけ麺", Price: 900},
	})
	job.Start()
	job.RecordSuccess()
	job.RecordFailure(2, "つけ麺", "エラー")

	summary := job.Summary()

	assert.Contains(t, summary, "total=2")
	assert.Contains(t, summary, "succeeded=1")
	assert.Contains(t, summary, "failed=1")
	assert.Contains(t, summary, "progress=100%")
}
