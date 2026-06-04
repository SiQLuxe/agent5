package orchestrator

import "github.com/example/agent-tui/internal/agent/runtime"

type Registry struct {
	agents map[string]*runtime.Agent
	roles  map[string][]string
}

func NewRegistry() *Registry {
	return &Registry{
		agents: make(map[string]*runtime.Agent),
		roles:  make(map[string][]string),
	}
}

func (r *Registry) Register(name string, agent *runtime.Agent, roleTags ...string) {
	r.agents[name] = agent
	for _, tag := range roleTags {
		r.roles[tag] = append(r.roles[tag], name)
	}
}

func (r *Registry) Get(name string) (*runtime.Agent, bool) {
	a, ok := r.agents[name]
	return a, ok
}

func (r *Registry) List() []*runtime.Agent {
	list := make([]*runtime.Agent, 0, len(r.agents))
	for _, a := range r.agents {
		list = append(list, a)
	}
	return list
}

func (r *Registry) FindByRole(role string) []*runtime.Agent {
	names := r.roles[role]
	agents := make([]*runtime.Agent, 0, len(names))
	for _, n := range names {
		if a, ok := r.agents[n]; ok {
			agents = append(agents, a)
		}
	}
	return agents
}
