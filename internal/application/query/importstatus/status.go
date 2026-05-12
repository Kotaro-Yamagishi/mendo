package importstatus

import (
	"context"

	"mendo/internal/apperrors"
	importjob "mendo/internal/domain/import"
)

// ImportStatusHandler はインポートジョブのプログレス確認。
type ImportStatusHandler struct {
	jobReader importjob.JobReader
}

// NewImportStatusHandler は ImportStatusHandler を作成する。
func NewImportStatusHandler(jr importjob.JobReader) *ImportStatusHandler {
	return &ImportStatusHandler{jobReader: jr}
}

// StatusResponse はプログレスのレスポンス。
type StatusResponse struct {
	ID            string                `json:"id"`
	Status        string                `json:"status"`
	TotalRows     int                   `json:"total_rows"`
	ProcessedRows int                   `json:"processed_rows"`
	SucceededRows int                   `json:"succeeded_rows"`
	FailedRows    int                   `json:"failed_rows"`
	Progress      int                   `json:"progress"`
	Errors        []importjob.ImportError `json:"errors"`
}

// Handle はジョブのプログレスを返す。
func (h *ImportStatusHandler) Handle(ctx context.Context, jobID string) (*StatusResponse, error) {
	job, err := h.jobReader.FindByID(ctx, importjob.JobID(jobID))
	if err != nil {
		return nil, apperrors.NotFound("import_job", jobID)
	}
	return &StatusResponse{
		ID:            string(job.ID),
		Status:        string(job.Status),
		TotalRows:     job.TotalRows,
		ProcessedRows: job.ProcessedRows,
		SucceededRows: job.SucceededRows,
		FailedRows:    job.FailedRows,
		Progress:      job.Progress(),
		Errors:        job.Errors,
	}, nil
}
