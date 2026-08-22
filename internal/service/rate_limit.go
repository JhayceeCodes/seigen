package service

import (
	"net/http"

	"github.com/JhayceeCodes/seigen/internal/identifier"
	"github.com/JhayceeCodes/seigen/internal/limiter"
	"github.com/JhayceeCodes/seigen/internal/store"
)

type RateLimitService struct {
	resolver    identifier.IdentifierResolver
	policyStore *store.PolicyStore
	manager     *limiter.Manager
}

func NewRateLimitService(
	resolver identifier.IdentifierResolver,
	policyStore *store.PolicyStore,
	manager *limiter.Manager,
) *RateLimitService {
	return &RateLimitService{
		resolver:    resolver,
		policyStore: policyStore,
		manager:     manager,
	}
}

func (r *RateLimitService) Evaluate(req *http.Request) (limiter.LimiterResult, error) {
	id, err := r.resolver.Resolve(req)
	if err != nil {
		return limiter.LimiterResult{}, err
	}

	policy, err := r.policyStore.Get(id)
	if err != nil {
		return limiter.LimiterResult{}, err
	}

	lim, err := r.manager.Get(policy.Identifier, policy.Limiter)
	if err != nil {
		return limiter.LimiterResult{}, err
	}

	return lim.Allow(), nil
}