package router

import (
	"net/http"

	"github.com/JhayceeCodes/rate-limiter-gateway/internal/handler"
)

func RegisterRoutes(apiKeyHandler *handler.APIKeyHandler) {
	http.HandleFunc("/apikeys", apiKeyHandler.Create)
}
