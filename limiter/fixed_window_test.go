package limiter_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JhayceeCodes/seigen/limiter"
)

func TestNewFixedWindowStartsWithZeroRequests(t *testing.T) {
	window := limiter.NewFixedWindow(
		5,
		time.Second,
	)

	if got := window.Requests(); got != 0 {
		t.Fatalf("expected zero requests, got %d", got)
	}
}

func TestNewFixedWindowStartsWithFullCapacity(t *testing.T) {
	window := limiter.NewFixedWindow(
		10,
		time.Second,
	)

	if got := window.Remaining(); got != 10 {
		t.Errorf("expected remaining 10, got %d", got)
	}
}

func TestFixedWindowAllowReturnsRemaining(t *testing.T) {
	window := limiter.NewFixedWindow(
		10,
		time.Second,
	)

	result := window.Allow()

	if !result.Allowed {
		t.Fatal("expected request to be allowed")
	}

	if result.Remaining != 9 {
		t.Errorf("expected 9 remaining requests, got %d", result.Remaining)
	}

	if result.RetryAfter != 0 {
		t.Errorf("expected no retry delay, got %v", result.RetryAfter)
	}

	if got := window.Requests(); got != 1 {
		t.Errorf("expected 1 request, got %d", got)
	}
}

func TestFixedWindowAllowRejectsWhenLimitReached(t *testing.T) {
	window := limiter.NewFixedWindow(
		2,
		time.Second,
	)

	first := window.Allow()

	if !first.Allowed {
		t.Fatal("expected first request to succeed")
	}

	if first.Remaining != 1 {
		t.Errorf("expected 1 remaining request, got %d", first.Remaining)
	}

	second := window.Allow()

	if !second.Allowed {
		t.Fatal("expected second request to succeed")
	}

	if second.Remaining != 0 {
		t.Errorf("expected 0 remaining requests, got %d", second.Remaining)
	}

	third := window.Allow()

	if third.Allowed {
		t.Fatal("expected third request to fail")
	}

	if third.Remaining != 0 {
		t.Errorf("expected 0 remaining requests, got %d", third.Remaining)
	}

	if third.RetryAfter <= 0 {
		t.Error("expected positive retry delay")
	}

	if third.RetryAfter > time.Second {
		t.Errorf(
			"expected retry delay to be at most %v, got %v",
			time.Second,
			third.RetryAfter,
		)
	}
}

func TestFixedWindowResetsAfterInterval(t *testing.T) {
	window := limiter.NewFixedWindow(
		2,
		time.Second,
	)

	window.Allow()
	window.Allow()

	time.Sleep(1100 * time.Millisecond)

	result := window.Allow()

	if !result.Allowed {
		t.Fatal("expected request to be allowed after window reset")
	}

	if result.Remaining != 1 {
		t.Errorf(
			"expected 1 remaining request after reset, got %d",
			result.Remaining,
		)
	}

	if result.RetryAfter != 0 {
		t.Errorf(
			"expected no retry delay after reset, got %v",
			result.RetryAfter,
		)
	}
}

func TestFixedWindowAllowIsConcurrentSafe(t *testing.T) {
	window := limiter.NewFixedWindow(
		5,
		time.Second,
	)

	var successful atomic.Int32
	var wg sync.WaitGroup

	for range 20 {
		wg.Go(func() {
			if window.Allow().Allowed {
				successful.Add(1)
			}
		})
	}

	wg.Wait()

	if got := successful.Load(); got != 5 {
		t.Errorf(
			"expected 5 successful requests, got %d",
			got,
		)
	}
}
