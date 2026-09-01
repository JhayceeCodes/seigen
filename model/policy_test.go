package model_test

import (
	"testing"
	"time"

	"github.com/JhayceeCodes/seigen/model"
)

func TestPolicyRejectsInvalidIdentifier(t *testing.T) {
	policy := model.Policy{
		Identifier: "",
		Limiter: model.LimiterConfig{
			Algorithm: model.TokenBucket,
			Config: model.TokenBucketConfig{
				Capacity:       10,
				RefillInterval: time.Second,
				RefillAmount:   1,
			},
		},
	}

	if err := policy.Validate(); err == nil {
		t.Fatal("expected rejection of invalid identifier")
	}
}

func TestPolicyCallsLimiterValidate(t *testing.T) {
	policy := model.Policy{
		Identifier: "user_id",
		Limiter: model.LimiterConfig{
			Algorithm: model.TokenBucket,
			Config: model.WindowConfig{
				Limit:  10,
				Window: time.Minute,
			},
		},
	}

	if err := policy.Validate(); err == nil {
		t.Fatal("expected validation error from limiter validate()")
	}
}


func TestPolicyAcceptsValidPolicy(t *testing.T) {
	policy := model.Policy{
		Identifier: "user:123",
		Limiter: model.LimiterConfig{
			Algorithm: model.TokenBucket,
			Config: model.TokenBucketConfig{
				Capacity:       100,
				RefillInterval: time.Second,
				RefillAmount:   10,
			},
		},
	}

	if err := policy.Validate(); err != nil {
		t.Fatalf("expected valid policy, got error: %v", err)
	}
}


func TestPolicyEqualOtherPolicy(t *testing.T) {
	policyA := model.Policy{
		Identifier: "user:123",
		Limiter: model.LimiterConfig{
			Algorithm: model.TokenBucket,
			Config: model.TokenBucketConfig{
				Capacity:       100,
				RefillInterval: time.Second,
				RefillAmount:   10,
			},
		},
	}

	policyB := model.Policy{
		Identifier: "user:123",
		Limiter: model.LimiterConfig{
			Algorithm: model.TokenBucket,
			Config: model.TokenBucketConfig{
				Capacity:       100,
				RefillInterval: time.Second,
				RefillAmount:   10,
			},
		},
	}

	if !policyA.Equal(policyB) {
		t.Fatal("expected equally configured policies to be similar")
	}
}


func TestPolicyNotEqualOtherPolicy(t *testing.T) {
	policyA := model.Policy{
		Identifier: "user:123",
		Limiter: model.LimiterConfig{
			Algorithm: model.TokenBucket,
			Config: model.TokenBucketConfig{
				Capacity:       100,
				RefillInterval: time.Second,
				RefillAmount:   10,
			},
		},
	}

	policyB := model.Policy{
		Identifier: "user:123",
		Limiter: model.LimiterConfig{
			Algorithm: model.TokenBucket,
			Config: model.TokenBucketConfig{
				Capacity:       200,
				RefillInterval: time.Second,
				RefillAmount:   10,
			},
		},
	}

	policyC := model.Policy{
		Identifier: "user:123",
		Limiter: model.LimiterConfig{
			Algorithm: model.FixedWindow,
			Config: model.WindowConfig{
				Limit:       200,
				Window: time.Second,
			},
		},
	}

	policyD := model.Policy{
		Identifier: "user:123",
		Limiter: model.LimiterConfig{
			Algorithm: model.TokenBucket,
			Config: model.TokenBucketConfig{
				Capacity:       100,
				RefillInterval: time.Second,
				RefillAmount:   5,
			},
		},
	}

	if policyA.Equal(policyB) {
		t.Fatal("expected policies to be different")
	}

	if policyB.Equal(policyC) {
		t.Fatal("expected policies of different algorithms to be unequal")
	}

	if policyA.Equal(policyD) {
		t.Fatal("expected token buckets of different refill amount to be unequal")
	}
}