package router

import (
	"net/http"

	"github.com/JhayceeCodes/seigen/internal/handler"
)

func RegisterAPIKeyRoutes(apiKeyHandler *handler.APIKeyHandler) {
	http.HandleFunc("POST /apikeys", apiKeyHandler.Create)
	http.HandleFunc("GET /apikeys", apiKeyHandler.List)
	http.HandleFunc("GET /apikeys/{key}", apiKeyHandler.GetByKey)
	http.HandleFunc("DELETE /apikeys/{key}", apiKeyHandler.Delete)
}
