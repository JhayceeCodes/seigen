package limiter_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JhayceeCodes/seigen/internal/limiter"
)

func TestNewSlidingWindowLogStartsWithZeroRequests(t *testing.T) {
	window := limiter.NewSlidingWindowLog(5, time.Second)

	if got := window.Requests(); got != 0 {
		t.Fatalf("expected zero requests, got %d", got)
	}
}

func TestSlidingWindowLogAllowIsConcurrentSafe(t *testing.T) {
	window := limiter.NewSlidingWindowLog(5, time.Second)

	var successful atomic.Int32
	var wg sync.WaitGroup

	for range 100 {
		wg.Go(func() {
			if window.Allow().Allowed {
				successful.Add(1)
			}
		})
	}

	wg.Wait()

	if got := successful.Load(); got != 5 {
		t.Errorf("expected 5 successful requests, got %d", got)
	}
}

func TestSlidingWindowLogStartsFull(t *testing.T) {
	limit := 5

	window := limiter.NewSlidingWindowLog(limit, time.Second)

	if got := window.Remaining(); got != limit {
		t.Errorf(
			"expected %d remaining, got %d",
			limit,
			got,
		)
	}
}

func TestSlidingWindowLogAllowRecordsRequests(t *testing.T) {
	window := limiter.NewSlidingWindowLog(5, time.Second)

	result := window.Allow()

	if !result.Allowed {
		t.Fatal("expected request to be allowed")
	}

	if result.Remaining != 4 {
		t.Errorf(
			"expected 4 remaining, got %d",
			result.Remaining,
		)
	}

	if result.RetryAfter != 0 {
		t.Errorf(
			"expected no retry delay, got %v",
			result.RetryAfter,
		)
	}

	if got := window.Requests(); got != 1 {
		t.Fatalf(
			"expected 1 request, got %d",
			got,
		)
	}
}

func TestSlidingWindowLogLimitIsEnforced(t *testing.T) {
	window := limiter.NewSlidingWindowLog(2, time.Second)

	window.Allow()
	window.Allow()

	result := window.Allow()

	if result.Allowed {
		t.Fatal("expected rejected request")
	}

	if result.Remaining != 0 {
		t.Errorf(
			"expected 0 remaining requests, got %d",
			result.Remaining,
		)
	}

	if result.RetryAfter <= 0 {
		t.Error("expected positive retry delay")
	}

	if result.RetryAfter > time.Second {
		t.Errorf(
			"expected retry delay no greater than %v, got %v",
			time.Second,
			result.RetryAfter,
		)
	}
}

func TestSlidingWindowLogExpiredRequestsAreRemoved(t *testing.T) {
	window := limiter.NewSlidingWindowLog(
		2,
		100*time.Millisecond,
	)

	window.Allow()

	time.Sleep(150 * time.Millisecond)

	result := window.Allow()

	if !result.Allowed {
		t.Fatal("expected request to be allowed after expiration")
	}

	if got := window.Requests(); got != 1 {
		t.Fatalf(
			"expected 1 active request after cleanup, got %d",
			got,
		)
	}
}

func TestSlidingWindowLogRemainingResetsAfterExpiration(t *testing.T) {
	window := limiter.NewSlidingWindowLog(
		2,
		100*time.Millisecond,
	)

	window.Allow()

	time.Sleep(150 * time.Millisecond)

	if got := window.Remaining(); got != 2 {
		t.Fatalf(
			"expected full remaining capacity, got %d",
			got,
		)
	}
}

func TestNewSlidingWindowLogRejectsInvalidLimit(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	limiter.NewSlidingWindowLog(-1, time.Second)
}

func TestNewSlidingWindowLogRejectsInvalidWindow(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	limiter.NewSlidingWindowLog(1, -1*time.Second)
}

func TestSlidingWindowLogExpiresOnlyOldRequests(t *testing.T) {
	window := limiter.NewSlidingWindowLog(
		2,
		100*time.Millisecond,
	)

	window.Allow()

	time.Sleep(60 * time.Millisecond)

	window.Allow()

	time.Sleep(60 * time.Millisecond)

	result := window.Allow()

	if !result.Allowed {
		t.Fatal("expected request to be allowed")
	}

	if result.Remaining != 0 {
		t.Errorf(
			"expected 0 remaining requests, got %d",
			result.Remaining,
		)
	}

	if got := window.Requests(); got != 2 {
		t.Fatalf(
			"expected 2 active requests, got %d",
			got,
		)
	}
}

func TestSlidingWindowLogRetryAfterTracksOldestRequest(t *testing.T) {
	window := limiter.NewSlidingWindowLog(
		2,
		200*time.Millisecond,
	)

	window.Allow()
	window.Allow()

	time.Sleep(50 * time.Millisecond)

	result := window.Allow()

	if result.Allowed {
		t.Fatal("expected request to be rejected")
	}

	if result.Remaining != 0 {
		t.Errorf("expected 0 remaining requests, got %d", result.Remaining)
	}

	if result.RetryAfter <= 0 {
		t.Fatal("expected positive retry delay")
	}

	if result.RetryAfter >= 200*time.Millisecond {
		t.Errorf(
			"expected retry delay to be less than %v, got %v",
			200*time.Millisecond,
			result.RetryAfter,
		)
	}
}
