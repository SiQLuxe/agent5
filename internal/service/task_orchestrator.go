package service

import (
	"errors"
	"fmt"

	"github.com/example/agent-tui/internal/data/history"
)

type TaskType string
type TaskStatus string

const (
	TaskAnalyze    TaskType = "analyze"
	TaskDesign     TaskType = "design"
	TaskCode       TaskType = "code"
	TaskReview     TaskType = "review"
	TaskExecute    TaskType = "execute"
)

const (
	StatusPending    TaskStatus = "pending"
	StatusRunning    TaskStatus = "running"
	StatusCompleted  TaskStatus = "completed"
	StatusFailed     TaskStatus = "failed"
)

type Task struct {
	ID       string
	Type     TaskType
	Content  string
	Status   TaskStatus
	AgentID  string
	Result   string
	Error    string
}

type AgentRole interface {
	Execute(task *Task) (string, error)
	GetRoleName() string
	GetSupportedTaskTypes() []TaskType
}

type TaskOrchestrator struct {
	agents    map[TaskType]AgentRole
	taskQueue []*Task
	history   *history.History
}

func NewTaskOrchestrator(history *history.History) *TaskOrchestrator {
	return &TaskOrchestrator{
		agents:  make(map[TaskType]AgentRole),
		history: history,
	}
}

func (o *TaskOrchestrator) RegisterAgent(taskType TaskType, agent AgentRole) {
	o.agents[taskType] = agent
}

func (o *TaskOrchestrator) Dispatch(task *Task) error {
	agent, ok := o.agents[task.Type]
	if !ok {
		return errors.New("no agent available for task type")
	}

	task.Status = StatusRunning

	result, err := agent.Execute(task)
	if err != nil {
		task.Status = StatusFailed
		task.Error = err.Error()
		return err
	}

	task.Status = StatusCompleted
	task.Result = result

	o.history.AddMessage(task.ID, "system", fmt.Sprintf("Task %s completed: %s", task.Type, result))
	return nil
}

func (o *TaskOrchestrator) GetTasks() []*Task {
	return o.taskQueue
}

func (o *TaskOrchestrator) AddTask(task *Task) {
	o.taskQueue = append(o.taskQueue, task)
}

func (o *TaskOrchestrator) GetAgentForTask(taskType TaskType) (AgentRole, bool) {
	agent, ok := o.agents[taskType]
	return agent, ok
}