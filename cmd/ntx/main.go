package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/voidarchive/ntx/internal/delivery/tui"
	"github.com/voidarchive/ntx/internal/service/market"
)

func main() {
	// Initialize market service
	marketService := market.NewWithShareSansar()

	// Create TUI app
	app := tui.NewApp(marketService)

	// Setup Bubble Tea program
	p := tea.NewProgram(app, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		log.Fatal("Error running program:", err)
	}
}
