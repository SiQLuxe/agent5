package main

import (
	"log"

	"github.com/example/agent-tui/internal/ai"
	"github.com/example/agent-tui/internal/data/config"
	"github.com/example/agent-tui/internal/data/history"
	"github.com/example/agent-tui/internal/service"
	"github.com/example/agent-tui/internal/ui"
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
	skillRegistry := service.NewSkillRegistry()
	if err := service.LoadSkillsDir(skillRegistry, "skills"); err != nil {
		log.Printf("warning: loading skills: %v", err)
	}
	skillExecutor := service.NewSkillExecutor(skillRegistry, aiAssistant)

	app := ui.NewApp()
	app.SetAIAssistant(aiAssistant)
	app.SetSkillExecutor(skillExecutor)

	// Sync skills into command registry
	app.CommandRegistry().SyncSkills(skillRegistry)

	app.AddWelcomeMessage()

	if err := app.Run(); err != nil {
		log.Fatalf("application error: %v", err)
	}
}
