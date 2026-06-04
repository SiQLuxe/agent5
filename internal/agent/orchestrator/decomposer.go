package orchestrator

type DecomposeStrategy func(task *Task) ([]*Task, error)

type Decomposer struct {
	strategies map[string]DecomposeStrategy
}

func NewDecomposer() *Decomposer {
	return &Decomposer{
		strategies: make(map[string]DecomposeStrategy),
	}
}

func (d *Decomposer) Register(taskType TaskType, strategy DecomposeStrategy) {
	d.strategies[string(taskType)] = strategy
}

func (d *Decomposer) Decompose(task *Task) ([]*Task, error) {
	strategy, ok := d.strategies[string(task.Type)]
	if !ok {
		return []*Task{task}, nil
	}
	return strategy(task)
}

func DefaultSequentialStrategy(task *Task) ([]*Task, error) {
	return []*Task{
		{ID: task.ID + "-analyze", Type: TaskAnalyze, Content: "Analyze: " + task.Content, Status: StatusPending},
		{ID: task.ID + "-design", Type: TaskDesign, Content: "Design: based on analysis", Status: StatusPending},
		{ID: task.ID + "-code", Type: TaskCode, Content: "Implement: based on design", Status: StatusPending},
		{ID: task.ID + "-review", Type: TaskReview, Content: "Review: verify implementation", Status: StatusPending},
	}, nil
}
