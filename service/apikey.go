package service

import (
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/JhayceeCodes/seigen/model"
)

func NewAPIKey(tier string) (model.APIKey, error) {
	apiTier := strings.TrimSpace(strings.ToLower(tier))

	if apiTier == "" {
		return model.APIKey{}, fmt.Errorf("tier is required")
	}

	key := make([]byte, 16)

	if _, err := rand.Read(key); err != nil {
		return model.APIKey{}, err
	}

	return model.APIKey{
		Key:  fmt.Sprintf("sgn-%x", key),
		Tier: apiTier,
	}, nil
}
