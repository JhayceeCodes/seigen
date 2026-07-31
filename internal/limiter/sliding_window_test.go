package limiter

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// construction starts with zero requests
// concurrency safety

func TestNewSlidingWindowStartsWithZeroRequests(t *testing.T) {
	window := NewSlidingWindow(
		5,
		time.Second,
	)

	if window.Requests() != 0 {
		t.Fatalf("expected zero requests, got %d", window.Requests())
	}
}

func TestSlidingWindowAllowIsConcurrentSafe(t *testing.T) {
	window := NewFixedWindow(
		5,
		time.Second,
	)

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
