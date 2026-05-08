package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/babything/agent/internal/agent"
	"github.com/babything/agent/internal/configfile"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx    context.Context
	client *agent.SignalingClient
	cfg    agent.Config
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Try to load config and auto-start
	cfg, err := agent.LoadConfig()
	if err == nil {
		a.cfg = cfg
		a.StartAgent()
	}
}

// GetConfig returns the current agent configuration
func (a *App) GetConfig() agent.Config {
	return a.cfg
}

// SaveConfig saves the configuration to file and restarts the agent if running
func (a *App) SaveConfig(cfg agent.Config) error {
	a.cfg = cfg

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = configfile.DefaultPath()
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := configfile.Save(configPath, cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Restart agent if running
	if a.client != nil {
		a.StopAgent()
		return a.StartAgent()
	}
	return nil
}

// StartAgent starts the signaling client
func (a *App) StartAgent() error {
	if a.client != nil {
		return nil // already running
	}

	client, err := agent.NewSignalingClient(a.cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	a.client = client
	go client.Run()
	log.Println("[desktop] agent started")
	return nil
}

// StopAgent stops the signaling client
func (a *App) StopAgent() error {
	if a.client == nil {
		return nil
	}
	a.client.Stop()
	a.client = nil
	log.Println("[desktop] agent stopped")
	return nil
}

// GetStatus returns the current agent status
func (a *App) GetStatus() string {
	if a.client == nil {
		return "stopped"
	}
	return "running"
}

// ShowWindow shows the settings window
func (a *App) ShowWindow() {
	runtime.WindowShow(a.ctx)
}

// HideWindow hides the settings window
func (a *App) HideWindow() {
	runtime.WindowHide(a.ctx)
}

// Shutdown is called on app shutdown
func (a *App) shutdown(ctx context.Context) {
	if a.client != nil {
		a.client.Stop()
	}
}
