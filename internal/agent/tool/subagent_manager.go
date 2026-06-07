package tool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type SubAgentConfig struct {
	Name         string
	SystemPrompt string
	Model        string
	MaxDepth     int
}

type SubAgent struct {
	ID        string
	Name      string
	Config    SubAgentConfig
	Status    string
	Result    string
	Error     string
	CreatedAt time.Time
	ctx       context.Context
	cancel    context.CancelFunc
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
		if a.Status == "running" {
			running++
		}
	}
	if running >= m.max {
		return &SubAgent{
			ID:     uuid.New().String(),
			Name:   cfg.Name,
			Config: cfg,
			Status: "failed",
			Error:  fmt.Sprintf("max concurrent agents reached (%d)", m.max),
		}
	}

	childCtx, cancel := context.WithCancel(ctx)
	sa := &SubAgent{
		ID:        uuid.New().String(),
		Name:      cfg.Name,
		Config:    cfg,
		Status:    "running",
		CreatedAt: time.Now(),
		ctx:       childCtx,
		cancel:    cancel,
	}
	m.agents[sa.ID] = sa

	go func() {
		<-childCtx.Done()
		m.mu.Lock()
		if a, ok := m.agents[sa.ID]; ok && a.Status == "running" {
			a.Status = "cancelled"
		}
		m.mu.Unlock()
	}()

	return sa
}

func (m *SubagentManager) Get(id string) *SubAgent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.agents[id]
}

func (m *SubagentManager) List() []*SubAgent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*SubAgent, 0, len(m.agents))
	for _, a := range m.agents {
		out = append(out, a)
	}
	return out
}

func (m *SubagentManager) Cancel(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if a, ok := m.agents[id]; ok {
		a.cancel()
		a.Status = "cancelled"
	}
}

func (m *SubagentManager) IsCancelled(id string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if a, ok := m.agents[id]; ok {
		if a.Status == "cancelled" {
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

func (m *SubagentManager) RunningCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, a := range m.agents {
		if a.Status == "running" {
			count++
		}
	}
	return count
}
