package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appmenu "mendo/internal/application/command/menu"
	"mendo/internal/application/query/importstatus"
	importjob "mendo/internal/domain/import"
	"mendo/internal/interface/handler"
	"mendo/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newImportHandler(jw *testutil.StubJobWriter, je *testutil.SpyJobEnqueuer, jr *testutil.StubJobReader) *handler.ImportHandler {
	importUC := appmenu.NewImportMenusUsecase(jw, je)
	statusUC := importstatus.NewImportStatusHandler(jr)
	return handler.NewImportHandler(importUC, statusUC)
}

// =============================================================================
// HandleImportMenus
// =============================================================================

func Test_ImportHandler_HandleImportMenus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		body              string
		jw                *testutil.StubJobWriter
		wantStatus        int
		wantSavedJobs     int
		wantTotalRows     int
		wantEnqueuedJobs  int
		wantJobID         bool
		wantRespStatus    string
	}{
		{
			name:             "正常系",
			body:             `{"items":[{"menu_name":"醤油ラーメン","price":800},{"menu_name":"味噌ラーメン","price":900}]}`,
			jw:               &testutil.StubJobWriter{},
			wantStatus:       http.StatusAccepted,
			wantSavedJobs:    1,
			wantTotalRows:    2,
			wantEnqueuedJobs: 1,
			wantJobID:        true,
			wantRespStatus:   "queued",
		},
		{
			name:             "不正JSON",
			body:             "invalid json",
			jw:               &testutil.StubJobWriter{},
			wantStatus:       http.StatusBadRequest,
			wantSavedJobs:    0,
			wantEnqueuedJobs: 0,
		},
		{
			name:             "JobWriter失敗",
			body:             `{"items":[{"menu_name":"ラーメン","price":700}]}`,
			jw:               &testutil.StubJobWriter{SaveErr: errors.New("db error")},
			wantStatus:       http.StatusInternalServerError,
			wantSavedJobs:    0,
			wantEnqueuedJobs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			je := &testutil.SpyJobEnqueuer{}
			jr := &testutil.StubJobReader{}
			h := newImportHandler(tt.jw, je, jr)

			req := httptest.NewRequest(http.MethodPost, "/admin/import/menus", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			wrap(h.HandleImportMenus)(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			require.Len(t, tt.jw.SavedJobs, tt.wantSavedJobs)
			require.Len(t, je.EnqueuedJobs, tt.wantEnqueuedJobs)

			if tt.wantRespStatus == "" {
				return
			}

			assert.Equal(t, tt.wantTotalRows, tt.jw.SavedJobs[0].TotalRows)

			var resp map[string]interface{}
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
			data := resp["data"].(map[string]interface{})
			if tt.wantJobID {
				assert.NotEmpty(t, data["job_id"])
			}
			assert.Equal(t, tt.wantRespStatus, data["status"])
		})
	}
}

// =============================================================================
// HandleImportStatus
// =============================================================================

func Test_ImportHandler_HandleImportStatus(t *testing.T) {
	t.Parallel()

	completedJob := importjob.NewJob("job-1", []importjob.ImportRow{
		{MenuName: "醤油ラーメン", Price: 800},
	})
	completedJob.Start()
	completedJob.RecordSuccess()
	completedJob.Complete()

	tests := []struct {
		name           string
		jr             *testutil.StubJobReader
		pathID         string
		wantStatus     int
		wantJobID      string
		wantJobStatus  string
		wantProgress   float64
	}{
		{
			name:          "正常系",
			jr:            &testutil.StubJobReader{Job: completedJob},
			pathID:        "job-1",
			wantStatus:    http.StatusOK,
			wantJobID:     "job-1",
			wantJobStatus: "completed",
			wantProgress:  100,
		},
		{
			name:       "存在しないジョブ",
			jr:         &testutil.StubJobReader{FindErr: errors.New("not found")},
			pathID:     "nonexistent",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			jw := &testutil.StubJobWriter{}
			je := &testutil.SpyJobEnqueuer{}
			h := newImportHandler(jw, je, tt.jr)

			req := httptest.NewRequest(http.MethodGet, "/admin/import/"+tt.pathID+"/status", nil)
			req.SetPathValue("id", tt.pathID)
			rec := httptest.NewRecorder()

			wrap(h.HandleImportStatus)(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantJobID == "" {
				return
			}

			var resp map[string]interface{}
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
			data := resp["data"].(map[string]interface{})
			assert.Equal(t, tt.wantJobID, data["id"])
			assert.Equal(t, tt.wantJobStatus, data["status"])
			assert.Equal(t, tt.wantProgress, data["progress"])
		})
	}
}
