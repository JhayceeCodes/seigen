package model

import "time"

type Algorithm string

const (
	TokenBucket          Algorithm = "token_bucket"
	LeakyBucket          Algorithm = "leaky_bucket"
	FixedWindow          Algorithm = "fixed_window"
	SlidingWindowLog     Algorithm = "sliding_window_log"
	SlidingWindowCounter Algorithm = "sliding_window_counter"
)

type LimiterConfig struct {
	Algorithm Algorithm
	Config    Config
}

type Config interface {
	isLimiterConfig()
}

type WindowConfig struct {
	Limit  int
	Window time.Duration
}

func (WindowConfig) isLimiterConfig() {}

type TokenBucketConfig struct {
	Capacity       int
	RefillInterval time.Duration
	RefillAmount   int
}

func (TokenBucketConfig) isLimiterConfig() {}

type LeakyBucketConfig struct {
	Capacity     int
	LeakInterval time.Duration
}

func (LeakyBucketConfig) isLimiterConfig() {}
