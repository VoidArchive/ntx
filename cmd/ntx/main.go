/**
 * NTX Portfolio Management TUI - Main Entry Point
 *
 * This file serves as the primary application entry point for the NTX (NEPSE Power Terminal)
 * Portfolio Management TUI. It initializes the application and starts the main program flow.
 *
 * The application is built using the Bubbletea framework for TUI functionality and follows
 * the Model-View-Update pattern. This Phase 1 implementation establishes the foundation
 * for future portfolio management features.
 */

package main

import (
	"fmt"
	"os"

	"ntx/internal/app"

	tea "github.com/charmbracelet/bubbletea"
)

// main initializes and starts the NTX Portfolio Management TUI application
// This creates a new Bubbletea application with the main model and starts the TUI
func main() {
	// Create the main application model following Bubbletea's Model-View-Update pattern
	model := app.NewModel()

	// Initialize the Bubbletea program with standard input/output
	// This starts the TUI event loop and handles user interaction
	program := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	// Start the TUI application and handle any errors
	// The application will run until the user quits (q key or Ctrl+C)
	if _, err := program.Run(); err != nil {
		fmt.Printf("Error running NTX Portfolio Management TUI: %v\n", err)
		os.Exit(1)
	}
}
