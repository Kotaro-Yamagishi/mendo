package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type successResponse struct {
	Data interface{} `json:"data"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeSuccess(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(successResponse{Data: data}); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(errorResponse{Error: msg}); err != nil {
		log.Printf("failed to write error response: %v", err)
	}
}

func readJSON(r *http.Request, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to decode request body: %w", err)
	}
	return nil
}
