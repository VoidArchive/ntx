package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/anish/ntx/internal/config"
	"github.com/anish/ntx/internal/logger"
	"github.com/anish/ntx/internal/delivery/tui"
)

type App struct {
	cfg    *config.Config
	logger logger.Logger
}

func NewApp(cfg *config.Config, logger logger.Logger) *App {
	return &App{
		cfg:    cfg,
		logger: logger,
	}
}

func (a *App) Run(ctx context.Context) error {
	rootCmd := &cobra.Command{
		Use:   "ntx",
		Short: "Nepal Stock Exchange Terminal",
		Long:  "A terminal-based stock market tracker for Nepal Stock Exchange (NEPSE)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runTUI(ctx)
		},
	}

	// Add flags
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file path")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "enable debug mode")
	rootCmd.PersistentFlags().StringP("log-level", "l", "info", "log level (debug, info, warn, error)")

	// Add subcommands
	rootCmd.AddCommand(
		newVersionCmd(),
		newConfigCmd(),
		newImportCmd(),
	)

	return rootCmd.ExecuteContext(ctx)
}

func (a *App) runTUI(ctx context.Context) error {
	app := tui.NewApp(a.cfg, a.logger)
	return app.Run(ctx)
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("ntx version %s\n", Version)
			fmt.Printf("Built: %s\n", BuildTime)
			fmt.Printf("Commit: %s\n", GitCommit)
		},
	}
}

func newConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Config commands: init, validate, show")
		},
	}
}

func newImportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "import",
		Short: "Import portfolio data from CSV",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement CSV import functionality
			fmt.Println("CSV import functionality not yet implemented")
			return nil
		},
	}
}