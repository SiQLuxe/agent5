package runtime

import (
	"github.com/example/agent-tui/internal/agent/session"
	"github.com/example/agent-tui/internal/agent/tool"
)

type AgentRunner struct {
	BaseConfig Config
	LLM        LLMClient
	Session    *session.Manager
}

func (r *AgentRunner) Run(sessionID, task, systemPrompt string, tools *tool.Registry) (string, error) {
	cfg := r.BaseConfig
	cfg.SystemPrompt = systemPrompt
	agent := NewAgent(cfg, tools, r.LLM, r.Session)
	return agent.Execute(sessionID, task)
}
