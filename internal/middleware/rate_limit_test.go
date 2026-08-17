package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JhayceeCodes/seigen/internal/limiter"
)

// fakeLimiter is a test double for limiter.Limiter.
// It lets us control whether Allow() returns true or false.
type fakeLimiter struct {
	allow bool
}

func (f fakeLimiter) Allow() bool {
	return f.allow
}

func TestMiddlewarePreventsDownstreamExecution(t *testing.T) {
	called := false

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	mw := RateLimit(fakeLimiter{allow: false}, handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)

	if called {
		t.Error("expected downstream handler NOT to be called when limiter denies request")
	}

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected status %d, got %d", http.StatusTooManyRequests, rec.Code)
	}
}

func TestMiddlewareAllowsDownstreamExecution(t *testing.T) {
	called := false

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	mw := RateLimit(fakeLimiter{allow: true}, handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)

	if !called {
		t.Error("expected downstream handler to be called when limiter allows request")
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestMiddlewareWorksWithTokenBucket(t *testing.T) {
	handlerCalls := 0

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalls++
		w.WriteHeader(http.StatusOK)
	})

	mw := RateLimit(limiter.NewTokenBucket(3, time.Second, 1), handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	var codes []int

	for range 4 {
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, req)
		codes = append(codes, rec.Code)
	}

	for i := range 3 {
		if codes[i] != http.StatusOK {
			t.Errorf("expected status ok, got %d", codes[i])
		}
	}

	if handlerCalls != 3 {
		t.Errorf("expected handler to be called 3 times, got %d", handlerCalls)
	}

	if codes[3] != http.StatusTooManyRequests {
		t.Errorf("expected status too many requests, got %d", codes[3])
	}
}
