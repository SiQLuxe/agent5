package service

import "sync"

type CollaborationManager struct {
	orchestrator *TaskOrchestrator
	mu           sync.RWMutex
}

func NewCollaborationManager(orchestrator *TaskOrchestrator) *CollaborationManager {
	return &CollaborationManager{orchestrator: orchestrator}
}

func (cm *CollaborationManager) ExecuteWorkflow(steps []*Task) ([]*Task, error) {
	var results []*Task

	for _, step := range steps {
		err := cm.orchestrator.Dispatch(step)
		if err != nil {
			return results, err
		}
		results = append(results, step)
	}

	return results, nil
}

func (cm *CollaborationManager) AutoDecompose(task *Task) ([]*Task, error) {
	switch task.Type {
	case TaskAnalyze:
		return cm.decomposeAnalysis(task)
	case TaskDesign:
		return cm.decomposeDesign(task)
	default:
		return []*Task{task}, nil
	}
}

func (cm *CollaborationManager) decomposeAnalysis(task *Task) ([]*Task, error) {
	return []*Task{
		{ID: task.ID + "-1", Type: TaskAnalyze, Content: "需求分析: " + task.Content, Status: StatusPending},
		{ID: task.ID + "-2", Type: TaskDesign, Content: "架构设计: 基于需求分析进行架构设计", Status: StatusPending},
		{ID: task.ID + "-3", Type: TaskCode, Content: "代码实现: 根据设计编写代码", Status: StatusPending},
		{ID: task.ID + "-4", Type: TaskReview, Content: "代码审查: 审查实现代码", Status: StatusPending},
	}, nil
}

func (cm *CollaborationManager) decomposeDesign(task *Task) ([]*Task, error) {
	return []*Task{
		{ID: task.ID + "-1", Type: TaskDesign, Content: "详细设计: " + task.Content, Status: StatusPending},
		{ID: task.ID + "-2", Type: TaskCode, Content: "代码实现: 根据设计文档编写代码", Status: StatusPending},
		{ID: task.ID + "-3", Type: TaskReview, Content: "设计审查: 验证实现是否符合设计", Status: StatusPending},
		{ID: task.ID + "-4", Type: TaskExecute, Content: "测试执行: 运行相关测试", Status: StatusPending},
	}, nil
}

func (cm *CollaborationManager) ExecuteWithAutoDecompose(task *Task) ([]*Task, error) {
	steps, err := cm.AutoDecompose(task)
	if err != nil {
		return nil, err
	}

	return cm.ExecuteWorkflow(steps)
}