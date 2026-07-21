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

	apiKey, err := service.NewAPIKey("free")
	if err != nil {
		http.Error(w, "error creating new api key", http.StatusInternalServerError)
		return
	}

	h.store.Create(apiKey)

	response := APIKeyResponse{
		Status:  "ok",
		Message: "API key created successfully",
		Data:    apiKey.Key,
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
