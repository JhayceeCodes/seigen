package service

import (
	"errors"
	"net/http"

	"github.com/JhayceeCodes/seigen/identifier"
	"github.com/JhayceeCodes/seigen/limiter"
	"github.com/JhayceeCodes/seigen/model"
	"github.com/JhayceeCodes/seigen/store"
)

type RateLimitService struct {
	resolver         identifier.IdentifierResolver
	policyStore      store.PolicyRepository
	groupMemberStore store.PolicyGroupMemberRepository
	manager          *limiter.Manager
}

// RateLimitResult contains the result of a rate-limit evaluation.
type RateLimitResult struct {
	limiter.LimiterResult
	Limit int
}


// NewRateLimitService creates a rate-limit service using the provided
// identifier resolver, policy repositories, and limiter manager.
func NewRateLimitService(
	resolver identifier.IdentifierResolver,
	policyStore store.PolicyRepository,
	groupMemberStore store.PolicyGroupMemberRepository,
	manager *limiter.Manager,
) *RateLimitService {
	return &RateLimitService{
		resolver:         resolver,
		policyStore:      policyStore,
		groupMemberStore: groupMemberStore,
		manager:          manager,
	}
}

// Evaluate resolves the request's identifier and evaluates it against the
// applicable rate-limiting policy.
//
// An individual policy takes precedence over a policy group. If no individual
// policy exists, the identifier's policy group is used when the identifier is
// a member of one.
func (r *RateLimitService) Evaluate(req *http.Request) (RateLimitResult, error) {
	id, err := r.resolver.Resolve(req)
	if err != nil {
		return RateLimitResult{}, err
	}

	policy, err := r.policyStore.Get(id)
	if err == nil {
		return r.evaluate(id, policy.Limiter)
	}

	if !errors.Is(err, store.ErrPolicyNotFound) {
		return RateLimitResult{}, err
	}

	if r.groupMemberStore == nil {
		return RateLimitResult{}, store.ErrPolicyNotFound
	}

	group, err := r.groupMemberStore.GetGroup(id)
	if err != nil {
		return RateLimitResult{}, err
	}

	return r.evaluate(id, group.Limiter)
}

func (r *RateLimitService) evaluate(id model.Identifier, config model.LimiterConfig) (RateLimitResult, error) {
	lim, err := r.manager.GetOrCreate(id, config)
	if err != nil {
		return RateLimitResult{}, err
	}

	return RateLimitResult{
		LimiterResult: lim.Allow(),
		Limit:         getLimit(config),
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
