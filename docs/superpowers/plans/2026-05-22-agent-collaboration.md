# Agent 协作系统实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现一个结合分层协作与角色分工的 Agent 协作系统，支持任务调度、动态角色匹配和多 Agent 协作。

**Architecture:** 采用三层架构：任务调度层负责任务拆解和分配，Agent 角色层包含规划、编码、审查、执行四种角色，共享数据层提供会话历史和项目状态。

**Tech Stack:** Go 1.22+, Bubble Tea, 现有 AI 客户端模块

---

## 文件结构

```
internal/
└── service/
    ├── task_orchestrator.go          # 任务调度器
    ├── agent_roles.go                 # Agent 角色定义
    ├── collaboration_manager.go       # 协作流程管理器
    ├── task_orchestrator_test.go      # 调度器测试
    └── collaboration_manager_test.go  # 协作管理器测试
```

---

### Task 22: 任务调度器核心模块

**Files:**
- Create: `internal/service/task_orchestrator.go`
- Create: `internal/service/task_orchestrator_test.go`

- [x] **Step 1: 定义任务类型和状态枚举**
- [x] **Step 2: 实现任务调度器核心结构**
- [x] **Step 3: 实现任务分发方法**
- [x] **Step 4: 编写测试用例**
- [x] **Step 5: 运行测试验证**
- [x] **Step 6: Commit**

---

### Task 23: Agent 角色定义与匹配

**Files:**
- Create: `internal/service/agent_roles.go`
- Modify: `internal/service/task_orchestrator.go`

- [x] **Step 1: 定义 AgentRole 接口**
- [x] **Step 2: 实现规划 Agent**
- [x] **Step 3: 实现编码 Agent**
- [x] **Step 4: 实现审查 Agent**
- [x] **Step 5: 实现执行 Agent**
- [x] **Step 6: 编译验证**
- [x] **Step 7: Commit**

---

### Task 24: 协作流程管理

**Files:**
- Create: `internal/service/collaboration_manager.go`
- Create: `internal/service/collaboration_manager_test.go`

- [x] **Step 1: 定义协作流程管理器**
- [x] **Step 2: 实现多步骤协作流程**
- [x] **Step 3: 实现自动任务拆解**
- [x] **Step 4: 编写测试用例**
- [x] **Step 5: 运行测试验证**
- [x] **Step 6: Commit**

---

## Self-Review

1. **Spec coverage:** 已覆盖任务调度器、Agent角色定义、协作流程管理
2. **Placeholder scan:** 无占位符，所有步骤包含完整代码
3. **Type consistency:** 类型和方法签名保持一致

---

## 核心代码示例

### 任务调度器核心

```go
type TaskOrchestrator struct {
    agents    map[TaskType]AgentRole
    taskQueue []*Task
    history   *history.History
}

func (o *TaskOrchestrator) Dispatch(task *Task) error {
    agent, ok := o.agents[task.Type]
    if !ok {
        return errors.New("no agent available for task type")
    }
    // ... 执行任务
}
```

### Agent 角色接口

```go
type AgentRole interface {
    Execute(task *Task) (string, error)
    GetRoleName() string
    GetSupportedTaskTypes() []TaskType
}
```

### 协作流程管理

```go
func (cm *CollaborationManager) ExecuteWithAutoDecompose(task *Task) ([]*Task, error) {
    steps, err := cm.AutoDecompose(task)
    if err != nil {
        return nil, err
    }
    return cm.ExecuteWorkflow(steps)
}
```

---

## 测试结果

全部 8 个测试用例通过：
- TestTaskOrchestrator ✅
- TestTaskOrchestrator_NoAgent ✅
- TestTaskOrchestrator_AddTask ✅
- TestTaskOrchestrator_GetAgentForTask ✅
- TestCollaborationManager ✅
- TestCollaborationManager_ExecuteWorkflow ✅
- TestCollaborationManager_AutoDecomposeDesign ✅
- TestCollaborationManager_AutoDecomposeDefault ✅