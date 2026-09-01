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
	lastLeak     time.Time
	mu           sync.Mutex
}

// NewLeakyBucket creates a leaky bucket rate limiter.
//
// capacity is the maximum number of queued requests.
//
// interval specifies how frequently one queued request is considered leaked.
func NewLeakyBucket(capacity int, interval time.Duration) *LeakyBucket {
	if capacity <= 0 {
		panic("capacity cannot be less than zero")
	}

	if interval <= 0 {
		panic("interval cannot be less than zero")
	}

	return &LeakyBucket{
		capacity:     capacity,
		leakInterval: interval,
		lastLeak:     time.Now(),
	}
}

func (lb *LeakyBucket) leak() {
	elapsed := time.Since(lb.lastLeak)
	leaked := int(elapsed / lb.leakInterval)

	if leaked <= 0 {
		return
	}

	if leaked >= len(lb.queue) {
		lb.queue = lb.queue[:0]
	} else {
		lb.queue = lb.queue[leaked:]
	}

	lb.lastLeak = lb.lastLeak.Add(
		time.Duration(leaked) * lb.leakInterval,
	)
}

func (lb *LeakyBucket) Allow() LimiterResult {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.leak()

	if len(lb.queue) >= lb.capacity {
		return LimiterResult{
			Allowed:    false,
			Remaining:  0,
			RetryAfter: lb.leakInterval,
		}
	}

	lb.queue = append(lb.queue, time.Now())

	return LimiterResult{
		Allowed:    true,
		Remaining:  lb.capacity - len(lb.queue),
		RetryAfter: 0,
	}
}

// Requests returns the number of active (unleaked) requests in the bucket.
func (lb *LeakyBucket) Requests() int {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.leak()

	return len(lb.queue)
}

// Remaining returns the remaining queue capacity in the bucket.
func (lb *LeakyBucket) Remaining() int {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.leak()

	return lb.capacity - len(lb.queue)
}