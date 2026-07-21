package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/JhayceeCodes/rate-limiter-gateway/internal/service"
	"github.com/JhayceeCodes/rate-limiter-gateway/internal/store"
)

type APIKeyHandler struct {
	store map[string]store.APIKeyStore
}

func NewAPIKeyHandler() *APIKeyHandler {
	return &APIKeyHandler{
		store: make(map[string]store.APIKeyStore),
	}
}

func (h *APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	apiKey, err := service.NewAPIKey("free")
	if err != nil {
		fmt.Printf("Error creating API key: %v\n", err)
	}

	newStore := store.NewAPIKeyStore()

	newStore.Create(apiKey)

	response := map[string]string{
		"message": "API key created successfully",
		"data":    apiKey.Key,
		"status":  "ok",
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)

}

func (h *APIKeyHandler) Retrieve() {

}

func (h *APIKeyHandler) List() {

}

func (h *APIKeyHandler) Delete() {

}
