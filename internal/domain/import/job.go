package importjob

import (
	"fmt"
	"sync"
	"time"
)

// JobID はインポートジョブの識別子。
type JobID string

// JobStatus はジョブのステータス。
type JobStatus string

const (
	StatusQueued     JobStatus = "queued"     // 受付済み。処理待ち
	StatusProcessing JobStatus = "processing" // 処理中
	StatusCompleted  JobStatus = "completed"  // 完了
	StatusFailed     JobStatus = "failed"     // 失敗
)

// ImportRow は CSV の1行。
type ImportRow struct {
	MenuName string
	Price    int
}

// ImportError は処理失敗した行の情報。
type ImportError struct {
	Row    int    `json:"row"`
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

// Job は非同期バッチインポートのジョブ。プログレスを管理する。
type Job struct {
	mu            sync.RWMutex
	ID            JobID         `json:"id"`
	Status        JobStatus     `json:"status"`
	TotalRows     int           `json:"total_rows"`
	ProcessedRows int           `json:"processed_rows"`
	SucceededRows int           `json:"succeeded_rows"`
	FailedRows    int           `json:"failed_rows"`
	Errors        []ImportError `json:"errors"`
	CreatedAt     time.Time     `json:"created_at"`
	CompletedAt   *time.Time    `json:"completed_at"`
	Rows          []ImportRow   `json:"-"` // 処理対象データ（レスポンスには含めない）
}

// NewJob はジョブを作成する。ステータスは queued。
func NewJob(id JobID, rows []ImportRow) *Job {
	return &Job{
		ID:        id,
		Status:    StatusQueued,
		TotalRows: len(rows),
		Rows:      rows,
		CreatedAt: time.Now(),
	}
}

// Start はジョブを処理中にする。
func (j *Job) Start() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Status = StatusProcessing
}

// RecordSuccess は成功を記録する。
func (j *Job) RecordSuccess() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.ProcessedRows++
	j.SucceededRows++
}

// RecordFailure は失敗を記録する。
func (j *Job) RecordFailure(row int, name, reason string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.ProcessedRows++
	j.FailedRows++
	j.Errors = append(j.Errors, ImportError{Row: row, Name: name, Reason: reason})
}

// Complete はジョブを完了にする。
func (j *Job) Complete() {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.FailedRows > 0 && j.SucceededRows == 0 {
		j.Status = StatusFailed
	} else {
		j.Status = StatusCompleted
	}
	now := time.Now()
	j.CompletedAt = &now
}

// Progress はプログレス（%）を返す。
func (j *Job) Progress() int {
	j.mu.RLock()
	defer j.mu.RUnlock()
	if j.TotalRows == 0 {
		return 100
	}
	return (j.ProcessedRows * 100) / j.TotalRows
}

// Summary はジョブのサマリーを返す。
func (j *Job) Summary() string {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return fmt.Sprintf("total=%d succeeded=%d failed=%d progress=%d%%",
		j.TotalRows, j.SucceededRows, j.FailedRows, j.Progress())
}
