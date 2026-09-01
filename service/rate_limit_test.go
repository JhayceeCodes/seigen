package service_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JhayceeCodes/seigen/limiter"
	"github.com/JhayceeCodes/seigen/model"
	"github.com/JhayceeCodes/seigen/service"
	"github.com/JhayceeCodes/seigen/store"
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

	policyStore := store.NewInMemoryPolicyRepository()

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

	result, err := rateLimiter.Evaluate(req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Limit != 2 {
		t.Fatalf("expected limit 2, got %d", result.Limit)
	}

	if !result.Allowed {
		t.Fatal("expected request to be allowed")
	}

	if result.Remaining != 1 {
		t.Fatalf("expected 1 remaining, got %d", result.Remaining)
	}

	if result.RetryAfter != 0 {
		t.Fatalf("expected no retry delay, got %v", result.RetryAfter)
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

	expectedRemaining := 1

	for i := range 2 {
		result, err := rateLimiter.Evaluate(req)

		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i+1, err)
		}

		if result.Limit != 2 {
			t.Fatalf(
				"request %d: expected limit 2, got %d",
				i+1,
				result.Limit,
			)
		}

		if !result.Allowed {
			t.Fatalf(
				"request %d: expected request to be allowed",
				i+1,
			)
		}

		if result.Remaining != expectedRemaining {
			t.Fatalf(
				"request %d: expected %d remaining, got %d",
				i+1,
				expectedRemaining,
				result.Remaining,
			)
		}

		expectedRemaining--
	}

	result, err := rateLimiter.Evaluate(req)

	if err != nil {
		t.Fatalf("third request: unexpected error: %v", err)
	}

	if result.Limit != 2 {
		t.Fatalf("expected limit 2, got %d", result.Limit)
	}

	if result.Allowed {
		t.Fatal("expected third request to be rate limited")
	}

	if result.Remaining != 0 {
		t.Fatalf("expected 0 remaining, got %d", result.Remaining)
	}

	if result.RetryAfter <= 0 {
		t.Fatalf(
			"expected positive retry delay, got %v",
			result.RetryAfter,
		)
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
		result, err := rateLimiter.Evaluate(req)

		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i+1, err)
		}

		if !result.Allowed {
			t.Fatalf(
				"request %d: expected request to be allowed",
				i+1,
			)
		}
	}

	// Two tokens remain. After two seconds, four refill intervals
	// have passed, but the bucket is capped at its capacity of five.
	time.Sleep(2 * time.Second)

	// The bucket should now contain five tokens.
	expectedRemaining := 4

	for i := range 5 {
		result, err := rateLimiter.Evaluate(req)

		if err != nil {
			t.Fatalf(
				"refilled request %d: unexpected error: %v",
				i+1,
				err,
			)
		}

		if result.Limit != 5 {
			t.Fatalf(
				"refilled request %d: expected limit 5, got %d",
				i+1,
				result.Limit,
			)
		}

		if !result.Allowed {
			t.Fatalf(
				"refilled request %d: expected request to be allowed",
				i+1,
			)
		}

		if result.Remaining != expectedRemaining {
			t.Fatalf(
				"refilled request %d: expected %d remaining, got %d",
				i+1,
				expectedRemaining,
				result.Remaining,
			)
		}

		expectedRemaining--
	}

	// The bucket is empty again.
	result, err := rateLimiter.Evaluate(req)

	if err != nil {
		t.Fatalf(
			"request after refill: unexpected error: %v",
			err,
		)
	}

	if result.Limit != 5 {
		t.Fatalf("expected limit 5, got %d", result.Limit)
	}

	if result.Allowed {
		t.Fatal("expected request after refill capacity to be rate limited")
	}

	if result.Remaining != 0 {
		t.Fatalf(
			"expected 0 remaining after refill capacity was exhausted, got %d",
			result.Remaining,
		)
	}

	if result.RetryAfter <= 0 {
		t.Fatalf(
			"expected positive retry delay, got %v",
			result.RetryAfter,
		)
	}
}

func TestEvaluateReturnsPolicyNotFound(t *testing.T) {
	resolver := &mockResolver{
		id: "api-key:unknown",
	}

	policyStore := store.NewInMemoryPolicyRepository()
	manager := limiter.NewManager()

	rateLimiter := service.NewRateLimitService(
		resolver,
		policyStore,
		manager,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	_, err := rateLimiter.Evaluate(req)

	if !errors.Is(err, store.ErrPolicyNotFound) {
		t.Fatalf(
			"expected ErrPolicyNotFound, got %v",
			err,
		)
	}
}
