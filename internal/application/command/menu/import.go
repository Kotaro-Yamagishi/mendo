package menu

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	importjob "mendo/internal/domain/import"
)

// ImportMenusUsecase はメニュー一括インポートの受付ユースケース。
// CSV を受け取ってジョブを作成し、ワーカーに投げて即座に返す。
type ImportMenusUsecase struct {
	jobWriter  importjob.JobWriter
	jobEnqueuer importjob.JobEnqueuer
}

// NewImportMenusUsecase は ImportMenusUsecase を作成する。
func NewImportMenusUsecase(jw importjob.JobWriter, je importjob.JobEnqueuer) *ImportMenusUsecase {
	return &ImportMenusUsecase{jobWriter: jw, jobEnqueuer: je}
}

// Execute はインポートジョブを作成し、ワーカーに投入して即座に jobID を返す。
func (uc *ImportMenusUsecase) Execute(ctx context.Context, rows []importjob.ImportRow) (string, error) {
	// 1. ジョブ作成
	jobID := importjob.JobID(uuid.New().String())
	job := importjob.NewJob(jobID, rows)

	// 2. ジョブを保存（プログレス確認用）
	if err := uc.jobWriter.Save(ctx, job); err != nil {
		return "", fmt.Errorf("failed to save import job: %w", err)
	}

	// 3. ワーカーに投げる（非同期。即座に返る）
	uc.jobEnqueuer.Enqueue(job)

	return string(jobID), nil
}
