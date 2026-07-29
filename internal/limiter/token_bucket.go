package limiter

import "time"

type RateLimiter struct {
	tokens     chan struct{}
	refillTime time.Duration
}


func NewRateLimiter(capacity int, refillTime time.Duration) *RateLimiter {
	rl := &RateLimiter{
		tokens:     make(chan struct{}, capacity),
		refillTime: refillTime,
	}

	for range capacity {
		rl.tokens <- struct{}{}
	}

	go rl.startRefill()

	return rl
}

func (rl *RateLimiter) startRefill() {
	ticker := time.NewTicker(rl.refillTime)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case rl.tokens <- struct{}{}:
		default:
		}
	}
}

func (rl *RateLimiter) Allow() bool {
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}

}


