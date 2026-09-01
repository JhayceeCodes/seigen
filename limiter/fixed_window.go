package limiter

import (
	"sync"
	"time"
)

type FixedWindow struct {
	limit     int
	requests  int
	window    time.Duration
	resetTime time.Time
	mu        sync.Mutex
}

// NewFixedWindow creates a fixed window rate limiter.
//
// limit is the maximum number of requests allowed during each window.
//
// window specifies the duration of a single rate-limiting window.
func NewFixedWindow(limit int, window time.Duration) *FixedWindow {
	if limit <= 0 {
		panic("capacity cannot be less than zero")
	}

	if window <= 0 {
		panic("window must be greater than zero")
	}

	return &FixedWindow{
		limit:  limit,
		window: window,
	}
}

// maybeReset starts a new window if the current one has elapsed.
//
// Caller must hold fw.mu.
func (fw *FixedWindow) maybeReset() {
	now := time.Now()

	if !now.Before(fw.resetTime) {
		fw.resetTime = now.Add(fw.window)
		fw.requests = 0
	}
}

func (fw *FixedWindow) Allow() LimiterResult {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	fw.maybeReset()

	if fw.requests >= fw.limit {
		return LimiterResult{
			Allowed:    false,
			Remaining:  0,
			RetryAfter: time.Until(fw.resetTime),
		}
	}

	fw.requests++

	return LimiterResult{
		Allowed:    true,
		Remaining:  fw.limit - fw.requests,
		RetryAfter: 0,
	}
}

func (fw *FixedWindow) Remaining() int {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	fw.maybeReset()

	return fw.limit - fw.requests
}

func (fw *FixedWindow) Requests() int {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	fw.maybeReset()

	return fw.requests
}