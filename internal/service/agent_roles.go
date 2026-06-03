package service

import (
	"context"
	"fmt"

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

type PlanningAgent struct {
	aiAssistant *AIAssistant
}

func NewPlanningAgent(ai *AIAssistant) *PlanningAgent {
	return &PlanningAgent{aiAssistant: ai}
}

func (p *PlanningAgent) Execute(task *Task) (string, error) {
	prompt := fmt.Sprintf("作为规划Agent，分析并规划以下任务：\n\n%s\n\n请输出详细的执行步骤。", task.Content)
	return p.aiAssistant.Chat(task.ID, prompt)
}

func (p *PlanningAgent) GetRoleName() string {
	return "规划Agent"
}

func (p *PlanningAgent) GetSupportedTaskTypes() []TaskType {
	return []TaskType{TaskAnalyze, TaskDesign}
}

type CodingAgent struct {
	aiAssistant *AIAssistant
}

func NewCodingAgent(ai *AIAssistant) *CodingAgent {
	return &CodingAgent{aiAssistant: ai}
}

func (c *CodingAgent) Execute(task *Task) (string, error) {
	prompt := fmt.Sprintf("作为编码Agent，根据需求编写Go代码：\n\n%s\n\n请输出完整的代码实现。", task.Content)
	return c.aiAssistant.Chat(task.ID, prompt)
}

func (c *CodingAgent) GetRoleName() string {
	return "编码Agent"
}

func (c *CodingAgent) GetSupportedTaskTypes() []TaskType {
	return []TaskType{TaskCode}
}

type ReviewAgent struct {
	aiAssistant *AIAssistant
}

func NewReviewAgent(ai *AIAssistant) *ReviewAgent {
	return &ReviewAgent{aiAssistant: ai}
}

func (r *ReviewAgent) Execute(task *Task) (string, error) {
	prompt := fmt.Sprintf("作为审查Agent，审查以下代码：\n\n%s\n\n请指出潜在问题和优化建议。", task.Content)
	return r.aiAssistant.Chat(task.ID, prompt)
}

func (r *ReviewAgent) GetRoleName() string {
	return "审查Agent"
}

func (r *ReviewAgent) GetSupportedTaskTypes() []TaskType {
	return []TaskType{TaskReview}
}

type ExecutionAgent struct {
	debugger *DebuggerService
}

func NewExecutionAgent(db *DebuggerService) *ExecutionAgent {
	return &ExecutionAgent{debugger: db}
}

func (e *ExecutionAgent) Execute(task *Task) (string, error) {
	return e.debugger.ExecuteGoCode(task.Content)
}

func (e *ExecutionAgent) GetRoleName() string {
	return "执行Agent"
}

func (e *ExecutionAgent) GetSupportedTaskTypes() []TaskType {
	return []TaskType{TaskExecute}
}