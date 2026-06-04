package orchestrator

type TaskType string

const (
	TaskAnalyze  TaskType = "task_analyze"
	TaskDesign   TaskType = "task_design"
	TaskCode     TaskType = "task_code"
	TaskReview   TaskType = "task_review"
	TaskExecute  TaskType = "task_execute"
	TaskResearch TaskType = "task_research"
	TaskCustom   TaskType = "task_custom"
)

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID      string
	Type    TaskType
	Content string
	Status  TaskStatus
	AgentID string
	Result  string
	Error   string
}
