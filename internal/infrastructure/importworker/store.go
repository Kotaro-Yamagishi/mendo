package importworker

import (
	"context"
	"fmt"
	"sync"

	importjob "mendo/internal/domain/import"
)

// InMemoryJobStore はインポートジョブのインメモリストア。
type InMemoryJobStore struct {
	mu   sync.RWMutex
	jobs map[string]*importjob.Job
}

// NewInMemoryJobStore は InMemoryJobStore を作成する。
func NewInMemoryJobStore() *InMemoryJobStore {
	return &InMemoryJobStore{jobs: make(map[string]*importjob.Job)}
}

// Save はジョブを保存する。
func (s *InMemoryJobStore) Save(_ context.Context, job *importjob.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[string(job.ID)] = job
	return nil
}

// FindByID はジョブを ID で取得する。
func (s *InMemoryJobStore) FindByID(_ context.Context, id importjob.JobID) (*importjob.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, ok := s.jobs[string(id)]
	if !ok {
		return nil, fmt.Errorf("import job not found: %s", id)
	}
	return job, nil
}
