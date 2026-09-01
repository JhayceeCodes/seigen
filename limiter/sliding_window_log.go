package limiter

import (
	"sync"
	"time"
)

type SlidingWindowLog struct {
	requests []time.Time
	limit    int
	window   time.Duration
	mu       sync.Mutex
}

// NewSlidingWindowLog creates a sliding window rate limiter.
//
// limit is the maximum number of requests allowed within the rolling time window.
//
// window specifies the duration over which requests are tracked.
func NewSlidingWindowLog(limit int, window time.Duration) *SlidingWindowLog {
	if limit <= 0 {
		panic("capacity cannot be less than zero")
	}

	if window <= 0 {
		panic("window must be greater than zero")
	}

	return &SlidingWindowLog{
		limit:  limit,
		window: window,
	}
}

// cleanUp removes expired timestamps.
//
// Caller must hold swl.mu.
func (swl *SlidingWindowLog) cleanUp() {
	currentTime := time.Now()
	cutoff := currentTime.Add(-swl.window)

	for len(swl.requests) > 0 {
		if swl.requests[0].Before(cutoff) {
			swl.requests = swl.requests[1:]
		} else {
			break
		}
	}
}

func (swl *SlidingWindowLog) Allow() LimiterResult {
	swl.mu.Lock()
	defer swl.mu.Unlock()

	swl.cleanUp()

	if len(swl.requests) >= swl.limit {
		retryAfter := time.Until(
			swl.requests[0].Add(swl.window),
		)

		return LimiterResult{
			Allowed:    false,
			Remaining:  0,
			RetryAfter: retryAfter,
		}
	}

	swl.requests = append(swl.requests, time.Now())

	return LimiterResult{
		Allowed:    true,
		Remaining:  swl.limit - len(swl.requests),
		RetryAfter: 0,
	}
}

// Requests returns the number of active requests
// after removing expired timestamps.
func (swl *SlidingWindowLog) Requests() int {
	swl.mu.Lock()
	defer swl.mu.Unlock()

	swl.cleanUp()

	return len(swl.requests)
}

// Remaining returns the remaining request capacity
// in the current sliding window.
func (swl *SlidingWindowLog) Remaining() int {
	swl.mu.Lock()
	defer swl.mu.Unlock()

	swl.cleanUp()

	return swl.limit - len(swl.requests)
}