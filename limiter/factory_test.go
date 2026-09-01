package limiter_test

import (
	"testing"
	"time"

	"github.com/JhayceeCodes/seigen/limiter"
	"github.com/JhayceeCodes/seigen/model"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config model.LimiterConfig
	}{
		{
			name: "token bucket",
			config: model.LimiterConfig{
				Algorithm: model.TokenBucket,
				Config: model.TokenBucketConfig{
					Capacity:       2,
					RefillInterval: time.Second,
					RefillAmount:   1,
				},
			},
		},
		{
			name: "leaky bucket",
			config: model.LimiterConfig{
				Algorithm: model.LeakyBucket,
				Config: model.LeakyBucketConfig{
					Capacity:     2,
					LeakInterval: time.Second,
				},
			},
		},
		{
			name: "fixed window",
			config: model.LimiterConfig{
				Algorithm: model.FixedWindow,
				Config: model.WindowConfig{
					Limit:  2,
					Window: time.Second,
				},
			},
		},
		{
			name: "sliding window log",
			config: model.LimiterConfig{
				Algorithm: model.SlidingWindowLog,
				Config: model.WindowConfig{
					Limit:  2,
					Window: time.Second,
				},
			},
		},
		{
			name: "sliding window counter",
			config: model.LimiterConfig{
				Algorithm: model.SlidingWindowCounter,
				Config: model.WindowConfig{
					Limit:  2,
					Window: time.Second,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := limiter.New(tt.config)

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if got == nil {
				t.Fatal("expected limiter, got nil")
			}
		})
	}
}

func TestNewRejectsInvalidConfig(t *testing.T) {
	config := model.LimiterConfig{
		Algorithm: model.TokenBucket,
		Config: model.WindowConfig{
			Limit:  10,
			Window: time.Minute,
		},
	}

	_, err := limiter.New(config)

	if err == nil {
		t.Fatal("expected invalid config to be rejected")
	}
}

func TestNewRejectsUnsupportedAlgorithm(t *testing.T) {
	config := model.LimiterConfig{
		Algorithm: model.Algorithm("something_random"),
		Config: model.WindowConfig{
			Limit:  10,
			Window: time.Minute,
		},
	}

	_, err := limiter.New(config)

	if err == nil {
		t.Fatal("expected unsupported algorithm to be rejected")
	}
}
