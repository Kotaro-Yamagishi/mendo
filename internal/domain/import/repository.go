package importjob

import "context"

// JobReader はインポートジョブの読み取り IF。
type JobReader interface {
	FindByID(ctx context.Context, id JobID) (*Job, error)
}

// JobWriter はインポートジョブの書き込み IF。
type JobWriter interface {
	Save(ctx context.Context, job *Job) error
}

// JobEnqueuer はインポートジョブをキューに投入する IF。
type JobEnqueuer interface {
	Enqueue(job *Job)
}
