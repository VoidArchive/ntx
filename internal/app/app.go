package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"ntx/internal/security"
	"ntx/internal/ui/dashboard"
)

// App represents the main application
type App struct {
	config    *Config
	logger    *slog.Logger
	program   *tea.Program
	configDir string
}

// Config holds application configuration
type Config struct {
	RefreshInterval string
	StartupView     string
	LogLevel        string
	DataDir         string
}

// New creates a new application instance
func New() (*App, error) {
	// Determine config directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	configDir := filepath.Join(homeDir, ".ntx")

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Initialize security (credentials management)
	if err := security.InitializeCredentials(configDir); err != nil {
		return nil, fmt.Errorf("failed to initialize credentials: %w", err)
	}

	// Load configuration
	config, err := loadConfig(configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Setup logger
	logger := setupLogger(config.LogLevel)

	// Create the dashboard model
	model := dashboard.NewModel()

	// Create Bubbletea program
	program := tea.NewProgram(model, tea.WithAltScreen())

	app := &App{
		config:    config,
		logger:    logger,
		program:   program,
		configDir: configDir,
	}

	return app, nil
}

// Start starts the application
func (a *App) Start(ctx context.Context) error {
	a.logger.Info("Starting NTX - NEPSE Power Terminal", 
		"version", "0.1.0",
		"config_dir", a.configDir)

	// Start the TUI in a goroutine so we can handle context cancellation
	go func() {
		if _, err := a.program.Run(); err != nil {
			a.logger.Error("TUI program failed", "error", err)
		}
	}()

	return nil
}

// Stop stops the application gracefully
func (a *App) Stop() error {
	a.logger.Info("Stopping application...")
	
	if a.program != nil {
		a.program.Quit()
	}

	a.logger.Info("Application stopped successfully")
	return nil
}

// loadConfig loads the application configuration
func loadConfig(configDir string) (*Config, error) {
	// For now, return default configuration
	// TODO: Implement proper config loading from TOML file
	return &Config{
		RefreshInterval: "60s",
		StartupView:     "portfolio",
		LogLevel:        "info",
		DataDir:         configDir,
	}, nil
}

// setupLogger configures structured logging
func setupLogger(level string) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewTextHandler(os.Stderr, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}