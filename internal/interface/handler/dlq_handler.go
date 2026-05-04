package handler

import (
	"fmt"
	"net/http"

	dlqcommand "mendo/internal/application/command/dlq"
	dlqquery "mendo/internal/application/query/dlq"
)

type DLQHandler struct {
	listHandler  *dlqquery.ListDLQHandler
	retryHandler *dlqcommand.RetryDLQUsecase
}

func NewDLQHandler(lh *dlqquery.ListDLQHandler, rh *dlqcommand.RetryDLQUsecase) *DLQHandler {
	return &DLQHandler{listHandler: lh, retryHandler: rh}
}

func (h *DLQHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	letters, err := h.listHandler.Handle(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("failed to list DLQ: %w", err).Error())
		return
	}

	type dlqResponse struct {
		ID          string `json:"id"`
		EventType   string `json:"event_type"`
		Error       string `json:"error"`
		FailCount   int    `json:"fail_count"`
		HandlerName string `json:"handler_name"`
	}

	result := make([]dlqResponse, 0, len(letters))
	for _, l := range letters {
		result = append(result, dlqResponse{
			ID:          l.ID,
			EventType:   l.Event.GetEventType(),
			Error:       l.Error,
			FailCount:   l.FailCount,
			HandlerName: l.HandlerName,
		})
	}

	writeSuccess(w, http.StatusOK, result)
}

func (h *DLQHandler) HandleRetry(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.retryHandler.Execute(r.Context(), id); err != nil {
		writeError(w, http.StatusUnprocessableEntity, fmt.Errorf("failed to retry: %w", err).Error())
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"status": "retried"})
}
