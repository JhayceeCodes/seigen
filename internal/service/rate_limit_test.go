package service_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JhayceeCodes/seigen/internal/limiter"
	"github.com/JhayceeCodes/seigen/internal/model"
	"github.com/JhayceeCodes/seigen/internal/service"
	"github.com/JhayceeCodes/seigen/internal/store"
)

type mockResolver struct {
	id model.Identifier
}

func (m *mockResolver) Resolve(r *http.Request) (model.Identifier, error) {
	return m.id, nil
}

func newTestRateLimitService(
	t *testing.T,
	policy model.Policy,
) (*service.RateLimitService, *http.Request) {
	t.Helper()

	policyStore := store.NewPolicyStore()

	if err := policyStore.Set(policy); err != nil {
		t.Fatalf("unexpected error setting policy: %v", err)
	}

	manager := limiter.NewManager()

	resolver := &mockResolver{
		id: policy.Identifier,
	}

	rateLimiter := service.NewRateLimitService(
		resolver,
		policyStore,
		manager,
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	return rateLimiter, req
}

func TestEvaluateAllowsValidRequest(t *testing.T) {
	policy := model.Policy{
		Identifier: "api-key:123",
		Limiter: model.LimiterConfig{
			Algorithm: model.TokenBucket,
			Config: model.TokenBucketConfig{
				Capacity:       2,
				RefillInterval: time.Second,
				RefillAmount:   1,
			},
		},
	}

	rateLimiter, req := newTestRateLimitService(t, policy)

	allowed, err := rateLimiter.Evaluate(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !allowed {
		t.Fatal("expected request to be allowed")
	}
}

func TestEvaluateEnforcesRateLimit(t *testing.T) {
	policy := model.Policy{
		Identifier: "api-key:123",
		Limiter: model.LimiterConfig{
			Algorithm: model.TokenBucket,
			Config: model.TokenBucketConfig{
				Capacity:       2,
				RefillInterval: time.Second,
				RefillAmount:   1,
			},
		},
	}

	rateLimiter, req := newTestRateLimitService(t, policy)

	// The bucket starts with two tokens, so the first two requests succeed.
	for i := range 2 {
		allowed, err := rateLimiter.Evaluate(req)

		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i+1, err)
		}

		if !allowed {
			t.Fatalf("request %d: expected request to be allowed", i+1)
		}
	}

	// The bucket is now empty.
	allowed, err := rateLimiter.Evaluate(req)

	if err != nil {
		t.Fatalf("third request: unexpected error: %v", err)
	}

	if allowed {
		t.Fatal("expected third request to be rate limited")
	}
}

func TestEvaluateEnforcesRefillLogic(t *testing.T) {
	policy := model.Policy{
		Identifier: "api-key:123",
		Limiter: model.LimiterConfig{
			Algorithm: model.TokenBucket,
			Config: model.TokenBucketConfig{
				Capacity:       5,
				RefillInterval: 500 * time.Millisecond,
				RefillAmount:   1,
			},
		},
	}

	rateLimiter, req := newTestRateLimitService(t, policy)

	// Consume three of the five available tokens.
	for i := range 3 {
		allowed, err := rateLimiter.Evaluate(req)

		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i+1, err)
		}

		if !allowed {
			t.Fatalf("request %d: expected request to be allowed", i+1)
		}
	}

	// Two tokens remain. After two seconds, four refill intervals
	// have passed, but the bucket is capped at its capacity of five.
	time.Sleep(2 * time.Second)

	// The bucket should now contain five tokens.
	for i := range 5 {
		allowed, err := rateLimiter.Evaluate(req)

		if err != nil {
			t.Fatalf("refilled request %d: unexpected error: %v", i+1, err)
		}

		if !allowed {
			t.Fatalf(
				"refilled request %d: expected request to be allowed",
				i+1,
			)
		}
	}

	// The bucket is empty again.
	allowed, err := rateLimiter.Evaluate(req)

	if err != nil {
		t.Fatalf("request after refill: unexpected error: %v", err)
	}

	if allowed {
		t.Fatal("expected request after refill capacity to be rate limited")
	}
}