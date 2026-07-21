package store

import (
	"sync"

	"github.com/JhayceeCodes/rate-limiter-gateway/internal/model"
)

// APIKeyStore stores API keys in memory.
// It is safe for concurrent use.
type APIKeyStore struct {
	mu   sync.RWMutex
	keys map[string]model.APIKey
}

func NewAPIKeyStore() *APIKeyStore {
	return &APIKeyStore{
		keys: make(map[string]model.APIKey),
	}
}

func (s *APIKeyStore) Create(apikey model.APIKey) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.keys[apikey.Key] = apikey
}

func (s *APIKeyStore) Retrieve(key string) (model.APIKey, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	apiKey, ok := s.keys[key]
	return apiKey, ok
}

func (s *APIKeyStore) List() []model.APIKey {
	s.mu.RLock()
	defer s.mu.RUnlock()

	values := make([]model.APIKey, 0, len(s.keys))

	for _, value := range s.keys {
		values = append(values, value)
	}

	return values
}

func (s *APIKeyStore) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.keys[key]
	if !ok {
		return false
	}

	delete(s.keys, key)
	return true
}
