package service

import (
	"net/http"

	"github.com/JhayceeCodes/seigen/internal/identifier"
	"github.com/JhayceeCodes/seigen/internal/limiter"
	"github.com/JhayceeCodes/seigen/internal/model"
	"github.com/JhayceeCodes/seigen/internal/store"
)

type RateLimitService struct {
	resolver    identifier.IdentifierResolver
	policyStore store.PolicyRepository
	manager     *limiter.Manager
}

type RateLimitResult struct {
	limiter.LimiterResult
	Limit int
}

func NewRateLimitService(
	resolver identifier.IdentifierResolver,
	policyStore store.PolicyRepository,
	manager *limiter.Manager,
) *RateLimitService {
	return &RateLimitService{
		resolver:    resolver,
		policyStore: policyStore,
		manager:     manager,
	}
}

func (r *RateLimitService) Evaluate(req *http.Request) (RateLimitResult, error) {
	id, err := r.resolver.Resolve(req)
	if err != nil {
		return RateLimitResult{}, err
	}

	policy, err := r.policyStore.Get(id)
	if err != nil {
		return RateLimitResult{}, err
	}

	lim, err := r.manager.Get(policy.Identifier, policy.Limiter)
	if err != nil {
		return RateLimitResult{}, err
	}

	return RateLimitResult{
		LimiterResult: lim.Allow(),
		Limit:         getLimit(policy.Limiter),
	}, nil
}

func getLimit(config model.LimiterConfig) int {
	switch config.Algorithm {
	case model.TokenBucket:
		cfg := config.Config.(model.TokenBucketConfig)
		return cfg.Capacity

	case model.LeakyBucket:
		cfg := config.Config.(model.LeakyBucketConfig)
		return cfg.Capacity

	case model.FixedWindow,
		model.SlidingWindowLog,
		model.SlidingWindowCounter:
		cfg := config.Config.(model.WindowConfig)
		return cfg.Limit
	}

	return 0
}
