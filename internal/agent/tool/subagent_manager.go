package tool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type AgentStatus string

const (
	StatusRunning   AgentStatus = "running"
	StatusFailed    AgentStatus = "failed"
	StatusCancelled AgentStatus = "cancelled"
	StatusCompleted AgentStatus = "completed"
)

const (
	heartbeatInterval    = 30 * time.Second
	heartbeatTimeoutMult = 2
)

type SubAgentConfig struct {
	Name         string
	SystemPrompt string
	Model        string
	MaxDepth     int
}

type SubAgent struct {
	ID            string
	Name          string
	Config        SubAgentConfig
	Status        AgentStatus
	Result        string
	Error         string
	CreatedAt     time.Time
	lastHeartbeat time.Time
	ctx           context.Context
	cancel        context.CancelFunc
}

type SubagentManager struct {
	mu     sync.RWMutex
	agents map[string]*SubAgent
	max    int
}

func NewSubagentManager(max int) *SubagentManager {
	if max <= 0 {
		max = 5
	}
	return &SubagentManager{
		agents: make(map[string]*SubAgent),
		max:    max,
	}
}

func (m *SubagentManager) Spawn(ctx context.Context, cfg SubAgentConfig) *SubAgent {
	m.mu.Lock()
	defer m.mu.Unlock()

	running := 0
	for _, a := range m.agents {
		if a.Status == StatusRunning {
			running++
		}
	}
	if running >= m.max {
		return &SubAgent{
			ID:     uuid.New().String(),
			Name:   cfg.Name,
			Config: cfg,
			Status: StatusFailed,
			Error:  fmt.Sprintf("max concurrent agents reached (%d)", m.max),
		}
	}

	if cfg.MaxDepth <= 0 {
		return &SubAgent{
			ID:     uuid.New().String(),
			Name:   cfg.Name,
			Config: cfg,
			Status: StatusFailed,
			Error:  "max nesting depth reached",
		}
	}

	childCtx, cancel := context.WithCancel(ctx)
	sa := &SubAgent{
		ID:            uuid.New().String(),
		Name:          cfg.Name,
		Config:        cfg,
		Status:        StatusRunning,
		CreatedAt:     time.Now(),
		lastHeartbeat: time.Now(),
		ctx:           childCtx,
		cancel:        cancel,
	}
	m.agents[sa.ID] = sa

	go func() {
		ticker := time.NewTicker(heartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-childCtx.Done():
				return
			case <-ticker.C:
				m.mu.Lock()
				if a, ok := m.agents[sa.ID]; ok && a.Status == StatusRunning {
					if time.Since(a.lastHeartbeat) > heartbeatInterval*heartbeatTimeoutMult {
						a.Status = StatusFailed
						a.Error = "agent heartbeat timeout"
						a.cancel()
					}
				}
				m.mu.Unlock()
			}
		}
	}()
	return sa
}

func (m *SubagentManager) Get(id string) *SubAgent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if a, ok := m.agents[id]; ok {
		copy := *a
		return &copy
	}
	return nil
}

func (m *SubagentManager) List() []*SubAgent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*SubAgent, 0, len(m.agents))
	for _, a := range m.agents {
		copy := *a
		out = append(out, &copy)
	}
	return out
}

func (m *SubagentManager) Cancel(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if a, ok := m.agents[id]; ok {
		a.cancel()
		a.Status = StatusCancelled
	}
}

func (m *SubagentManager) IsCancelled(id string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if a, ok := m.agents[id]; ok {
		if a.Status == StatusCancelled {
			return true
		}
		select {
		case <-a.ctx.Done():
			return true
		default:
			return false
		}
	}
	return false
}

func (m *SubagentManager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if a, ok := m.agents[id]; ok {
		a.cancel()
		delete(m.agents, id)
	}
}

func (m *SubagentManager) RunningCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, a := range m.agents {
		if a.Status == StatusRunning {
			count++
		}
	}
	return count
}

func (m *SubagentManager) Complete(id, result string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if a, ok := m.agents[id]; ok {
		a.Status = StatusCompleted
		a.Result = result
		a.cancel()
	}
}

func (m *SubagentManager) Fail(id, errMsg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if a, ok := m.agents[id]; ok {
		a.Status = StatusFailed
		a.Error = errMsg
		a.cancel()
	}
}

func (m *SubagentManager) Heartbeat(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if a, ok := m.agents[id]; ok {
		a.lastHeartbeat = time.Now()
	}
}
