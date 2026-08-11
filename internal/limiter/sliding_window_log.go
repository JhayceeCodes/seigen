package limiter

import (
	"sync"
	"time"
)

type SlidingWindowLog struct {
	requests []time.Time
	limit    int
	window   time.Duration
	mu       sync.RWMutex
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
// Caller must hold sw.mu.
func (swc *SlidingWindowLog) cleanUp() {
	currentTime := time.Now()

	cutoff := currentTime.Add(-swc.window)

	for len(swc.requests) > 0 {
		if swc.requests[0].Before(cutoff) {
			swc.requests = swc.requests[1:]
		} else {
			break
		}
	}
}

func (swc *SlidingWindowLog) Allow() bool {
	swc.mu.Lock()
	defer swc.mu.Unlock()

	swc.cleanUp()

	if len(swc.requests) >= swc.limit {
		return false
	}

	swc.requests = append(swc.requests, time.Now())
	return true
}

// Requests returns the number of active requests
// after removing expired timestamps.
func (sw *SlidingWindowLog) Requests() int {
	sw.mu.RLock()
	defer sw.mu.RUnlock()

	sw.cleanUp()

	return len(sw.requests)
}

// Remaining returns the remaining request capacity
// in the current sliding window.
func (swc *SlidingWindowLog) Remaining() int {
	swc.mu.RLock()
	defer swc.mu.RUnlock()

	swc.cleanUp()

	return swc.limit - len(swc.requests)
}
