package model

import (
	"testing"
	"time"
)

func TestLimiterConfigValidate(t *testing.T) {
	tests := []struct {
		name   string
		config LimiterConfig
	}{
		{
			name: "valid token bucket",
			config: LimiterConfig{
				Algorithm: TokenBucket,
				Config: TokenBucketConfig{
					Capacity:       10,
					RefillInterval: time.Second,
					RefillAmount:   2,
				},
			},
		},
		{
			name: "valid leaky bucket",
			config: LimiterConfig{
				Algorithm: LeakyBucket,
				Config: LeakyBucketConfig{
					Capacity:     10,
					LeakInterval: time.Second,
				},
			},
		},
		{
			name: "valid fixed window",
			config: LimiterConfig{
				Algorithm: FixedWindow,
				Config: WindowConfig{
					Limit:  10,
					Window: time.Minute,
				},
			},
		},
		{
			name: "valid sliding window log",
			config: LimiterConfig{
				Algorithm: SlidingWindowLog,
				Config: WindowConfig{
					Limit:  10,
					Window: time.Minute,
				},
			},
		},
		{
			name: "valid sliding window counter",
			config: LimiterConfig{
				Algorithm: SlidingWindowCounter,
				Config: WindowConfig{
					Limit:  10,
					Window: time.Minute,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); err != nil {
				t.Fatalf("expected valid config, got error: %v", err)
			}
		})
	}
}


func TestLimiterConfigValidateRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name   string
		config LimiterConfig
	}{
		{
			name: "token bucket invalid capacity",
			config: LimiterConfig{
				Algorithm: TokenBucket,
				Config: TokenBucketConfig{
					Capacity:       0,
					RefillInterval: time.Second,
					RefillAmount:   1,
				},
			},
		},
		{
			name: "token bucket invalid refill interval",
			config: LimiterConfig{
				Algorithm: TokenBucket,
				Config: TokenBucketConfig{
					Capacity:       10,
					RefillInterval: 0,
					RefillAmount:   1,
				},
			},
		},
		{
			name: "token bucket invalid refill amount",
			config: LimiterConfig{
				Algorithm: TokenBucket,
				Config: TokenBucketConfig{
					Capacity:       10,
					RefillInterval: time.Second,
					RefillAmount:   0,
				},
			},
		},
		{
			name: "fixed window invalid limit",
			config: LimiterConfig{
				Algorithm: FixedWindow,
				Config: WindowConfig{
					Limit:  0,
					Window: time.Minute,
				},
			},
		},
		{
			name: "fixed window invalid window",
			config: LimiterConfig{
				Algorithm: FixedWindow,
				Config: WindowConfig{
					Limit:  10,
					Window: 0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); err == nil {
				t.Fatalf("expected validation error")
			}
		})
	}
}



func TestLimiterConfigValidateRejectsWrongConfigType(t *testing.T) {
	tests := []struct {
		name   string
		config LimiterConfig
	}{
		{
			name: "token bucket with window config",
			config: LimiterConfig{
				Algorithm: TokenBucket,
				Config: WindowConfig{
					Limit:  10,
					Window: time.Minute,
				},
			},
		},
		{
			name: "leaky bucket with token bucket config",
			config: LimiterConfig{
				Algorithm: LeakyBucket,
				Config: TokenBucketConfig{
					Capacity:       10,
					RefillInterval: time.Second,
					RefillAmount:   1,
				},
			},
		},
		{
			name: "fixed window with token bucket config",
			config: LimiterConfig{
				Algorithm: FixedWindow,
				Config: TokenBucketConfig{
					Capacity:       10,
					RefillInterval: time.Second,
					RefillAmount:   1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}