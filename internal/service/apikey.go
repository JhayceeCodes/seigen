package service

import (
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/JhayceeCodes/rate-limiter-gateway/internal/model"
)

func NewAPIKey(tier string) (model.APIKey, error) {
	apiTier := strings.TrimSpace(strings.ToLower(tier))

	if apiTier == "" {
		return model.APIKey{}, fmt.Errorf("tier is required")
	}

	if apiTier != model.TierFree && apiTier != model.TierPremium {
		return model.APIKey{}, fmt.Errorf("tier must either be free or premium")
	}

	key := make([]byte, 16)

	_, err := rand.Read((key))
	if err != nil {
		return model.APIKey{}, err
	}

	apiKey := fmt.Sprintf("rlg-%x", key)

	return model.APIKey{
		Key:  apiKey,
		Tier: apiTier,
	}, nil
}
