package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/JhayceeCodes/seigen/internal/identifier"
	"github.com/JhayceeCodes/seigen/internal/limiter"
	"github.com/JhayceeCodes/seigen/internal/middleware"
	"github.com/JhayceeCodes/seigen/internal/model"
	"github.com/JhayceeCodes/seigen/internal/service"
	"github.com/JhayceeCodes/seigen/internal/store"
)

func main() {

	policyStore := store.NewPolicyStore()

	policies := []model.Policy{
		{
			// APIKeyResolver returns the API key itself.
			Identifier: "key-1",
			Limiter: model.LimiterConfig{
				Algorithm: model.TokenBucket,
				Config: model.TokenBucketConfig{
					Capacity:       2,
					RefillInterval: 5 * time.Second,
					RefillAmount:   1,
				},
			},
		},
		{
			Identifier: "key-2",
			Limiter: model.LimiterConfig{
				Algorithm: model.TokenBucket,
				Config: model.TokenBucketConfig{
					Capacity:       5,
					RefillInterval: 5 * time.Second,
					RefillAmount:   1,
				},
			},
		},
	}

	for _, policy := range policies {
		if err := policyStore.Set(policy); err != nil {
			log.Fatalf("failed to set policy: %v", err)
		}
	}

	// --------------------------------------------------
	// Limiter manager
	// --------------------------------------------------

	manager := limiter.NewManager()

	// --------------------------------------------------
	// Built-in API key resolver
	// --------------------------------------------------

	resolver := identifier.NewAPIKeyResolver()

	// --------------------------------------------------
	// Rate limit service
	// --------------------------------------------------

	rateLimitService := service.NewRateLimitService(
		resolver,
		policyStore,
		manager,
	)

	// --------------------------------------------------
	// Application handler
	// --------------------------------------------------

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from Seigen!")
	})

	// --------------------------------------------------
	// Rate-limit middleware
	// --------------------------------------------------

	rateLimitedHandler := middleware.RateLimit(
		rateLimitService,
		handler,
	)

	// --------------------------------------------------
	// HTTP server
	// --------------------------------------------------

	log.Println("Seigen running on http://localhost:8081")

	if err := http.ListenAndServe(":8081", rateLimitedHandler); err != nil {
		log.Fatal(err)
	}
}