package store

import (
	"sync"

	"github.com/JhayceeCodes/rate-limiter-gateway/internal/model"
)

type APIKeyStore struct {
	mu   sync.RWMutex
	keys map[string]model.APIKey
}
