package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/JhayceeCodes/rate-limiter-gateway/internal/service"
	"github.com/JhayceeCodes/rate-limiter-gateway/internal/store"
)

type APIKeyHandler struct {
	store *store.APIKeyStore
}

func NewAPIKeyHandler(apiKeyStore *store.APIKeyStore) *APIKeyHandler {
	return &APIKeyHandler{
		store: apiKeyStore,
	}
}

func (h *APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateAPIKeyRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	apiKey, err := service.NewAPIKey(req.Tier)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.store.Create(apiKey)

	response := APIKeyResponse{
		Status:  "ok",
		Message: "API key created successfully",
		Data:    apiKey,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("failed to encode response: %v", err)
	}

}

func (h *APIKeyHandler) Retrieve() {

}

func (h *APIKeyHandler) List() {

}

func (h *APIKeyHandler) Delete() {

}
