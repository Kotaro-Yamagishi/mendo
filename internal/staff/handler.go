package staff

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Handler はスタッフ管理の HTTP ハンドラ。
// アクティブレコードなので usecase 層がない。handler が直接 Store を呼ぶ。
type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /staff", h.HandleList)
	mux.HandleFunc("GET /staff/{id}", h.HandleGetByID)
	mux.HandleFunc("POST /staff", h.HandleCreate)
	mux.HandleFunc("DELETE /staff/{id}", h.HandleDelete)
}

func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	staffs, err := h.store.FindAll(r.Context())
	if err != nil {
		http.Error(w, fmt.Errorf("failed to list staff: %w", err).Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, staffs)
}

func (h *Handler) HandleGetByID(w http.ResponseWriter, r *http.Request) {
	s, err := h.store.FindByID(r.Context(), r.PathValue("id"))
	if err != nil {
		http.Error(w, fmt.Errorf("failed to find staff: %w", err).Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var s Staff
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.store.Save(r.Context(), &s); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	writeJSON(w, http.StatusCreated, s)
}

func (h *Handler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Delete(r.Context(), r.PathValue("id")); err != nil {
		http.Error(w, fmt.Errorf("failed to delete staff: %w", err).Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
