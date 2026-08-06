package limiter_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JhayceeCodes/rate-limiter-gateway/internal/limiter"
)

func TestNewFixedWindowStartsWithZeroRequests(t *testing.T) {
	window := limiter.NewFixedWindow(
		5,
		time.Second,
	)

	if window.Requests() != 0 {
		t.Fatalf("expected zero requests, got %d", window.Requests())
	}
}

func TestNewFixedWindowStartsWithFullCapacity(t *testing.T) {
	window := limiter.NewFixedWindow(
		10,
		time.Second,
	)
	if window.Remaining() != 10 {
		t.Errorf(
			"expected remaining %d, got %d",
			10,
			window.Remaining(),
		)
	}

}

func TestFixedWindowAllowDecrementsRemaining(t *testing.T) {
	limit := 10
	window := limiter.NewFixedWindow(
		limit,
		time.Second,
	)
	window.Allow()
	window.Allow()

	if window.Remaining() != limit-2 {
		t.Errorf("expected decreased limit, got %d", window.Remaining())
	}

	if window.Requests() != 2 {
		t.Errorf("expected 2 requests, got %d", window.Requests())
	}
}

func TestFixedWindowAllowRejectsWhenLimitReached(t *testing.T) {
	window := limiter.NewFixedWindow(
		2,
		time.Second,
	)

	if !window.Allow() {
		t.Fatal("expected first request to succeed")
	}

	if !window.Allow() {
		t.Fatal("expected second request to succeed")
	}

	if window.Allow() {
		t.Fatal("expected third request to fail")
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

	if !window.Allow() {
		t.Error("expected window reset")
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

			if window.Allow() {
				successful.Add(1)
			}
		})
	}

	wg.Wait()

	if successful.Load() != 5 {
		t.Errorf("expected 5 successful requests, got %d", successful.Load())
	}
}
