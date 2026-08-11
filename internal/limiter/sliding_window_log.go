package limiter

import (
	"sync"
	"time"
)

type SlidingWindow struct {
	requests []time.Time
	limit    int
	window   time.Duration
	mu       sync.Mutex
}



// NewSlidingWindow creates a sliding window rate limiter.
//
// limit is the maximum number of requests allowed within the rolling time window.
//
// window specifies the duration over which requests are tracked.
func NewSlidingWindow(limit int, window time.Duration) *SlidingWindow {
	if limit <= 0 {
		panic("capacity cannot be less than zero")
	}

	if window <= 0 {
		panic("window must be greater than zero")
	}

	return &SlidingWindow{
		limit:  limit,
		window: window,
	}
}

// cleanUp removes expired timestamps.
//
// Caller must hold sw.mu.
func (sw *SlidingWindow) cleanUp() {
	currentTime := time.Now()

	cutoff := currentTime.Add(-sw.window)

	for len(sw.requests) > 0 {
		if sw.requests[0].Before(cutoff) {
			sw.requests = sw.requests[1:]
		} else {
			break
		}
	}
}

func (sw *SlidingWindow) Allow() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.cleanUp()

	if len(sw.requests) >= sw.limit {
		return false
	}

	sw.requests = append(sw.requests, time.Now())
	return true
}

// Requests returns the number of active requests
// after removing expired timestamps.
func (sw *SlidingWindow) Requests() int {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.cleanUp()

	return len(sw.requests)
}

// Remaining returns the remaining request capacity
// in the current sliding window.
func (sw *SlidingWindow) Remaining() int {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.cleanUp()

	return sw.limit - len(sw.requests)
}
