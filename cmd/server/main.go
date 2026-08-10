package main

import (
	"fmt"
	"net/http"

	"github.com/JhayceeCodes/seigen/internal/handler"
	"github.com/JhayceeCodes/seigen/internal/router"
	"github.com/JhayceeCodes/seigen/internal/store"
)

func main() {

	http.HandleFunc("/health", handler.Health)

	apiKeyStore := store.NewAPIKeyStore()
	apiKeyHandler := handler.NewAPIKeyHandler(apiKeyStore)

	router.RegisterAPIKeyRoutes(apiKeyHandler)

	fmt.Println("Server running on 8080.")

	http.ListenAndServe(":8080", nil)

}
