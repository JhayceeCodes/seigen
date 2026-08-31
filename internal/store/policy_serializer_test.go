package store

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/JhayceeCodes/seigen/internal/model"
)

func TestSerializeLimiterConfigTokenBucket(t *testing.T) {
	config := model.LimiterConfig{
		Algorithm: model.TokenBucket,
		Config: model.TokenBucketConfig{
			Capacity:       5,
			RefillInterval: time.Second,
			RefillAmount:   1,
		},
	}

	data, err := serializeLimiterConfig(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got struct {
		Algorithm     model.Algorithm `json:"algorithm"`
		Configuration struct {
			Capacity       int           `json:"capacity"`
			RefillInterval time.Duration `json:"refill_interval"`
			RefillAmount   int           `json:"refill_amount"`
		} `json:"configuration"`
	}

	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to unmarshal serialized config: %v", err)
	}

	if got.Algorithm != model.TokenBucket {
		t.Errorf(
			"expected algorithm %q, got %q",
			model.TokenBucket,
			got.Algorithm,
		)
	}

	if got.Configuration.Capacity != 5 {
		t.Errorf(
			"expected capacity 5, got %d",
			got.Configuration.Capacity,
		)
	}

	if got.Configuration.RefillInterval != time.Second {
		t.Errorf(
			"expected refill interval %v, got %v",
			time.Second,
			got.Configuration.RefillInterval,
		)
	}

	if got.Configuration.RefillAmount != 1 {
		t.Errorf(
			"expected refill amount 1, got %d",
			got.Configuration.RefillAmount,
		)
	}
}

func TestSerializeLimiterConfigLeakyBucket(t *testing.T) {
	config := model.LimiterConfig{
		Algorithm: model.LeakyBucket,
		Config: model.LeakyBucketConfig{
			Capacity:     5,
			LeakInterval: time.Second,
		},
	}

	data, err := serializeLimiterConfig(config)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	var got struct {
		Algorithm     model.Algorithm `json:"algorithm"`
		Configuration struct {
			Capacity     int           `json:"capacity"`
			LeakInterval time.Duration `json:"leak_interval"`
		} `json:"configuration"`
	}

	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to unmarshal serialized config: %v", err)
	}

	if got.Algorithm != model.LeakyBucket {
		t.Errorf(
			"expected algorithm %q, got %q",
			model.LeakyBucket,
			got.Algorithm,
		)
	}

	if got.Configuration.Capacity != 5 {
		t.Errorf(
			"expected capacity 5, got %d",
			got.Configuration.Capacity,
		)
	}

	if got.Configuration.LeakInterval != time.Second {
		t.Errorf(
			"expected leak interval %v, got %v",
			time.Second,
			got.Configuration.LeakInterval,
		)
	}

}

func TestSerializeLimiterConfigWindow(t *testing.T) {
	config := model.LimiterConfig{
		Algorithm: model.FixedWindow,
		Config: model.WindowConfig{
			Limit:  100,
			Window: time.Minute,
		},
	}

	data, err := serializeLimiterConfig(config)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	var got struct {
		Algorithm     model.Algorithm `json:"algorithm"`
		Configuration struct {
			Limit  int           `json:"limit"`
			Window time.Duration `json:"window"`
		} `json:"configuration"`
	}

	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to unmarshal serialized config: %v", err)
	}

	if got.Algorithm != model.FixedWindow {
		t.Errorf(
			"expected algorithm %q, got %q",
			model.FixedWindow,
			got.Algorithm,
		)
	}

	if got.Configuration.Limit != 100 {
		t.Errorf(
			"expected capacity 100, got %d",
			got.Configuration.Limit,
		)
	}

	if got.Configuration.Window != time.Minute {
		t.Errorf(
			"expected leak interval %v, got %v",
			time.Minute,
			got.Configuration.Window,
		)
	}
}
