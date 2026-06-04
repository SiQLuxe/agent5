package service

import (
	"context"

	"github.com/example/agent-tui/internal/backend"
)

type ExternalAgent struct {
	backend backend.AgentBackend
}

func NewExternalAgent(b backend.AgentBackend) *ExternalAgent {
	return &ExternalAgent{backend: b}
}

func (e *ExternalAgent) Execute(task *Task) (string, error) {
	result, err := e.backend.SendMessage(context.Background(), task.ID, &backend.Message{
		Role:    backend.RoleUser,
		Content: task.Content,
	})
	if err != nil {
		return "", err
	}
	return result.Content, nil
}

func (e *ExternalAgent) GetRoleName() string {
	return "外部Agent"
}

func (e *ExternalAgent) GetSupportedTaskTypes() []TaskType {
	return []TaskType{TaskExternal}
}