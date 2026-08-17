package service

import (
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/JhayceeCodes/seigen/internal/model"
)

func NewAPIKey(tier string) (model.APIKey, error) {
	apiTier := strings.TrimSpace(strings.ToLower(tier))

	if apiTier == "" {
		return model.APIKey{}, fmt.Errorf("tier is required")
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
