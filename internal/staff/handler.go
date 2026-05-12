package staff

import (
	"encoding/json"
	"net/http"

	"mendo/internal/apperrors"
	"mendo/internal/interface/handler"
)

// Handler はスタッフ管理の HTTP ハンドラ。
// アクティブレコードなので usecase 層がない。handler が直接 Store を呼ぶ。
type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) error {
	staffs, err := h.store.FindAll(r.Context())
	if err != nil {
		return err
	}
	handler.WriteSuccess(w, http.StatusOK, staffs)
	return nil
}

func (h *Handler) HandleGetByID(w http.ResponseWriter, r *http.Request) error {
	s, err := h.store.FindByID(r.Context(), r.PathValue("id"))
	if err != nil {
		return err
	}
	handler.WriteSuccess(w, http.StatusOK, s)
	return nil
}

func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) error {
	var s Staff
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		return apperrors.Validation("INVALID_REQUEST_BODY", "invalid request body")
	}
	if err := h.store.Save(r.Context(), &s); err != nil {
		return err
	}
	handler.WriteSuccess(w, http.StatusCreated, s)
	return nil
}

func (h *Handler) HandleDelete(w http.ResponseWriter, r *http.Request) error {
	if err := h.store.Delete(r.Context(), r.PathValue("id")); err != nil {
		return err
	}
	handler.WriteSuccess(w, http.StatusOK, map[string]string{"status": "deleted"})
	return nil
}
