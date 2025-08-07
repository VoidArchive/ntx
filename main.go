package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/voidarchive/ntx/internal/delivery/tui"
	"github.com/voidarchive/ntx/internal/scraper"
	"github.com/voidarchive/ntx/internal/service/market"
)

func main() {
	svc := market.New(scraper.NewShareSansarScraper())

	p := tea.NewProgram(tui.New(svc), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
