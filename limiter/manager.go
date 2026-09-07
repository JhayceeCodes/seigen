package limiter

import (
	"sync"

	"github.com/JhayceeCodes/seigen/model"
)

type Instance struct {
	Policy  model.Policy
	Limiter Limiter
}

type Manager struct {
	mu        sync.Mutex
	instances map[model.Identifier]Instance
}

func NewManager() *Manager {
	return &Manager{
		instances: make(map[model.Identifier]Instance),
	}
}

// Get fetches an existing limiter instance or creates a new one.
func (m *Manager) GetOrCreate(
	identifier model.Identifier,
	config model.LimiterConfig,
) (Limiter, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	policy := model.Policy{
		Identifier: identifier,
		Limiter:    config,
	}

	if instance, ok := m.instances[identifier]; ok {
		if instance.Policy.Equal(policy) {
			return instance.Limiter, nil
		}
	}

	lim, err := New(config)
	if err != nil {
		return nil, err
	}

	m.instances[identifier] = Instance{
		Policy:  policy,
		Limiter: lim,
	}

	return lim, nil
}
