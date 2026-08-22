package limiter_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JhayceeCodes/seigen/internal/limiter"
)

func TestNewSlidingWindowCounterStartsWithZeroRequests(t *testing.T) {
	counter := limiter.NewSlidingWindowCounter(5, time.Second)

	if got := counter.Requests(); got != 0 {
		t.Fatalf("expected zero requests, got %d", got)
	}
}

func TestSlidingWindowCounterStartsFull(t *testing.T) {
	limit := 5

	counter := limiter.NewSlidingWindowCounter(limit, time.Second)

	if got := counter.Remaining(); got != limit {
		t.Errorf("expected %d remaining, got %d", limit, got)
	}
}

func TestSlidingWindowCounterAllowRecordsRequests(t *testing.T) {
	counter := limiter.NewSlidingWindowCounter(5, time.Second)

	result := counter.Allow()

	if !result.Allowed {
		t.Fatal("expected request to be allowed")
	}

	if result.Remaining != 4 {
		t.Fatalf("expected 4 remaining, got %d", result.Remaining)
	}

	if result.RetryAfter != 0 {
		t.Fatalf("expected no retry delay, got %v", result.RetryAfter)
	}

	if got := counter.Requests(); got != 1 {
		t.Fatalf("expected 1 request, got %d", got)
	}
}

func TestSlidingWindowCounterLimitIsEnforcedWithinWindow(t *testing.T) {
	counter := limiter.NewSlidingWindowCounter(2, time.Second)

	counter.Allow()
	counter.Allow()

	result := counter.Allow()

	if result.Allowed {
		t.Fatal("expected rejected request")
	}

	if result.Remaining != 0 {
		t.Errorf("expected 0 remaining, got %d", result.Remaining)
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

func TestSlidingWindowCounterAllowIsConcurrentSafe(t *testing.T) {
	counter := limiter.NewSlidingWindowCounter(5, time.Second)

	var successful atomic.Int32
	var wg sync.WaitGroup

	for range 100 {
		wg.Go(func() {
			if counter.Allow().Allowed {
				successful.Add(1)
			}
		})
	}

	wg.Wait()

	if got := successful.Load(); got != 5 {
		t.Errorf("expected 5 successful requests, got %d", got)
	}
}

func TestSlidingWindowCounterResetsAfterGapLongerThanOneWindow(t *testing.T) {
	window := 50 * time.Millisecond
	counter := limiter.NewSlidingWindowCounter(5, window)

	counter.Allow()
	counter.Allow()

	// Let more than one full window pass with no activity - the
	// previous window's count should NOT carry forward here, since
	// there's no overlap left with the current sliding range at all.
	time.Sleep(3 * window)

	if got := counter.Requests(); got != 0 {
		t.Fatalf("expected 0 requests after multi-window gap, got %d", got)
	}

	if got := counter.Remaining(); got != 5 {
		t.Fatalf("expected full capacity after multi-window gap, got %d", got)
	}
}

func TestSlidingWindowCounterWeightedCountDecaysOverTime(t *testing.T) {
	window := 300 * time.Millisecond
	counter := limiter.NewSlidingWindowCounter(10, window)

	for range 4 {
		counter.Allow()
	}

	// Cross into the next fixed window, but only slightly - most of
	// the previous window's weight should still be counted.
	time.Sleep(window + 20*time.Millisecond)

	early := counter.Requests()
	if early < 3 {
		t.Errorf("expected weighted count still close to 4 shortly after rollover, got %d", early)
	}

	// Let most of the current window elapse - the previous window's
	// overlap should have decayed close to zero by now.
	time.Sleep(window - 40*time.Millisecond)

	late := counter.Requests()
	if late > 1 {
		t.Errorf("expected weighted count close to 0 near end of window, got %d", late)
	}
}

func TestNewSlidingWindowCounterRejectsInvalidLimit(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	limiter.NewSlidingWindowCounter(-1, time.Second)
}

func TestNewSlidingWindowCounterRejectsInvalidWindow(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	limiter.NewSlidingWindowCounter(1, -1*time.Second)
}

func TestSlidingWindowCounterConcurrentReadsAndWritesAreRaceFree(t *testing.T) {
	counter := limiter.NewSlidingWindowCounter(50, time.Second)

	var wg sync.WaitGroup

	for range 50 {
		wg.Go(func() {
			counter.Allow()
		})
		wg.Go(func() {
			counter.Requests()
		})
		wg.Go(func() {
			counter.Remaining()
		})
	}

	wg.Wait()
}

func TestSlidingWindowCounterReturnsRetryAfter(t *testing.T) {
	window := 200 * time.Millisecond
	counter := limiter.NewSlidingWindowCounter(2, window)

	counter.Allow()
	counter.Allow()

	result := counter.Allow()

	if result.Allowed {
		t.Fatal("expected request to be rejected")
	}

	if result.RetryAfter <= 0 {
		t.Fatal("expected positive retry delay")
	}

	if result.RetryAfter > window {
		t.Errorf(
			"expected retry delay no greater than %v, got %v",
			window,
			result.RetryAfter,
		)
	}
}
