package limiter

import (
	"sync"
	"time"
)

// TODO - make queue a request struct instead of plain timestamps
// type request struct {
// 	id        int
// 	timestamp time.Time
// }

type LeakyBucket struct {
	capacity     int
	queue        []time.Time
	leakInterval time.Duration
	mu           sync.RWMutex
}

func NewLeakyBucket(capacity int, interval time.Duration) *LeakyBucket {
	if capacity <= 0 {
		panic("capacity cannot be less than zero")
	}

	if interval <= 0 {
		panic("interval cannot be less than zero")
	}

	lb := &LeakyBucket{
		capacity:     capacity,
		leakInterval: interval,
	}

	// Start the background worker that leaks one queued request
	// every leakInterval.
	go lb.startLeak()

	return lb
}

func (lb *LeakyBucket) startLeak() {
	ticker := time.NewTicker(lb.leakInterval)
	defer ticker.Stop()

	for range ticker.C {
		lb.mu.Lock()

		if len(lb.queue) > 0 {
			lb.queue = lb.queue[1:]
		}

		lb.mu.Unlock()
	}
}

func (lb *LeakyBucket) Allow() bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if len(lb.queue) >= lb.capacity {
		return false
	}

	lb.queue = append(lb.queue, time.Now())
	return true
}

func (lb *LeakyBucket) Requests() int {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	return len(lb.queue)
}

func (lb *LeakyBucket) Remaining() int {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	return lb.capacity - len(lb.queue)
}
