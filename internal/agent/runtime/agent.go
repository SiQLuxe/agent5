package runtime

import (
	"github.com/example/agent-tui/internal/agent/tool"
)

type Config struct {
	Name         string
	Model        string
	SystemPrompt string
	MaxReActLoop int
	Temperature  float64
	ContextLimit int
	SandboxDir   string
	ApprovalFn   func(toolName string, params map[string]interface{}, oldContent, newContent string) bool
}

type Agent struct {
	Name   string
	Config Config
	Tools  *tool.Registry
	Memory *Memory
	Logger *Logger
	llm    LLMClient
}

func NewAgent(cfg Config, tools *tool.Registry, llm LLMClient) *Agent {
	if cfg.MaxReActLoop == 0 {
		cfg.MaxReActLoop = 20
	}
	if cfg.Temperature == 0 {
		cfg.Temperature = 0.7
	}
	if cfg.ContextLimit == 0 {
		cfg.ContextLimit = 50
	}
	return &Agent{
		Name:   cfg.Name,
		Config: cfg,
		Tools:  tools,
		Memory: NewMemory(),
		Logger: NewLogger(100),
		llm:    llm,
	}
}

func (a *Agent) Execute(task string) (string, error) {
	a.Memory.Clear()
	a.Logger.Clear()
	return a.reactLoop(task)
}

func (a *Agent) ExecuteStream(task string, onChunk func(string)) (string, error) {
	a.Memory.Clear()
	a.Logger.Clear()
	return a.reactLoopStream(task, onChunk)
}
