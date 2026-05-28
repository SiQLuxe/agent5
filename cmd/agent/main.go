package main

import (
	"github.com/example/agent-tui/internal/ai"
	"github.com/example/agent-tui/internal/data/config"
	"github.com/example/agent-tui/internal/data/history"
	"github.com/example/agent-tui/internal/service"
	"github.com/example/agent-tui/internal/ui"

	"charm.land/bubbletea/v2"
)

func main() {
	// Load config
	cfg, err := config.LoadConfig("configs/config.toml")
	if err != nil {
		cfg = config.GetDefaultConfig()
	}

	// Ensure default client is set
	if cfg.DefaultClient == "" {
		cfg.DefaultClient = "local"
	}

	// Create AI client using factory (dynamic selection by config.DefaultClient)
	aiClient, err := ai.NewClientFromConfig(cfg)
	if err != nil {
		panic(err)
	}

	// Create history and AI assistant
	h := history.NewHistory("")
	aiAssistant := service.NewAIAssistant(aiClient, h)

	// Create UI model
	model := ui.NewModel()
	model.SetAIAssistant(aiAssistant)

	// Welcome message
	model.AddChatMessage("system", "Hello! Welcome to the Agent TUI.")

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}