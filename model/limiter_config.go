package model

import (
	"errors"
	"time"
)

type Algorithm string

// Supported rate-limiting algorithms.
const (
	TokenBucket          Algorithm = "token_bucket"
	LeakyBucket          Algorithm = "leaky_bucket"
	FixedWindow          Algorithm = "fixed_window"
	SlidingWindowLog     Algorithm = "sliding_window_log"
	SlidingWindowCounter Algorithm = "sliding_window_counter"
)

// LimiterConfig defines the algorithm and configuration used by a limiter.
type LimiterConfig struct {
	Algorithm Algorithm `json:"algorithm"`
	Config    Config    `json:"comfiguration"`
}

type Config interface {
	isLimiterConfig()
}

type WindowConfig struct {
	Limit  int           `json:"limit"`
	Window time.Duration `json:"window"`
}

type TokenBucketConfig struct {
	Capacity       int           `json:"capacity"`
	RefillInterval time.Duration `json:"refill_interval"`
	RefillAmount   int           `json:"refill_amount"`
}

type LeakyBucketConfig struct {
	Capacity     int           `json:"capacity"`
	LeakInterval time.Duration `json:"leak_interval"`
}

func (WindowConfig) isLimiterConfig() {}

func (TokenBucketConfig) isLimiterConfig() {}

func (LeakyBucketConfig) isLimiterConfig() {}

func (c LimiterConfig) Validate() error {
	switch c.Algorithm {
	case TokenBucket:
		config, ok := c.Config.(TokenBucketConfig)
		if !ok {
			return errors.New("token bucket requires TokenBucketConfig")
		}

		if config.Capacity <= 0 {
			return errors.New("token bucket capacity must be greater than zero")
		}

		if config.RefillInterval <= 0 {
			return errors.New("token bucket refill interval must be greater than zero")
		}

		if config.RefillAmount <= 0 {
			return errors.New("token bucket refill amount must be greater than zero")
		}

	case LeakyBucket:
		config, ok := c.Config.(LeakyBucketConfig)
		if !ok {
			return errors.New("leaky bucket requires LeakyBucketConfig")
		}

		if config.Capacity <= 0 {
			return errors.New("leaky bucket capacity must be greater than zero")
		}

		if config.LeakInterval <= 0 {
			return errors.New("leaky bucket leak interval must be greater than zero")
		}

	case FixedWindow, SlidingWindowLog, SlidingWindowCounter:
		config, ok := c.Config.(WindowConfig)
		if !ok {
			return errors.New("window-based limiter requires WindowConfig")
		}

		if config.Limit <= 0 {
			return errors.New("window limit must be greater than zero")
		}

		if config.Window <= 0 {
			return errors.New("window duration must be greater than zero")
		}

	default:
		return errors.New("unknown limiter algorithm")
	}

	return nil
}
