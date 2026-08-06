package limiter

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewLeakyBucketStartsEmpty(t *testing.T) {
	bucket := NewLeakyBucket(
		10,
		5*time.Second,
	)

	if bucket.Requests() != 0 {
		t.Errorf("expected 0 requests, got %d", bucket.Requests())
	}

	if bucket.Remaining() != 10 {
		t.Errorf("expected 10 remaining requests, got %d", bucket.Remaining())
	}
}

func TestAllowAddsRequest(t *testing.T) {
	bucket := NewLeakyBucket(
		10,
		5*time.Second,
	)

	bucket.Allow()

	if bucket.Requests() != 1 {
		t.Errorf("expected 1 request, got %d", bucket.Requests())
	}

	if bucket.Remaining() != 9 {
		t.Errorf("expected 9 remaining requests, got %d", bucket.Remaining())
	}
}

func TestCapacityIsEnforced(t *testing.T) {
	bucket := NewLeakyBucket(
		2,
		5*time.Second,
	)

	bucket.Allow()
	bucket.Allow()

	if bucket.Allow() {
		t.Error("expected rejected request")
	}
}

func TestLeakRemovesOneRequest(t *testing.T) {
	bucket := NewLeakyBucket(
		5,
		time.Second,
	)

	bucket.Allow()
	bucket.Allow()
	bucket.Allow()
	bucket.Allow()
	bucket.Allow()

	time.Sleep(1100 * time.Millisecond)

	if bucket.Requests() != 4 {
		t.Errorf("expected 4 requests left, got %d", bucket.Requests())
	}
}

func TestMultipleLeaks(t *testing.T) {
	bucket := NewLeakyBucket(
		5,
		time.Second,
	)

	bucket.Allow()
	bucket.Allow()
	bucket.Allow()
	bucket.Allow()
	bucket.Allow()

	time.Sleep(3300 * time.Millisecond)

	if bucket.Requests() != 2 {
		t.Errorf("expected 2 requests left, got %d", bucket.Requests())
	}
}

func TestNewLeakyBucketRejectsInvalidLimit(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	NewLeakyBucket(-1, time.Second)

}

func TestNewLeakyBucketRejectsInvalidWindow(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	NewLeakyBucket(1, -1*time.Second)
}

func TestCanAllowAfterLeak(t *testing.T) {
	bucket := NewLeakyBucket(2, 100*time.Millisecond)

	bucket.Allow()
	bucket.Allow()

	if bucket.Allow() {
		t.Fatal("expected rejection")
	}

	time.Sleep(150 * time.Millisecond)

	if !bucket.Allow() {
		t.Fatal("expected request after leak")
	}
}

func TestLeakyBucketAllowIsConcurrentSafe(t *testing.T) {
	bucket := NewLeakyBucket(5, 100*time.Millisecond)

	var successful atomic.Int32
	var wg sync.WaitGroup

	for range 100 {
		wg.Go(func() {

			if bucket.Allow() {
				successful.Add(1)
			}
		})
	}

	wg.Wait()

	if successful.Load() != 5 {
		t.Errorf("expected 5 successful requests, got %d", successful.Load())
	}
}

func TestRemainingUpdatesAfterLeak(t *testing.T) {
	bucket := NewLeakyBucket(2, 100*time.Millisecond)

	bucket.Allow()
	bucket.Allow()

	if bucket.Remaining() != 0 {
		t.Errorf("expected zero remaining requests, got %d", bucket.Remaining())
	}

	time.Sleep(150 * time.Millisecond)

	if bucket.Remaining() != 1 {
		t.Errorf("expected one remaining request, got %d", bucket.Remaining())
	}

	if bucket.Requests() != 1 {
		t.Errorf("expected one allowed request, got %d", bucket.Requests())
	}
}
