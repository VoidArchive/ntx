/*
NTX Portfolio Management TUI - Main Entry Point

Bootstrap configuration prioritizes CLI flags > env vars > config file > defaults
to support both casual users and power users with complex setups.

Bubbletea's Model-View-Update pattern chosen for predictable state management
and excellent terminal event handling - critical for financial data accuracy.
*/

package main

import (
	"flag"
	"fmt"
	"ntx/internal/app"
	"ntx/internal/config"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
)

// Bootstrap handles configuration cascade and error recovery
// Early exit prevents corrupted financial data from invalid configs
func main() {
	// CLI flags override config file to support CI/CD and scripting
	var (
		themeFlag  = flag.String("theme", "", "Theme to use (tokyo_night, rose_pine, gruvbox, default)")
		configFlag = flag.String("config", "", "Path to config file")
	)
	flag.Parse()

	// Viper cascade ensures consistent config precedence across deployments
	if *themeFlag != "" {
		viper.Set("ui.theme", *themeFlag)
	}
	if *configFlag != "" {
		viper.SetConfigFile(*configFlag)
	}

	// Config validation prevents runtime errors with financial calculations
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Dependency injection pattern enables testing and configuration flexibility
	model := app.NewModelWithConfig(cfg)

	// Alt screen preserves user's terminal history during portfolio sessions
	// Mouse support enables modern interaction patterns for data exploration
	program := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	// Graceful error handling prevents data corruption during market sessions
	// NOTE: Application state persists across restarts for session continuity
	if _, err := program.Run(); err != nil {
		fmt.Printf("Error running NTX Portfolio Management TUI: %v\n", err)
		os.Exit(1)
	}
}
