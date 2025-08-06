package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gocolly/colly/v2"
)

// Quote represents a single stock quote
type Quote struct {
	Symbol    string
	Open      float64
	High      float64
	Low       float64
	LTP       float64 // Last Traded Price
	Volume    float64
	PrevClose float64
}

// Model holds our application state
type model struct {
	stocks       []string // List of stock symbols
	currentIndex int      // Which stock we're currently showing
	currentQuote *Quote   // Current stock's data
	loading      bool     // Are we loading data?
	err          error    // Any error that occurred
}

// Messages that can be sent to Update()
type (
	stockDataMsg *Quote
	errorMsg     error
)

// Initialize the model
func initialModel() model {
	return model{
		stocks:       []string{"NABIL", "AHPC", "ADBL", "API", "CHCL", "HIDCL"}, // Add more stocks here
		currentIndex: 0,
		loading:      true,
	}
}

// Init runs when the program starts
func (m model) Init() tea.Cmd {
	// Start by loading the first stock
	return loadStock(m.stocks[m.currentIndex])
}

// Update handles all the messages/events
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Quit the program
			return m, tea.Quit

		case "right", "l":
			// Go to next stock
			if m.currentIndex < len(m.stocks)-1 {
				m.currentIndex++
			} else {
				m.currentIndex = 0 // Wrap around to first stock
			}
			m.loading = true
			m.err = nil
			return m, loadStock(m.stocks[m.currentIndex])

		case "left", "h":
			// Go to previous stock
			if m.currentIndex > 0 {
				m.currentIndex--
			} else {
				m.currentIndex = len(m.stocks) - 1 // Wrap around to last stock
			}
			m.loading = true
			m.err = nil
			return m, loadStock(m.stocks[m.currentIndex])

		case "r":
			// Reload current stock
			m.loading = true
			m.err = nil
			return m, loadStock(m.stocks[m.currentIndex])
		}

	case stockDataMsg:
		// We received stock data
		m.currentQuote = msg
		m.loading = false
		return m, nil

	case errorMsg:
		// We received an error
		m.err = msg
		m.loading = false
		return m, nil
	}

	return m, nil
}

// View renders the current state
func (m model) View() string {
	var s strings.Builder

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")). // Bright blue
		Render("📈 NEPSE Stock Viewer")

	s.WriteString(title + "\n\n")

	// Stock navigation indicator
	navStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")) // Gray
	nav := fmt.Sprintf("Stock %d of %d", m.currentIndex+1, len(m.stocks))
	s.WriteString(navStyle.Render(nav) + "\n\n")

	// Loading state
	if m.loading {
		loading := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")). // Orange
			Render(fmt.Sprintf("Loading %s...", m.stocks[m.currentIndex]))
		s.WriteString(loading + "\n")
		return s.String()
	}

	// Error state
	if m.err != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n")
		return s.String()
	}

	// Stock data display
	if m.currentQuote != nil {
		s.WriteString(renderStockCard(m.currentQuote))
	}

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")). // Gray
		MarginTop(2)

	help := "← → : Navigate stocks  •  r: Reload  •  q: Quit"
	s.WriteString(helpStyle.Render(help))

	return s.String()
}

// Render a single stock's data as a nice card
func renderStockCard(quote *Quote) string {
	// Card style
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")). // Purple
		Padding(1, 2).
		Width(50)

	// Symbol style (big and bold)
	symbolStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("46")). // Green
		Align(lipgloss.Center)

	// Price style
	priceStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("226")). // Yellow
		Align(lipgloss.Center)

	// Calculate change
	change := quote.LTP - quote.PrevClose
	changePct := 0.0
	if quote.PrevClose > 0 {
		changePct = (change / quote.PrevClose) * 100
	}

	// Change color (green for positive, red for negative)
	changeColor := lipgloss.Color("196") // Red
	changeSymbol := "▼"
	if change >= 0 {
		changeColor = lipgloss.Color("46") // Green
		changeSymbol = "▲"
	}

	changeStyle := lipgloss.NewStyle().
		Foreground(changeColor).
		Align(lipgloss.Center)

	// Build the card content
	var content strings.Builder
	content.WriteString(symbolStyle.Render(quote.Symbol) + "\n\n")
	content.WriteString(priceStyle.Render(fmt.Sprintf("Rs %.2f", quote.LTP)) + "\n")
	content.WriteString(changeStyle.Render(fmt.Sprintf("%s %.2f (%.2f%%)", changeSymbol, change, changePct)) + "\n\n")

	// Data rows
	content.WriteString(fmt.Sprintf("Open:     Rs %.2f\n", quote.Open))
	content.WriteString(fmt.Sprintf("High:     Rs %.2f\n", quote.High))
	content.WriteString(fmt.Sprintf("Low:      Rs %.2f\n", quote.Low))
	content.WriteString(fmt.Sprintf("Volume:   %.0f\n", quote.Volume))
	content.WriteString(fmt.Sprintf("Prev:     Rs %.2f", quote.PrevClose))

	return cardStyle.Render(content.String())
}

// Command to load stock data (this returns a Cmd that Bubble Tea will execute)
func loadStock(symbol string) tea.Cmd {
	return func() tea.Msg {
		quote, err := scrapeQuote(symbol)
		if err != nil {
			return errorMsg(err)
		}
		return stockDataMsg(quote)
	}
}

// Your existing scraper function (slightly modified)
func scrapeQuote(symbol string) (*Quote, error) {
	var foundQuote *Quote

	c := colly.NewCollector(
		colly.AllowedDomains("www.sharesansar.com"),
	)

	c.UserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36"

	// Add a small delay to be nice to the server
	c.OnRequest(func(r *colly.Request) {
		time.Sleep(500 * time.Millisecond)
	})

	c.OnHTML("table tr", func(e *colly.HTMLElement) {
		if e.ChildText("td:nth-child(2)") == "" {
			return
		}

		cellSymbol := strings.TrimSpace(e.ChildText("td:nth-child(2)"))
		if cellSymbol != symbol {
			return
		}

		quote := &Quote{Symbol: cellSymbol}

		if ltp, err := parseFloat(e.ChildText("td:nth-child(3)")); err == nil {
			quote.LTP = ltp
		}
		if open, err := parseFloat(e.ChildText("td:nth-child(6)")); err == nil {
			quote.Open = open
		}
		if high, err := parseFloat(e.ChildText("td:nth-child(7)")); err == nil {
			quote.High = high
		}
		if low, err := parseFloat(e.ChildText("td:nth-child(8)")); err == nil {
			quote.Low = low
		}
		if volume, err := parseFloat(e.ChildText("td:nth-child(9)")); err == nil {
			quote.Volume = volume
		}
		if prevClose, err := parseFloat(e.ChildText("td:nth-child(10)")); err == nil {
			quote.PrevClose = prevClose
		}

		foundQuote = quote
	})

	err := c.Visit("https://www.sharesansar.com/live-trading")
	if err != nil {
		return nil, fmt.Errorf("failed to visit page: %w", err)
	}

	if foundQuote == nil {
		return nil, fmt.Errorf("symbol %s not found", symbol)
	}

	return foundQuote, nil
}

func parseFloat(s string) (float64, error) {
	cleaned := strings.ReplaceAll(strings.TrimSpace(s), ",", "")
	if cleaned == "" {
		return 0, fmt.Errorf("empty string")
	}
	return strconv.ParseFloat(cleaned, 64)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
