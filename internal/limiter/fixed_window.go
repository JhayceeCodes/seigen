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
	mu        sync.RWMutex
}

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

// First request starts the initial window.
func (fw *FixedWindow) Allow() bool {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	now := time.Now()

	if !now.Before(fw.resetTime) {
		fw.resetTime = now.Add(fw.window)
		fw.requests = 0
	}

	if fw.requests < fw.limit {
		fw.requests++
		return true
	}

	return false
}

func (fw *FixedWindow) Remaining() int {
	fw.mu.RLock()
	defer fw.mu.RUnlock()

	return fw.limit - fw.requests
}

func (fw *FixedWindow) Requests() int {
	fw.mu.RLock()
	defer fw.mu.RUnlock()

	return fw.requests
}
