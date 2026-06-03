package backend

import (
	"context"
	"sync"
)

type Registry struct {
	mu       sync.RWMutex
	backends map[AgentType]AgentBackend
}

func NewRegistry() *Registry {
	return &Registry{
		backends: make(map[AgentType]AgentBackend),
	}
}

func (r *Registry) Register(t AgentType, b AgentBackend) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.backends[t] = b
}

func (r *Registry) Get(t AgentType) (AgentBackend, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.backends[t]
	return b, ok
}

func (r *Registry) GetAll() []AgentBackend {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]AgentBackend, 0, len(r.backends))
	for _, b := range r.backends {
		result = append(result, b)
	}
	return result
}

func (r *Registry) ActiveBackends() []AgentBackend {
	all := r.GetAll()
	active := make([]AgentBackend, 0, len(all))
	for _, b := range all {
		h, err := b.Health(context.Background())
		if err == nil && h != nil && h.Healthy {
			active = append(active, b)
		}
	}
	return active
}
