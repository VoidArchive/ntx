package watchlist

import (
	"fmt"
	"ntx/internal/ui/common"
	"slices"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// WatchlistPane displays tracked symbols with real-time price updates and alerts
type WatchlistPane struct {
	common.BasePane

	// Watchlist data
	Symbols []string
	Prices  map[string]WatchedSymbol
	Alerts  map[string]PriceAlert

	// Display preferences
	selectedIndex int
	showAlerts    bool
	sortBy        WatchlistSort
	autoRefresh   bool

	// Input mode
	inputMode   bool
	inputBuffer string

	// Statistics
	LastUpdate  time.Time
	UpdateCount int64
}

// WatchedSymbol represents a symbol being watched
type WatchedSymbol struct {
	Symbol        string
	LastPrice     float64
	Change        float64
	ChangePercent float64
	Volume        int64
	High          float64
	Low           float64
	Open          float64
	LastUpdate    time.Time
	AlertStatus   AlertStatus
}

// PriceAlert represents a price alert for a symbol
type PriceAlert struct {
	Symbol      string
	Type        AlertType
	Price       float64
	Triggered   bool
	CreatedAt   time.Time
	TriggeredAt time.Time
	Message     string
}

// AlertType defines the type of price alert
type AlertType string

const (
	AlertAbove AlertType = "above"
	AlertBelow AlertType = "below"
)

// AlertStatus indicates if a symbol has active alerts
type AlertStatus string

const (
	AlertStatusNone      AlertStatus = "none"
	AlertStatusActive    AlertStatus = "active"
	AlertStatusTriggered AlertStatus = "triggered"
)

// WatchlistSort defines sorting options for the watchlist
type WatchlistSort string

const (
	SortBySymbol      WatchlistSort = "symbol"
	SortByPrice       WatchlistSort = "price"
	SortByChange      WatchlistSort = "change"
	SortByPercent     WatchlistSort = "percent"
	SortByVolume      WatchlistSort = "volume"
	SortByAlertStatus WatchlistSort = "alert_status"
)

// NewWatchlistPane creates a new watchlist pane
func NewWatchlistPane() *WatchlistPane {
	return &WatchlistPane{
		BasePane: common.BasePane{
			Type:  common.PaneTypeWatchlist,
			Title: "Watchlist",
		},
		Symbols:     make([]string, 0),
		Prices:      make(map[string]WatchedSymbol),
		Alerts:      make(map[string]PriceAlert),
		showAlerts:  true,
		sortBy:      SortBySymbol,
		autoRefresh: true,
	}
}

// Init initializes the watchlist pane
func (wp *WatchlistPane) Init() tea.Cmd {
	// Initialize with some default NEPSE symbols
	wp.Symbols = []string{"NABIL", "EBL", "KTM", "HIDCL", "ADBL", "GBIME"}
	return nil
}

// Update handles messages and updates the pane
func (wp *WatchlistPane) Update(msg tea.Msg) (common.Pane, tea.Cmd) {
	switch msg := msg.(type) {
	case common.MarketDataMsg:
		wp.updateSymbolData(msg)
		wp.MarkUpdated()

	case common.WatchlistUpdateMsg:
		wp.updateWatchlistData(msg)
		wp.MarkUpdated()

	case tea.KeyMsg:
		if wp.Active {
			return wp, wp.HandleKeypress(msg.String())
		}
	}

	return wp, nil
}

// View renders the watchlist pane
func (wp *WatchlistPane) View() string {
	style := wp.GetStyle()
	content := wp.renderContent()

	return style.Render(content)
}

// HandleKeypress handles pane-specific key presses
func (wp *WatchlistPane) HandleKeypress(key string) tea.Cmd {
	if wp.inputMode {
		return wp.handleInputMode(key)
	}

	switch key {
	case "j", "down":
		if wp.selectedIndex < len(wp.Symbols)-1 {
			wp.selectedIndex++
		}

	case "k", "up":
		if wp.selectedIndex > 0 {
			wp.selectedIndex--
		}

	case "a":
		wp.showAlerts = !wp.showAlerts

	case "s":
		wp.cycleSortOption()
		wp.sortWatchlist()

	case "enter":
		wp.inputMode = true
		wp.inputBuffer = ""

	case "x":
		if len(wp.Symbols) > 0 && wp.selectedIndex < len(wp.Symbols) {
			wp.removeSelectedSymbol()
		}

	case "t":
		if len(wp.Symbols) > 0 && wp.selectedIndex < len(wp.Symbols) {
			return wp.createPriceAlert()
		}

	case "c":
		wp.clearTriggeredAlerts()
	}

	return nil
}

// Refresh triggers a refresh of watchlist data
func (wp *WatchlistPane) Refresh() tea.Cmd {
	return common.NewRefreshMsg(false)
}

// renderContent renders the pane content
func (wp *WatchlistPane) renderContent() string {
	var sections []string

	// Title with statistics
	sections = append(sections, wp.renderTitleWithStats())

	// Input mode overlay
	if wp.inputMode {
		sections = append(sections, wp.renderInputMode())
		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	}

	// Watchlist table
	sections = append(sections, wp.renderWatchlist())

	// Alerts section
	if wp.showAlerts {
		sections = append(sections, wp.renderAlerts())
	}

	// Help section
	sections = append(sections, wp.renderHelp())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderTitleWithStats renders the title with watchlist statistics
func (wp *WatchlistPane) renderTitleWithStats() string {
	title := wp.RenderTitle()

	if len(wp.Symbols) > 0 {
		stats := fmt.Sprintf(" (%d symbols)", len(wp.Symbols))
		title += common.MutedStyle.Render(stats)
	}

	return title
}

// renderWatchlist renders the main watchlist table
func (wp *WatchlistPane) renderWatchlist() string {
	if len(wp.Symbols) == 0 {
		return common.MutedStyle.Render("No symbols in watchlist. Press Enter to add symbols.")
	}

	var lines []string

	// Table header
	headerLine := wp.renderTableHeader()
	lines = append(lines, headerLine)

	// Symbol rows
	for i, symbol := range wp.Symbols {
		isSelected := i == wp.selectedIndex && wp.Active
		symbolLine := wp.renderSymbolRow(symbol, isSelected)
		lines = append(lines, symbolLine)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderTableHeader renders the watchlist table header
func (wp *WatchlistPane) renderTableHeader() string {
	headers := []string{
		padCenter("Symbol", 8),
		padCenter("Price", 10),
		padCenter("Change", 10),
		padCenter("Change%", 10),
		padCenter("Volume", 10),
		padCenter("H/L", 12),
		padCenter("Alert", 6),
	}

	headerRow := strings.Join(headers, " ")
	return common.TableHeaderStyle.Render(headerRow)
}

// renderSymbolRow renders a single symbol row
func (wp *WatchlistPane) renderSymbolRow(symbol string, isSelected bool) string {
	style := common.TableCellStyle
	if isSelected {
		style = common.TableSelectedStyle
	}

	// Get symbol data
	data, exists := wp.Prices[symbol]
	if !exists {
		// No data available
		cells := []string{
			padCenter(symbol, 8),
			padCenter("--", 10),
			padCenter("--", 10),
			padCenter("--", 10),
			padCenter("--", 10),
			padCenter("--", 12),
			padCenter("--", 6),
		}
		row := strings.Join(cells, " ")
		return style.Render(row)
	}

	// Format cells with data
	cells := []string{
		padCenter(symbol, 8),
		padCenter(fmt.Sprintf("%.2f", data.LastPrice), 10),
		padCenter(formatChange(data.Change), 10),
		padCenter(formatPercentage(data.ChangePercent), 10),
		padCenter(formatVolume(data.Volume), 10),
		padCenter(fmt.Sprintf("%.2f/%.2f", data.High, data.Low), 12),
		padCenter(formatAlertStatus(data.AlertStatus), 6),
	}

	row := strings.Join(cells, " ")
	return style.Render(row)
}

// renderAlerts renders active and triggered alerts
func (wp *WatchlistPane) renderAlerts() string {
	var lines []string

	// Section header
	header := common.SubHeaderStyle.Render(" Price Alerts ")
	lines = append(lines, header)

	hasAlerts := false
	for _, alert := range wp.Alerts {
		if alert.Triggered {
			alertLine := wp.formatTriggeredAlert(alert)
			lines = append(lines, alertLine)
			hasAlerts = true
		}
	}

	for _, alert := range wp.Alerts {
		if !alert.Triggered {
			alertLine := wp.formatActiveAlert(alert)
			lines = append(lines, alertLine)
			hasAlerts = true
		}
	}

	if !hasAlerts {
		lines = append(lines, common.MutedStyle.Render("No active alerts"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderInputMode renders the symbol input interface
func (wp *WatchlistPane) renderInputMode() string {
	var lines []string

	// Input prompt
	prompt := common.InfoStyle.Render("Add Symbol: ")
	input := common.ContentStyle.Render(wp.inputBuffer + "█")
	lines = append(lines, prompt+input)

	// Help
	help := common.HelpStyle.Render("Enter: Add symbol • Esc: Cancel")
	lines = append(lines, help)

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderHelp renders pane-specific help
func (wp *WatchlistPane) renderHelp() string {
	if !wp.Active {
		return ""
	}

	help := []string{
		common.KeyStyle.Render("↑↓") + ": Select",
		common.KeyStyle.Render("Enter") + ": Add symbol",
		common.KeyStyle.Render("x") + ": Remove",
		common.KeyStyle.Render("t") + ": Set alert",
		common.KeyStyle.Render("s") + ": Sort",
		common.KeyStyle.Render("a") + ": Toggle alerts",
	}

	return common.HelpStyle.Render(strings.Join(help, " • "))
}

// Helper methods

// handleInputMode handles key presses in input mode
func (wp *WatchlistPane) handleInputMode(key string) tea.Cmd {
	switch key {
	case "esc":
		wp.inputMode = false
		wp.inputBuffer = ""

	case "enter":
		if wp.inputBuffer != "" {
			wp.addSymbol(strings.ToUpper(wp.inputBuffer))
			wp.inputMode = false
			wp.inputBuffer = ""
		}

	case "backspace":
		if len(wp.inputBuffer) > 0 {
			wp.inputBuffer = wp.inputBuffer[:len(wp.inputBuffer)-1]
		}

	default:
		if len(key) == 1 && len(wp.inputBuffer) < 10 {
			wp.inputBuffer += key
		}
	}

	return nil
}

// updateSymbolData updates data for a specific symbol
func (wp *WatchlistPane) updateSymbolData(msg common.MarketDataMsg) {
	// Check if symbol is in watchlist
	symbolExists := slices.Contains(wp.Symbols, msg.Symbol)

	if !symbolExists {
		return
	}

	// Calculate change percentage safely to avoid division by zero
	changePercent := 0.0
	if prevPrice := msg.LastPrice - msg.Change; prevPrice != 0 {
		changePercent = (msg.Change / prevPrice) * 100
	}

	// Update symbol data
	data := WatchedSymbol{
		Symbol:        msg.Symbol,
		LastPrice:     msg.LastPrice,
		Change:        msg.Change,
		ChangePercent: changePercent,
		Volume:        msg.Volume,
		LastUpdate:    msg.Timestamp,
		AlertStatus:   AlertStatusNone,
	}

	// Preserve high/low if updating existing data
	if existing, exists := wp.Prices[msg.Symbol]; exists {
		data.High = existing.High
		data.Low = existing.Low
		data.Open = existing.Open

		// Update high/low
		if msg.LastPrice > data.High {
			data.High = msg.LastPrice
		}
		if msg.LastPrice < data.Low || data.Low == 0 {
			data.Low = msg.LastPrice
		}
	} else {
		// First update for this symbol
		data.High = msg.LastPrice
		data.Low = msg.LastPrice
		data.Open = msg.LastPrice
	}

	// Check for alert triggers
	wp.checkAlerts(msg.Symbol, msg.LastPrice)

	// Update alert status
	if alert, exists := wp.Alerts[msg.Symbol]; exists {
		if alert.Triggered {
			data.AlertStatus = AlertStatusTriggered
		} else {
			data.AlertStatus = AlertStatusActive
		}
	}

	wp.Prices[msg.Symbol] = data
	wp.UpdateCount++
	wp.LastUpdate = time.Now()
}

// updateWatchlistData updates multiple symbols from watchlist message
func (wp *WatchlistPane) updateWatchlistData(msg common.WatchlistUpdateMsg) {
	for _, symbol := range msg.Symbols {
		if price, exists := msg.Prices[symbol]; exists {
			change := 0.0
			if changeVal, hasChange := msg.Changes[symbol]; hasChange {
				change = changeVal
			}

			marketMsg := common.MarketDataMsg{
				Symbol:    symbol,
				LastPrice: price,
				Change:    change,
				Timestamp: msg.LastUpdate,
			}
			wp.updateSymbolData(marketMsg)
		}
	}
}

// addSymbol adds a new symbol to the watchlist with validation
func (wp *WatchlistPane) addSymbol(symbol string) {
	// Validate symbol format (NEPSE symbols are typically 2-10 alphanumeric characters)
	if len(symbol) < 2 || len(symbol) > 10 {
		return // Invalid symbol length
	}

	// Basic alphanumeric validation for NEPSE symbols
	for _, r := range symbol {
		if (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			return // Invalid character in symbol
		}
	}

	// Check if symbol already exists
	if slices.Contains(wp.Symbols, symbol) {
		return // Already exists
	}

	wp.Symbols = append(wp.Symbols, symbol)
	wp.sortWatchlist()
}

// removeSelectedSymbol removes the currently selected symbol
func (wp *WatchlistPane) removeSelectedSymbol() {
	if len(wp.Symbols) == 0 || wp.selectedIndex >= len(wp.Symbols) {
		return
	}

	symbol := wp.Symbols[wp.selectedIndex]

	// Remove from symbols list
	wp.Symbols = slices.Delete(wp.Symbols, wp.selectedIndex, wp.selectedIndex+1)

	// Remove associated data and alerts
	delete(wp.Prices, symbol)
	delete(wp.Alerts, symbol)

	// Adjust selected index
	if wp.selectedIndex >= len(wp.Symbols) && len(wp.Symbols) > 0 {
		wp.selectedIndex = len(wp.Symbols) - 1
	}
}

// createPriceAlert creates a price alert for the selected symbol
func (wp *WatchlistPane) createPriceAlert() tea.Cmd {
	// This would typically open a dialog or input mode for alert parameters
	// For now, we'll create a simple alert above current price
	if len(wp.Symbols) == 0 || wp.selectedIndex >= len(wp.Symbols) {
		return nil
	}

	symbol := wp.Symbols[wp.selectedIndex]
	if data, exists := wp.Prices[symbol]; exists {
		// Create alert 5% above current price
		alert := PriceAlert{
			Symbol:    symbol,
			Type:      AlertAbove,
			Price:     data.LastPrice * 1.05,
			CreatedAt: time.Now(),
			Message:   fmt.Sprintf("%s above %.2f", symbol, data.LastPrice*1.05),
		}
		wp.Alerts[symbol] = alert
	}

	return nil
}

// checkAlerts checks if any alerts should be triggered
func (wp *WatchlistPane) checkAlerts(symbol string, price float64) {
	alert, exists := wp.Alerts[symbol]
	if !exists || alert.Triggered {
		return
	}

	triggered := false
	switch alert.Type {
	case AlertAbove:
		triggered = price >= alert.Price
	case AlertBelow:
		triggered = price <= alert.Price
	}

	if triggered {
		alert.Triggered = true
		alert.TriggeredAt = time.Now()
		wp.Alerts[symbol] = alert
	}
}

// clearTriggeredAlerts removes all triggered alerts
func (wp *WatchlistPane) clearTriggeredAlerts() {
	for symbol, alert := range wp.Alerts {
		if alert.Triggered {
			delete(wp.Alerts, symbol)
		}
	}
}

// cycleSortOption cycles through sort options
func (wp *WatchlistPane) cycleSortOption() {
	switch wp.sortBy {
	case SortBySymbol:
		wp.sortBy = SortByPrice
	case SortByPrice:
		wp.sortBy = SortByChange
	case SortByChange:
		wp.sortBy = SortByPercent
	case SortByPercent:
		wp.sortBy = SortByVolume
	case SortByVolume:
		wp.sortBy = SortByAlertStatus
	case SortByAlertStatus:
		wp.sortBy = SortBySymbol
	}
}

// sortWatchlist sorts the watchlist based on current sort option using efficient Go sort
func (wp *WatchlistPane) sortWatchlist() {
	// Use Go's built-in efficient sorting algorithm (typically Introsort: O(n log n))
	sort.Slice(wp.Symbols, func(i, j int) bool {
		return wp.shouldSwap(wp.Symbols[i], wp.Symbols[j])
	})
}

// shouldSwap determines if two symbols should be swapped based on sort criteria
func (wp *WatchlistPane) shouldSwap(symbol1, symbol2 string) bool {
	data1, exists1 := wp.Prices[symbol1]
	data2, exists2 := wp.Prices[symbol2]

	if !exists1 && !exists2 {
		return symbol1 > symbol2
	}
	if !exists1 {
		return false
	}
	if !exists2 {
		return true
	}

	switch wp.sortBy {
	case SortBySymbol:
		return symbol1 > symbol2
	case SortByPrice:
		return data1.LastPrice < data2.LastPrice
	case SortByChange:
		return data1.Change < data2.Change
	case SortByPercent:
		return data1.ChangePercent < data2.ChangePercent
	case SortByVolume:
		return data1.Volume < data2.Volume
	case SortByAlertStatus:
		return data1.AlertStatus < data2.AlertStatus
	}

	return false
}

// Formatting helper functions

// formatChange formats price change with color
func formatChange(change float64) string {
	if change > 0 {
		return common.GainStyle.Render(fmt.Sprintf("+%.2f", change))
	} else if change < 0 {
		return common.LossStyle.Render(fmt.Sprintf("%.2f", change))
	}
	return common.NeutralStyle.Render("0.00")
}

// formatPercentage formats percentage change with color
func formatPercentage(percent float64) string {
	if percent > 0 {
		return common.GainStyle.Render(fmt.Sprintf("+%.2f%%", percent))
	} else if percent < 0 {
		return common.LossStyle.Render(fmt.Sprintf("%.2f%%", percent))
	}
	return common.NeutralStyle.Render("0.00%")
}

// formatVolume formats volume with appropriate units
func formatVolume(volume int64) string {
	if volume >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(volume)/1000000)
	} else if volume >= 1000 {
		return fmt.Sprintf("%.1fK", float64(volume)/1000)
	}
	return fmt.Sprintf("%d", volume)
}

// formatAlertStatus formats alert status indicator
func formatAlertStatus(status AlertStatus) string {
	switch status {
	case AlertStatusTriggered:
		return common.ErrorStyle.Render("🔔")
	case AlertStatusActive:
		return common.WarningStyle.Render("⏰")
	default:
		return ""
	}
}

// formatTriggeredAlert formats a triggered alert
func (wp *WatchlistPane) formatTriggeredAlert(alert PriceAlert) string {
	indicator := common.ErrorStyle.Render("🔔 TRIGGERED:")
	message := common.ContentStyle.Render(alert.Message)
	timestamp := common.MutedStyle.Render(alert.TriggeredAt.Format("15:04"))

	return fmt.Sprintf("%s %s (%s)", indicator, message, timestamp)
}

// formatActiveAlert formats an active alert
func (wp *WatchlistPane) formatActiveAlert(alert PriceAlert) string {
	indicator := common.WarningStyle.Render("⏰ ACTIVE:")
	message := common.ContentStyle.Render(alert.Message)

	return fmt.Sprintf("%s %s", indicator, message)
}

// padCenter pads a string to center it within the given width
func padCenter(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}

	padding := width - len(s)
	leftPad := padding / 2
	rightPad := padding - leftPad

	return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
}
