package main

import (
	"fmt"
	"net/http"

	"github.com/JhayceeCodes/rate-limiter-gateway/internal/handler"
)

func main() {

	http.HandleFunc("/health", handler.Health)

	fmt.Println("Server running on 8080.")

	http.ListenAndServe(":8080", nil)

}
