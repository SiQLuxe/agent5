package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/agent-tui/internal/ai"
	"github.com/example/agent-tui/internal/agent/orchestrator"
	"github.com/example/agent-tui/internal/agent/runtime"
	"github.com/example/agent-tui/internal/agent/save"
	"github.com/example/agent-tui/internal/agent/tool"
	"github.com/example/agent-tui/internal/backend"
	_ "github.com/example/agent-tui/internal/backend/opencode"
	"github.com/example/agent-tui/internal/data/config"
	"github.com/example/agent-tui/internal/data/history"
	"github.com/example/agent-tui/internal/server"
	"github.com/example/agent-tui/internal/service"
	"github.com/example/agent-tui/internal/ui"
)

// aiLLMAdapter wraps ai.Client to implement runtime.LLMClient
type aiLLMAdapter struct {
	client ai.Client
	model  string
}

func (a *aiLLMAdapter) ChatWithTools(msgs []runtime.Message, tools []map[string]interface{}, model string) (*runtime.LLMResponse, error) {
	if model == "" {
		model = a.model
	}
	req := ai.ChatCompletionRequest{
		Model:    model,
		Messages: make([]ai.Message, len(msgs)),
	}
	for i, m := range msgs {
		req.Messages[i] = ai.Message{Role: m.Role, Content: m.Content}
	}
	resp, err := a.client.ChatCompletion(req)
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return &runtime.LLMResponse{Type: "final", Content: ""}, nil
	}
	return &runtime.LLMResponse{Type: "final", Content: resp.Choices[0].Message.Content}, nil
}

func (a *aiLLMAdapter) ChatWithToolsStream(msgs []runtime.Message, tools []map[string]interface{}, model string, onChunk func(string)) (*runtime.LLMResponse, error) {
	if model == "" {
		model = a.model
	}
	req := ai.ChatCompletionRequest{
		Model:    model,
		Messages: make([]ai.Message, len(msgs)),
		Stream:   true,
	}
	for i, m := range msgs {
		req.Messages[i] = ai.Message{Role: m.Role, Content: m.Content}
	}
	var fullContent string
	err := a.client.ChatCompletionStream(req, func(chunk string) {
		fullContent += chunk
		onChunk(chunk)
	})
	if err != nil {
		return nil, err
	}
	return &runtime.LLMResponse{Type: "final", Content: fullContent}, nil
}

// aiLLMProvider wraps ai.Client to implement tool.LLMProvider
type aiLLMProvider struct {
	client ai.Client
}

func (p *aiLLMProvider) Chat(model, systemPrompt, userPrompt string) (string, error) {
	messages := []ai.Message{}
	if systemPrompt != "" {
		messages = append(messages, ai.Message{Role: "system", Content: systemPrompt})
	}
	messages = append(messages, ai.Message{Role: "user", Content: userPrompt})
	req := ai.ChatCompletionRequest{
		Model:    model,
		Messages: messages,
	}
	resp, err := p.client.ChatCompletion(req)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", nil
	}
	return resp.Choices[0].Message.Content, nil
}

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cfg, err := config.LoadConfig("configs/config.toml")
	if err != nil {
		cfg = config.GetDefaultConfig()
	}
	if cfg.DefaultClient == "" {
		cfg.DefaultClient = "local"
	}

	aiClient, err := ai.NewClientFromConfig(cfg)
	if err != nil {
		log.Fatalf("failed to create AI client: %v", err)
	}

	h := history.NewHistory("")
	aiAssistant := service.NewAIAssistant(aiClient, h)

	// Initialize agent system
	toolReg := tool.NewRegistry()
	toolReg.Register(&tool.ReadFileTool{})
	toolReg.Register(&tool.WriteFileTool{})
	toolReg.Register(&tool.SearchTextTool{})
	toolReg.Register(&tool.ExecCommandTool{})
	if aiClient != nil {
		toolReg.Register(&tool.ChatLLMTool{Provider: &aiLLMProvider{client: aiClient}})
	}

	skillRegistry := service.NewSkillRegistry()
	if err := service.LoadSkillsDir(skillRegistry, "skills"); err != nil {
		log.Printf("warning: loading skills: %v", err)
	}
	skillExecutor := service.NewSkillExecutor(skillRegistry, aiAssistant)

	app := ui.NewApp()
	app.SetAIAssistant(aiAssistant)
	app.SetSkillExecutor(skillExecutor)
	app.SetSkillRegistry(skillRegistry)
	app.SetSkillsDir("skills")

	approvalFn := func(toolName string, params map[string]interface{}, oldContent, newContent string) bool {
		path, _ := params["path"].(string)
		diffContent := save.UnifiedDiff(path, oldContent, newContent)

		modal := app.ApprovalModal()
		app.QueueUpdateDraw(func() {
			modal.SetContent(path, diffContent)
			app.ShowApproval()
		})

		result := modal.Wait()

		app.QueueUpdateDraw(func() {
			app.HideApproval()
		})

		return result
	}

	agentLLM := &aiLLMAdapter{client: aiClient, model: cfg.DefaultClient}
	agentReg := orchestrator.NewRegistry()
	for _, ac := range cfg.AgentRoles {
		if !ac.Enabled {
			continue
		}
		agentTools := tool.NewRegistry()
		for _, name := range ac.Tools {
			if t, ok := toolReg.Get(name); ok {
				agentTools.Register(t)
			}
		}
		agent := runtime.NewAgent(runtime.Config{
			Name:         ac.Name,
			Model:        ac.Model,
			SystemPrompt: ac.SystemPrompt,
			MaxReActLoop: ac.MaxReActLoop,
			SandboxDir:   ac.SandboxDir,
			ApprovalFn:   approvalFn,
		}, agentTools, agentLLM)
		agentReg.Register(ac.Name, agent,
			string(orchestrator.TaskExecute),
			string(orchestrator.TaskAnalyze),
			string(orchestrator.TaskDesign),
			string(orchestrator.TaskCode),
			string(orchestrator.TaskReview),
		)
	}
	orch := orchestrator.NewOrchestrator(agentReg, orchestrator.NewDecomposer(), orchestrator.NewMerger())
	app.SetOrchestrator(orch)

	// Sync skills into command registry
	app.CommandRegistry().SyncSkills(skillRegistry)

	// Initialize external agent backends
	backendRegistry := backend.NewRegistry()
	for name, bc := range cfg.Agent.Backends {
		if !bc.Enabled {
			continue
		}
		b, err := backend.NewBackend(bc.Type, backend.BackendConfig{
			Type:      backend.AgentType(bc.Type),
			Enabled:   bc.Enabled,
			AutoStart: bc.AutoStart,
			Binary:    bc.Binary,
			APIURL:    bc.APIURL,
			APIKey:    bc.APIKey,
		})
		if err != nil {
			log.Printf("failed to create backend %s: %v", name, err)
			continue
		}
		if err := b.Start(context.Background()); err != nil {
			log.Printf("failed to start backend %s: %v", name, err)
			continue
		}
		backendRegistry.Register(backend.AgentType(bc.Type), b)
		_ = service.NewExternalAgent(b)
	}

	// Start reverse-control server
	ctrlServer := server.New(app)
	go func() {
		if err := ctrlServer.Start(context.Background(), ":0"); err != nil {
			log.Printf("control server exited: %v", err)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, b := range backendRegistry.GetAll() {
			b.Stop(shutdownCtx)
		}
		ctrlServer.Stop(shutdownCtx)
		app.Stop()
	}()

	app.StartSkillWatcher()
	app.AddWelcomeMessage()

	if err := app.Run(); err != nil {
		log.Fatalf("application error: %v", err)
	}
}
