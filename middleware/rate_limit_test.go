package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JhayceeCodes/seigen/limiter"
	"github.com/JhayceeCodes/seigen/middleware"
	"github.com/JhayceeCodes/seigen/model"
	"github.com/JhayceeCodes/seigen/service"
	"github.com/JhayceeCodes/seigen/store"
)

type mockResolver struct {
	id  model.Identifier
	err error
}

func (m *mockResolver) Resolve(r *http.Request) (model.Identifier, error) {
	return m.id, m.err
}

func newTestRateLimitService(
	t *testing.T,
	policy model.Policy,
) *service.RateLimitService {
	t.Helper()

	policyStore := store.NewInMemoryPolicyRepository()

	if err := policyStore.Set(policy); err != nil {
		t.Fatalf("unexpected error setting policy: %v", err)
	}

	manager := limiter.NewManager()

	resolver := &mockResolver{
		id: policy.Identifier,
	}

	return service.NewRateLimitService(
		resolver,
		policyStore,
		manager,
	)
}

func TestRateLimitAllowsDownstreamExecution(t *testing.T) {
	policy := model.Policy{
		Identifier: "api-key:123",
		Limiter: model.LimiterConfig{
			Algorithm: model.TokenBucket,
			Config: model.TokenBucketConfig{
				Capacity:       3,
				RefillInterval: time.Second,
				RefillAmount:   1,
			},
		},
	}

	rateLimitService := newTestRateLimitService(t, policy)

	called := false

	handler := middleware.RateLimit(
		rateLimitService,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected downstream handler to be called")
	}

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	if got := rec.Header().Get("X-RateLimit-Limit"); got != "3" {
		t.Errorf("expected rate limit 3, got %q", got)
	}

	if got := rec.Header().Get("X-RateLimit-Remaining"); got != "2" {
		t.Errorf("expected 2 remaining, got %q", got)
	}

	if got := rec.Header().Get("Retry-After"); got != "" {
		t.Errorf("expected no Retry-After header, got %q", got)
	}
}

func TestRateLimitIncludesCorrectRemainingValue(t *testing.T) {
	policy := model.Policy{
		Identifier: "api-key:123",
		Limiter: model.LimiterConfig{
			Algorithm: model.TokenBucket,
			Config: model.TokenBucketConfig{
				Capacity:       3,
				RefillInterval: time.Second,
				RefillAmount:   1,
			},
		},
	}

	rateLimitService := newTestRateLimitService(t, policy)

	handler := middleware.RateLimit(
		rateLimitService,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	expectedRemaining := []string{"2", "1", "0"}

	for i, expected := range expectedRemaining {
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if got := rec.Header().Get("X-RateLimit-Remaining"); got != expected {
			t.Errorf(
				"request %d: expected %s remaining, got %s",
				i+1,
				expected,
				got,
			)
		}
	}
}

func TestRateLimitReturnsPolicyNotConfigured(t *testing.T) {
	policyStore := store.NewInMemoryPolicyRepository()
	manager := limiter.NewManager()

	resolver := &mockResolver{
		id: "api-key:unknown",
	}

	rateLimitService := service.NewRateLimitService(
		resolver,
		policyStore,
		manager,
	)

	called := false

	handler := middleware.RateLimit(
		rateLimitService,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if called {
		t.Fatal("expected downstream handler not to be called")
	}

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}

func TestRateLimitRejectsRateLimitedRequest(t *testing.T) {
	policy := model.Policy{
		Identifier: "api-key:123",
		Limiter: model.LimiterConfig{
			Algorithm: model.TokenBucket,
			Config: model.TokenBucketConfig{
				Capacity:       1,
				RefillInterval: time.Second,
				RefillAmount:   1,
			},
		},
	}

	rateLimitService := newTestRateLimitService(t, policy)

	called := false

	handler := middleware.RateLimit(
		rateLimitService,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// First request should consume the only token.
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected first request to reach downstream handler")
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected first request status %d, got %d",
			http.StatusOK, rec.Code)
	}

	// Reset the flag so we can specifically test the second request.
	called = false

	// Second request should be rejected.
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if called {
		t.Fatal("expected downstream handler not to be called for rejected request")
	}

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusTooManyRequests,
			rec.Code,
		)
	}

	if got := rec.Header().Get("X-RateLimit-Limit"); got != "1" {
		t.Errorf("expected rate limit 1, got %q", got)
	}

	if got := rec.Header().Get("X-RateLimit-Remaining"); got != "0" {
		t.Errorf("expected 0 remaining, got %q", got)
	}

	if got := rec.Header().Get("Retry-After"); got == "" {
		t.Error("expected Retry-After header")
	}
}

func TestRateLimitReturnsInternalServerErrorForUnexpectedError(t *testing.T) {
	policyStore := store.NewInMemoryPolicyRepository()
	manager := limiter.NewManager()

	resolver := &mockResolver{
		err: errors.New("resolver failure"),
	}

	rateLimitService := service.NewRateLimitService(
		resolver,
		policyStore,
		manager,
	)

	called := false

	handler := middleware.RateLimit(
		rateLimitService,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if called {
		t.Fatal("expected downstream handler not to be called")
	}

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}
