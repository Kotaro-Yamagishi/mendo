package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	dlqcommand "mendo/internal/application/command/dlq"
	dlqquery "mendo/internal/application/query/dlq"
	"mendo/internal/domain"
	"mendo/internal/domain/order"
	"mendo/internal/interface/handler"
	"mendo/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDLQHandler(dlq *testutil.StubDeadLetterQueue, pub *testutil.SpyEventPublisher) *handler.DLQHandler {
	listH := dlqquery.NewListDLQHandler(dlq)
	retryUC := dlqcommand.NewRetryDLQUsecase(dlq, pub)
	return handler.NewDLQHandler(listH, retryUC)
}

// =============================================================================
// HandleList
// =============================================================================

func Test_DLQHandler_HandleList(t *testing.T) {
	t.Parallel()

	event := order.NewOrderCreated("order-1", 3, "")

	tests := []struct {
		name           string
		dlq            *testutil.StubDeadLetterQueue
		wantStatus     int
		wantDataLen    int // -1 = skip check
		wantFirstID    string
		wantFirstError string
		wantFirstFail  float64
	}{
		{
			name: "正常系",
			dlq: &testutil.StubDeadLetterQueue{
				Letters: []domain.DeadLetter{
					{
						ID:          "dlq-1",
						Event:       event,
						Error:       "timeout",
						FailCount:   2,
						HandlerName: "KitchenSubscriber",
						LastFailAt:  time.Now(),
					},
				},
			},
			wantStatus:     http.StatusOK,
			wantDataLen:    1,
			wantFirstID:    "dlq-1",
			wantFirstError: "timeout",
			wantFirstFail:  2,
		},
		{
			name:        "空",
			dlq:         &testutil.StubDeadLetterQueue{Letters: nil},
			wantStatus:  http.StatusOK,
			wantDataLen: 0,
		},
		{
			name:        "DLQ取得エラー_インフラエラーは500",
			dlq:         &testutil.StubDeadLetterQueue{ListErr: errors.New("db error")},
			wantStatus:  http.StatusInternalServerError,
			wantDataLen: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pub := &testutil.SpyEventPublisher{}
			h := newDLQHandler(tt.dlq, pub)

			req := httptest.NewRequest(http.MethodGet, "/admin/dlq", nil)
			rec := httptest.NewRecorder()

			wrap(h.HandleList)(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantDataLen < 0 {
				return
			}

			var resp map[string]interface{}
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
			data := resp["data"].([]interface{})
			require.Len(t, data, tt.wantDataLen)

			if tt.wantDataLen == 0 {
				assert.Empty(t, data)
				return
			}

			item := data[0].(map[string]interface{})
			assert.Equal(t, tt.wantFirstID, item["id"])
			assert.Equal(t, order.EventTypeOrderCreated, item["event_type"])
			assert.Equal(t, tt.wantFirstError, item["error"])
			assert.Equal(t, tt.wantFirstFail, item["fail_count"])
		})
	}
}

// =============================================================================
// HandleRetry
// =============================================================================

func Test_DLQHandler_HandleRetry(t *testing.T) {
	t.Parallel()

	event := order.NewOrderCreated("order-1", 3, "")

	tests := []struct {
		name            string
		dlq             *testutil.StubDeadLetterQueue
		pathID          string
		wantStatus      int
		wantPublished   int
		wantRemovedID   string
		wantRespStatus  string
	}{
		{
			name: "正常系",
			dlq: &testutil.StubDeadLetterQueue{
				SingleLetter: &domain.DeadLetter{
					ID:    "dlq-1",
					Event: event,
				},
			},
			pathID:         "dlq-1",
			wantStatus:     http.StatusOK,
			wantPublished:  1,
			wantRemovedID:  "dlq-1",
			wantRespStatus: "retried",
		},
		{
			name:          "存在しないID_NotFoundは404",
			dlq:           &testutil.StubDeadLetterQueue{FindErr: errors.New("not found")},
			pathID:        "nonexistent",
			wantStatus:    http.StatusNotFound,
			wantPublished: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pub := &testutil.SpyEventPublisher{}
			h := newDLQHandler(tt.dlq, pub)

			req := httptest.NewRequest(http.MethodPost, "/admin/dlq/"+tt.pathID+"/retry", nil)
			req.SetPathValue("id", tt.pathID)
			rec := httptest.NewRecorder()

			wrap(h.HandleRetry)(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			require.Len(t, pub.Published, tt.wantPublished)

			if tt.wantRespStatus == "" {
				return
			}

			require.Len(t, tt.dlq.RemovedIDs, 1)
			assert.Equal(t, tt.wantRemovedID, tt.dlq.RemovedIDs[0])

			var resp map[string]interface{}
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
			data := resp["data"].(map[string]interface{})
			assert.Equal(t, tt.wantRespStatus, data["status"])
		})
	}
}
