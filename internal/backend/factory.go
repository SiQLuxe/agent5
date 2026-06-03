package backend

import "fmt"

var backendBuilders = map[AgentType]func(BackendConfig) (AgentBackend, error){}

func RegisterBackendBuilder(t AgentType, fn func(BackendConfig) (AgentBackend, error)) {
	backendBuilders[t] = fn
}

func NewBackend(agentType string, cfg BackendConfig) (AgentBackend, error) {
	builder, ok := backendBuilders[AgentType(agentType)]
	if !ok {
		return nil, fmt.Errorf("unknown backend type: %s", agentType)
	}
	return builder(cfg)
}
