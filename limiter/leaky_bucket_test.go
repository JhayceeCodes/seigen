package limiter_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JhayceeCodes/seigen/limiter"
)

func TestNewLeakyBucketStartsEmpty(t *testing.T) {
	bucket := limiter.NewLeakyBucket(
		10,
		5*time.Second,
	)

	if got := bucket.Requests(); got != 0 {
		t.Errorf("expected 0 requests, got %d", got)
	}

	if got := bucket.Remaining(); got != 10 {
		t.Errorf("expected 10 remaining requests, got %d", got)
	}
}

func TestAllowAddsRequest(t *testing.T) {
	bucket := limiter.NewLeakyBucket(
		10,
		5*time.Second,
	)

	result := bucket.Allow()

	if !result.Allowed {
		t.Fatal("expected request to be allowed")
	}

	if result.Remaining != 9 {
		t.Errorf("expected 9 remaining requests, got %d", result.Remaining)
	}

	if result.RetryAfter != 0 {
		t.Errorf("expected no retry delay, got %v", result.RetryAfter)
	}

	if got := bucket.Requests(); got != 1 {
		t.Errorf("expected 1 request, got %d", got)
	}
}

func TestCapacityIsEnforced(t *testing.T) {
	bucket := limiter.NewLeakyBucket(
		2,
		500*time.Millisecond,
	)

	bucket.Allow()
	bucket.Allow()

	result := bucket.Allow()

	if result.Allowed {
		t.Fatal("expected request to be rejected")
	}

	if result.Remaining != 0 {
		t.Errorf("expected 0 remaining requests, got %d", result.Remaining)
	}

	if result.RetryAfter != 500*time.Millisecond {
		t.Errorf(
			"expected retry after %v, got %v",
			5*time.Millisecond,
			result.RetryAfter,
		)
	}
}

func TestLeakRemovesOneRequest(t *testing.T) {
	bucket := limiter.NewLeakyBucket(
		5,
		time.Second,
	)

	for range 5 {
		bucket.Allow()
	}

	time.Sleep(1100 * time.Millisecond)

	if got := bucket.Requests(); got != 4 {
		t.Errorf("expected 4 requests left, got %d", got)
	}

	if got := bucket.Remaining(); got != 1 {
		t.Errorf("expected 1 remaining request, got %d", got)
	}
}

func TestMultipleLeaks(t *testing.T) {
	bucket := limiter.NewLeakyBucket(
		5,
		time.Second,
	)

	for range 5 {
		bucket.Allow()
	}

	time.Sleep(3300 * time.Millisecond)

	if got := bucket.Requests(); got != 2 {
		t.Errorf("expected 2 requests left, got %d", got)
	}

	if got := bucket.Remaining(); got != 3 {
		t.Errorf("expected 3 remaining requests, got %d", got)
	}
}

func TestNewLeakyBucketRejectsInvalidLimit(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	limiter.NewLeakyBucket(-1, time.Second)
}

func TestNewLeakyBucketRejectsInvalidWindow(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	limiter.NewLeakyBucket(1, -1*time.Second)
}

func TestCanAllowAfterLeak(t *testing.T) {
	bucket := limiter.NewLeakyBucket(
		2,
		100*time.Millisecond,
	)

	bucket.Allow()
	bucket.Allow()

	result := bucket.Allow()

	if result.Allowed {
		t.Fatal("expected rejection")
	}

	if result.RetryAfter != 100*time.Millisecond {
		t.Errorf(
			"expected retry after %v, got %v",
			100*time.Millisecond,
			result.RetryAfter,
		)
	}

	time.Sleep(150 * time.Millisecond)

	result = bucket.Allow()

	if !result.Allowed {
		t.Fatal("expected request after leak")
	}

	if result.Remaining != 0 {
		t.Errorf("expected 0 remaining requests, got %d", result.Remaining)
	}
}

func TestLeakyBucketAllowIsConcurrentSafe(t *testing.T) {
	bucket := limiter.NewLeakyBucket(
		5,
		100*time.Millisecond,
	)

	var successful atomic.Int32
	var wg sync.WaitGroup

	for range 100 {
		wg.Go(func() {
			if bucket.Allow().Allowed {
				successful.Add(1)
			}
		})
	}

	wg.Wait()

	if got := successful.Load(); got != 5 {
		t.Errorf("expected 5 successful requests, got %d", got)
	}
}

func TestRemainingUpdatesAfterLeak(t *testing.T) {
	bucket := limiter.NewLeakyBucket(
		2,
		100*time.Millisecond,
	)

	bucket.Allow()
	bucket.Allow()

	if got := bucket.Remaining(); got != 0 {
		t.Errorf("expected 0 remaining requests, got %d", got)
	}

	time.Sleep(150 * time.Millisecond)

	if got := bucket.Remaining(); got != 1 {
		t.Errorf("expected 1 remaining request, got %d", got)
	}

	if got := bucket.Requests(); got != 1 {
		t.Errorf("expected 1 active request, got %d", got)
	}
}
