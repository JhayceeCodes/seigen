package limiter_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JhayceeCodes/rate-limiter-gateway/internal/limiter"
)

func TestNewSlidingWindowStartsWithZeroRequests(t *testing.T) {
	window := limiter.NewSlidingWindow(5, time.Second)

	if window.Requests() != 0 {
		t.Fatalf("expected zero requests, got %d", window.Requests())
	}
}

func TestSlidingWindowAllowIsConcurrentSafe(t *testing.T) {
	window := limiter.NewSlidingWindow(5, time.Second)

	var successful atomic.Int32
	var wg sync.WaitGroup

	for range 100 {
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

func TestSlidingWindowStartsFull(t *testing.T) {
	limit := 5

	window := limiter.NewSlidingWindow(limit, time.Second)

	if window.Remaining() != limit {
		t.Errorf(
			"expected %d remaining, got %d",
			limit,
			window.Remaining(),
		)
	}
}

func TestSlidingWindowAllowRecordsRequests(t *testing.T) {
	window := limiter.NewSlidingWindow(5, time.Second)

	window.Allow()

	if window.Requests() != 1 {
		t.Fatalf(
			"expected 1 request, got %d",
			window.Requests(),
		)
	}

	if window.Remaining() != 4 {
		t.Fatalf(
			"expected 4 remaining, got %d",
			window.Remaining(),
		)
	}
}

func TestSlidingWindowLimitIsEnforced(t *testing.T) {
	window := limiter.NewSlidingWindow(2, time.Second)

	window.Allow()
	window.Allow()

	if window.Allow() {
		t.Error("expected rejected request")
	}
}

func TestSlidingWindowExpiredRequestsAreRemoved(t *testing.T) {
	window := limiter.NewSlidingWindow(2, 100*time.Millisecond)

	window.Allow()

	time.Sleep(150 * time.Millisecond)

	window.Allow()
	if window.Requests() != 1 {
		t.Fatalf(
			"expected 1 active request after cleanup, got %d",
			window.Requests(),
		)
	}
}

func TestSlidingWindowRemainingResetsAfterExpiration(t *testing.T) {
	window := limiter.NewSlidingWindow(2, 100*time.Millisecond)

	window.Allow()

	time.Sleep(150 * time.Millisecond)

	if window.Remaining() != 2 {
		t.Fatalf(
			"expected full remaining capacity, got %d",
			window.Remaining(),
		)
	}
}

func TestNewSlidingWindowRejectsInvalidLimit(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	window := limiter.NewSlidingWindow(-1, time.Second)

	window.Allow()
}

func TestNewSlidingWindowRejectsInvalidWindow(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	window := limiter.NewSlidingWindow(1, -1*time.Second)

	window.Allow()
}

func TestSlidingWindowExpiresOnlyOldRequests(t *testing.T) {
	window := limiter.NewSlidingWindow(2, 100*time.Millisecond)

	window.Allow()

	time.Sleep(60 * time.Millisecond)

	window.Allow()

	time.Sleep(60 * time.Millisecond)

	if !window.Allow() {
		t.Fatal("expected request to be allowed")
	}

	if window.Requests() != 2 {
		t.Fatalf(
			"expected 2 active requests, got %d",
			window.Requests(),
		)
	}
}
