package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/agent-tui/internal/ai"
	"github.com/example/agent-tui/internal/backend"
	_ "github.com/example/agent-tui/internal/backend/opencode"
	"github.com/example/agent-tui/internal/data/config"
	"github.com/example/agent-tui/internal/data/history"
	"github.com/example/agent-tui/internal/server"
	"github.com/example/agent-tui/internal/service"
	"github.com/example/agent-tui/internal/ui"
)

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
