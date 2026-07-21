package main

import (
	"fmt"
	"net/http"

	"github.com/JhayceeCodes/rate-limiter-gateway/internal/handler"
	"github.com/JhayceeCodes/rate-limiter-gateway/internal/router"
	"github.com/JhayceeCodes/rate-limiter-gateway/internal/store"
)

func main() {

	http.HandleFunc("/health", handler.Health)

	apiKeyStore := store.NewAPIKeyStore()
	apiKeyHandler := handler.NewAPIKeyHandler(apiKeyStore)

	router.RegisterRoutes(apiKeyHandler)

	fmt.Println("Server running on 8080.")

	http.ListenAndServe(":8080", nil)

}
