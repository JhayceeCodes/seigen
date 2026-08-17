package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/JhayceeCodes/seigen/internal/limiter"
	"github.com/JhayceeCodes/seigen/internal/middleware"
)

func Hello(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"message": "Hello from Seigen!",
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)

}

func main() {
	bucket := limiter.NewTokenBucket(
		3,
		30*time.Second,
		1,
	)

	handler := middleware.RateLimit(
		bucket,
		http.HandlerFunc(Hello),
	)

	http.Handle("/hello", handler)

	fmt.Println("Server running on 8081.")
	http.ListenAndServe(":8081", nil)
}
