package testutil

import (
	"context"
	"errors"

	importjob "mendo/internal/domain/import"
)

// StubJobWriter は importjob.JobWriter の Stub + Spy。
type StubJobWriter struct {
	SaveErr   error
	SavedJobs []*importjob.Job
}

func (s *StubJobWriter) Save(_ context.Context, job *importjob.Job) error {
	if s.SaveErr != nil {
		return s.SaveErr
	}
	s.SavedJobs = append(s.SavedJobs, job)
	return nil
}

// SpyJobEnqueuer は importjob.JobEnqueuer の Spy。
type SpyJobEnqueuer struct {
	EnqueuedJobs []*importjob.Job
}

func (s *SpyJobEnqueuer) Enqueue(job *importjob.Job) {
	s.EnqueuedJobs = append(s.EnqueuedJobs, job)
}

// StubJobReader は importjob.JobReader の Stub。
type StubJobReader struct {
	Job     *importjob.Job
	FindErr error
}

func (s *StubJobReader) FindByID(_ context.Context, _ importjob.JobID) (*importjob.Job, error) {
	if s.FindErr != nil {
		return nil, s.FindErr
	}
	if s.Job == nil {
		return nil, errors.New("job not found")
	}
	return s.Job, nil
}
