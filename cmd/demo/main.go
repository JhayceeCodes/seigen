package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/JhayceeCodes/seigen/internal/limiter"
	"github.com/JhayceeCodes/seigen/internal/middleware"
	"github.com/JhayceeCodes/seigen/internal/model"
	"github.com/JhayceeCodes/seigen/internal/store"
)

func Hello(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"message": "Hello from Seigen!",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	policyStore := store.NewPolicyStore()

	policy := model.Policy{
		Identifier: "user:123",
		Limiter: model.LimiterConfig{
			Algorithm: model.TokenBucket,
			Config: model.TokenBucketConfig{
				Capacity:       3,
				RefillInterval: 10 * time.Second,
				RefillAmount:   1,
			},
		},
	}

	if err := policyStore.Set(policy); err != nil {
		panic(err)
	}

	policy, err := policyStore.Get("user:123")
	if err != nil {
		panic(err)
	}

	config := policy.Limiter.Config.(model.TokenBucketConfig)

	bucket := limiter.NewTokenBucket(
		config.Capacity,
		config.RefillInterval,
		config.RefillAmount,
	)

	handler := middleware.RateLimit(
		bucket,
		http.HandlerFunc(Hello),
	)

	http.Handle("/hello", handler)

	fmt.Println("Server running on 8081.")
	http.ListenAndServe(":8081", nil)
}
