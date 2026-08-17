package model_test

import (
	"testing"
	"time"

	"github.com/JhayceeCodes/seigen/internal/model"
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
