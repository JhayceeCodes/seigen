package limiter

import (
	"sync"
	"time"
)

type SlidingWindowCounter struct {
	limit         int
	window        time.Duration
	currentCount  int
	previousCount int
	currentStart  time.Time
	mu            sync.Mutex
}

func NewSlidingWindowCounter(limit int, window time.Duration) *SlidingWindowCounter {
	if limit <= 0 {
		panic("capacity cannot be less than zero")
	}

	if window <= 0 {
		panic("window must be greater than zero")
	}

	return &SlidingWindowCounter{
		limit:        limit,
		window:       window,
		currentStart: time.Now(),
	}
}

// advance rolls the window forward if the current fixed window has expired.
func (swc *SlidingWindowCounter) advance() {
	elapsed := time.Since(swc.currentStart)

	windowsPassed := int(elapsed / swc.window)

	if windowsPassed <= 0 {
		return
	} else if windowsPassed == 1 {
		swc.previousCount = swc.currentCount
	} else {
		swc.previousCount = 0
	}

	swc.currentCount = 0
	swc.currentStart = swc.currentStart.Add(time.Duration(windowsPassed) * swc.window)

}

func (swc *SlidingWindowCounter) weightedCount() float64 {
	elapsed := time.Since(swc.currentStart)

	overlap := float64(swc.window-elapsed) / float64(swc.window)
	if overlap < 0 {
		overlap = 0
	}

	return float64(swc.previousCount)*overlap + float64(swc.currentCount)
}

func (swc *SlidingWindowCounter) Allow() LimiterResult {
	swc.mu.Lock()
	defer swc.mu.Unlock()

	swc.advance()

	weighted := swc.weightedCount()

	if weighted >= float64(swc.limit) {
		remaining := int(float64(swc.limit) - weighted)
		if remaining < 0 {
			remaining = 0
		}

		return LimiterResult{
			Allowed:    false,
			Remaining:  remaining,
			RetryAfter: swc.retryAfter(),
		}
	}

	swc.currentCount++
	remaining := int(float64(swc.limit) - swc.weightedCount())

	return LimiterResult{
		Allowed:    true,
		Remaining:  remaining,
		RetryAfter: 0,
	}
}

// Requests returns the estimated number of active requests
// in the current sliding window.
func (swc *SlidingWindowCounter) Requests() int {
	swc.mu.Lock()
	defer swc.mu.Unlock()

	swc.advance()

	return int(swc.weightedCount())
}

// Remaining returns the estimated remaining request capacity
// in the current sliding window.
func (swc *SlidingWindowCounter) Remaining() int {
	swc.mu.Lock()
	defer swc.mu.Unlock()

	swc.advance()

	remaining := float64(swc.limit) - swc.weightedCount()
	if remaining < 0 {
		return 0
	}

	return int(remaining)
}

func (swc *SlidingWindowCounter) retryAfter() time.Duration {
	elapsed := time.Since(swc.currentStart)

	if swc.previousCount == 0 {
		return swc.timeUntilWindowReset()
	}

	targetOverlap :=
		float64(swc.limit-swc.currentCount) /
			float64(swc.previousCount)

	if targetOverlap <= 0 {
		return swc.timeUntilWindowReset()
	}

	currentOverlap :=
		float64(swc.window-elapsed) /
			float64(swc.window)

	if currentOverlap <= targetOverlap {
		return 0
	}

	wait := (currentOverlap - targetOverlap) *
		float64(swc.window)

	return time.Duration(wait)
}

func (swc *SlidingWindowCounter) timeUntilWindowReset() time.Duration {
	remaining := swc.window - time.Since(swc.currentStart)

	if remaining < 0 {
		return 0
	}

	return remaining
}
