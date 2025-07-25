package ui

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/VoidArchive/ntx/internal/domain"
	"github.com/VoidArchive/ntx/internal/infrastructure/importer"
)

// TransactionFilter represents filter options
type TransactionFilter struct {
	Symbol string
	Type   domain.TransactionType
	All    bool
}

// Import progress message
type ImportProgressMsg struct {
	Current int
	Total   int
	Message string
}

// Import complete message
type ImportCompleteMsg struct {
	Result *importer.ImportResult
}

// TransactionsModel handles the transactions view
type TransactionsModel struct {
	portfolio     *domain.Portfolio
	importer      *importer.CSVImporter
	transactions  []domain.Transaction
	filteredTxns  []domain.Transaction
	selectedIndex int
	filter        TransactionFilter
	showImport    bool
	importPath    string
	importing     bool
	importProgress ImportProgressMsg
	windowSize    tea.WindowSizeMsg
	symbols       []string // Available symbols for filtering
}

// NewTransactionsModel creates a new transactions model
func NewTransactionsModel(portfolio *domain.Portfolio, csvImporter *importer.CSVImporter) *TransactionsModel {
	return &TransactionsModel{
		portfolio:     portfolio,
		importer:      csvImporter,
		transactions:  []domain.Transaction{},
		filteredTxns:  []domain.Transaction{},
		selectedIndex: 0,
		filter:        TransactionFilter{All: true},
		symbols:       []string{},
	}
}

// Init initializes the transactions model
func (m *TransactionsModel) Init() tea.Cmd {
	m.refreshTransactions()
	return nil
}

// Update handles messages and updates the model
func (m *TransactionsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.importing {
			// During import, only allow cancellation
			if msg.String() == "esc" || msg.String() == "ctrl+c" {
				m.importing = false
				return m, nil
			}
			return m, nil
		}

		if m.showImport {
			return m.handleImportKeys(msg)
		}
		return m.handleTableKeys(msg)

	case tea.WindowSizeMsg:
		m.windowSize = msg

	case ImportProgressMsg:
		m.importProgress = msg
		return m, nil

	case ImportCompleteMsg:
		m.importing = false
		m.showImport = false
		m.importPath = ""
		
		// Process import result
		if len(msg.Result.Errors) > 0 {
			return m, func() tea.Msg {
				return ErrorMsg{Error: fmt.Errorf("Import failed: %s", strings.Join(msg.Result.Errors, "; "))}
			}
		}

		// Add transactions to portfolio
		for _, txn := range msg.Result.Transactions {
			if err := m.portfolio.AddTransaction(txn); err != nil {
				return m, func() tea.Msg {
					return ErrorMsg{Error: err}
				}
			}
		}

		m.refreshTransactions()
		
		successMsg := fmt.Sprintf("Imported %d transactions successfully", len(msg.Result.Transactions))
		if len(msg.Result.Warnings) > 0 {
			successMsg += fmt.Sprintf(" (%d warnings)", len(msg.Result.Warnings))
		}
		
		return m, func() tea.Msg {
			return SuccessMsg{Message: successMsg}
		}

	case RefreshPortfolioMsg:
		m.refreshTransactions()
	}

	return m, nil
}

// handleTableKeys handles key presses in table view
func (m *TransactionsModel) handleTableKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}
	case "down", "j":
		if m.selectedIndex < len(m.filteredTxns)-1 {
			m.selectedIndex++
		}
	case "i":
		m.showImport = true
		return m, nil
	case "a":
		// TODO: Add manual transaction
		return m, func() tea.Msg {
			return SuccessMsg{Message: "Add transaction feature coming soon"}
		}
	case "e":
		// TODO: Edit selected transaction
		if len(m.filteredTxns) > 0 && m.selectedIndex < len(m.filteredTxns) {
			return m, func() tea.Msg {
				return SuccessMsg{Message: "Edit transaction feature coming soon"}
			}
		}
	case "d":
		// TODO: Delete selected transaction
		if len(m.filteredTxns) > 0 && m.selectedIndex < len(m.filteredTxns) {
			return m, func() tea.Msg {
				return SuccessMsg{Message: "Delete transaction feature coming soon"}
			}
		}
	case "f":
		// TODO: Toggle filter options
		return m, func() tea.Msg {
			return SuccessMsg{Message: "Filter options coming soon"}
		}
	}
	return m, nil
}

// handleImportKeys handles key presses in import dialog
func (m *TransactionsModel) handleImportKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.showImport = false
		m.importPath = ""
		return m, nil
	case "enter":
		if m.importPath != "" {
			return m, m.startImport()
		}
	case "backspace":
		if len(m.importPath) > 0 {
			m.importPath = m.importPath[:len(m.importPath)-1]
		}
	default:
		// Add character to import path
		if len(msg.String()) == 1 {
			m.importPath += msg.String()
		}
	}
	return m, nil
}

// View renders the transactions view
func (m *TransactionsModel) View() string {
	if m.windowSize.Width == 0 {
		return "Loading transactions..."
	}

	if m.importing {
		return m.renderImportProgress()
	}

	if m.showImport {
		return m.renderImportDialog()
	}

	return m.renderTransactionsView()
}

// renderTransactionsView renders the main transactions view
func (m *TransactionsModel) renderTransactionsView() string {
	var content strings.Builder

	// Section title
	title := SectionTitleStyle.Render("📋 Transaction History")
	content.WriteString(title)
	content.WriteString("\n\n")

	// Create split layout
	leftPanelWidth := 30
	rightPanelWidth := m.windowSize.Width - leftPanelWidth - 8

	// Left panel (actions and filters)
	leftPanel := m.renderLeftPanel(leftPanelWidth)
	
	// Right panel (transactions table)
	rightPanel := m.renderTransactionsTable(rightPanelWidth)

	// Join panels horizontally
	splitView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		strings.Repeat(" ", 2),
		rightPanel,
	)
	
	content.WriteString(splitView)
	content.WriteString("\n")

	// Help text
	helpText := m.renderHelpText()
	content.WriteString(helpText)

	return content.String()
}

// renderLeftPanel renders the left action panel
func (m *TransactionsModel) renderLeftPanel(width int) string {
	var panel strings.Builder

	// Actions section
	panel.WriteString(SectionTitleStyle.Render("Actions"))
	panel.WriteString("\n")
	
	actions := []string{
		KeybindStyle.Render("[i]") + " Import CSV",
		KeybindStyle.Render("[a]") + " Add Transaction",
		KeybindStyle.Render("[e]") + " Edit Selected",
		KeybindStyle.Render("[d]") + " Delete Selected",
		"",
		SectionTitleStyle.Render("Filter:"),
		fmt.Sprintf("[ ] All Symbols (%d)", len(m.transactions)),
	}

	// Add symbol filters
	for _, symbol := range m.symbols {
		count := m.countTransactionsForSymbol(symbol)
		checkbox := "[ ]"
		if m.filter.Symbol == symbol {
			checkbox = "[x]"
		}
		actions = append(actions, fmt.Sprintf("%s %s (%d)", checkbox, symbol, count))
	}

	actions = append(actions, "")
	actions = append(actions, fmt.Sprintf("Total: %d transactions", len(m.filteredTxns)))

	actionText := strings.Join(actions, "\n")
	
	return PanelStyle.
		Width(width).
		Height(m.windowSize.Height - 10).
		Render(actionText)
}

// renderTransactionsTable renders the transactions table
func (m *TransactionsModel) renderTransactionsTable(width int) string {
	if len(m.filteredTxns) == 0 {
		noData := PanelStyle.
			Width(width).
			Height(m.windowSize.Height - 10).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No transactions found.\nPress 'i' to import from CSV.")
		return noData
	}

	headers := []string{"Date", "Symbol", "Type", "Qty", "Price", "Total", "Description"}
	colWidths := []int{12, 8, 6, 8, 12, 12, width - 58} // Remaining width for description

	// Render headers
	var headerCells []string
	for i, header := range headers {
		cell := TableHeaderStyle.Width(colWidths[i]).Render(header)
		headerCells = append(headerCells, cell)
	}
	headerRow := lipgloss.JoinHorizontal(lipgloss.Left, headerCells...)

	// Render data rows (show only visible rows for performance)
	var rows []string
	rows = append(rows, headerRow)

	visibleRows := 20 // Maximum visible rows
	startIdx := m.selectedIndex - visibleRows/2
	if startIdx < 0 {
		startIdx = 0
	}
	endIdx := startIdx + visibleRows
	if endIdx > len(m.filteredTxns) {
		endIdx = len(m.filteredTxns)
		startIdx = endIdx - visibleRows
		if startIdx < 0 {
			startIdx = 0
		}
	}

	for i := startIdx; i < endIdx; i++ {
		txn := m.filteredTxns[i]
		
		// Style based on selection
		rowStyle := StyleForTableRow(i, i == m.selectedIndex)

		// Format date
		dateStr := txn.Date.Format("2006-01-02")
		
		// Format type with color
		typeStyle := rowStyle
		switch txn.Type {
		case domain.TransactionBuy:
			typeStyle = typeStyle.Foreground(ColorSuccess)
		case domain.TransactionSell:
			typeStyle = typeStyle.Foreground(ColorDanger)
		case domain.TransactionBonus, domain.TransactionRights:
			typeStyle = typeStyle.Foreground(ColorWarning)
		}

		// Format price and total
		priceText := txn.Price.String()
		totalText := txn.Cost.String()
		
		// Check for default prices (assuming Rs.100.00 is default)
		if txn.Price.Paisa() == 10000 { // Rs.100.00 in paisa
			priceText += "*"
			totalText += "*"
		}

		// Truncate description
		description := txn.Description
		if len(description) > colWidths[6]-2 {
			description = description[:colWidths[6]-5] + "..."
		}

		cells := []string{
			rowStyle.Width(colWidths[0]).Render(dateStr),
			rowStyle.Width(colWidths[1]).Render(txn.StockSymbol),
			typeStyle.Width(colWidths[2]).Render(string(txn.Type)),
			rowStyle.Width(colWidths[3]).Align(lipgloss.Right).Render(strconv.Itoa(txn.Quantity)),
			rowStyle.Width(colWidths[4]).Align(lipgloss.Right).Render(priceText),
			rowStyle.Width(colWidths[5]).Align(lipgloss.Right).Render(totalText),
			rowStyle.Width(colWidths[6]).Render(description),
		}

		row := lipgloss.JoinHorizontal(lipgloss.Left, cells...)
		rows = append(rows, row)
	}

	// Add scroll indicator if needed
	if len(m.filteredTxns) > visibleRows {
		scrollInfo := fmt.Sprintf("Showing %d-%d of %d", startIdx+1, endIdx, len(m.filteredTxns))
		rows = append(rows, MutedStyle.Render(scrollInfo))
	}

	table := strings.Join(rows, "\n")
	return PanelStyle.
		Width(width).
		Height(m.windowSize.Height - 10).
		Render(table)
}

// renderImportDialog renders the CSV import dialog
func (m *TransactionsModel) renderImportDialog() string {
	var content strings.Builder

	content.WriteString(DialogTitleStyle.Render("Import CSV File"))
	content.WriteString("\n\n")

	content.WriteString("Enter the path to your MeroShare CSV file:")
	content.WriteString("\n\n")

	// Input field
	inputValue := m.importPath
	if inputValue == "" {
		inputValue = "Enter file path..."
	}
	
	input := FocusedInputStyle.Width(60).Render(inputValue)
	content.WriteString(input)
	content.WriteString("\n\n")

	content.WriteString("Examples:")
	content.WriteString("\n")
	content.WriteString("  ./Transaction History.csv")
	content.WriteString("\n")
	content.WriteString("  /home/user/Downloads/Portfolio.csv")
	content.WriteString("\n\n")

	content.WriteString(HelpStyle.Render("Press Enter to import, ESC to cancel"))

	width := m.windowSize.Width - 8
	height := 15

	return DialogStyle.
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content.String())
}

// renderImportProgress renders the import progress dialog
func (m *TransactionsModel) renderImportProgress() string {
	var content strings.Builder

	content.WriteString(DialogTitleStyle.Render("Importing CSV File"))
	content.WriteString("\n\n")

	content.WriteString(m.importProgress.Message)
	content.WriteString("\n\n")

	// Progress bar
	if m.importProgress.Total > 0 {
		progress := float64(m.importProgress.Current) / float64(m.importProgress.Total)
		progressWidth := 40
		filled := int(progress * float64(progressWidth))
		
		progressBar := strings.Repeat("█", filled) + strings.Repeat("░", progressWidth-filled)
		content.WriteString(ProgressBarStyle.Render(progressBar))
		content.WriteString("\n")
		content.WriteString(fmt.Sprintf("%d / %d (%.1f%%)", 
			m.importProgress.Current, m.importProgress.Total, progress*100))
	} else {
		content.WriteString("Processing...")
	}

	content.WriteString("\n\n")
	content.WriteString(HelpStyle.Render("Press ESC or Ctrl+C to cancel"))

	width := m.windowSize.Width - 8
	height := 12

	return DialogStyle.
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content.String())
}

// renderHelpText renders help text for the transactions view
func (m *TransactionsModel) renderHelpText() string {
	helpItems := []string{
		KeybindStyle.Render("↑/↓") + " navigate",
		KeybindStyle.Render("i") + " import CSV",
		KeybindStyle.Render("a") + " add transaction",
		KeybindStyle.Render("e") + " edit",
		KeybindStyle.Render("d") + " delete",
		KeybindStyle.Render("f") + " filter",
	}

	return HelpStyle.Render(strings.Join(helpItems, " | "))
}

// startImport starts the CSV import process
func (m *TransactionsModel) startImport() tea.Cmd {
	m.importing = true
	
	return func() tea.Msg {
		// Validate file exists
		if _, err := os.Stat(m.importPath); os.IsNotExist(err) {
			return ErrorMsg{Error: fmt.Errorf("file not found: %s", m.importPath)}
		}

		// Open file
		file, err := os.Open(m.importPath)
		if err != nil {
			return ErrorMsg{Error: fmt.Errorf("failed to open file: %w", err)}
		}
		defer file.Close()

		// Create progress callback matching the expected signature
		progressCallback := func(processed int, stats importer.ImportStats) {
			// Note: This won't work in bubbletea as expected
			// We'd need to use tea.Cmd properly for real progress updates
		}

		// Import with progress
		result, err := m.importer.ImportFromReaderWithCallback(file, progressCallback)
		if err != nil {
			return ImportCompleteMsg{Result: &importer.ImportResult{
				Transactions: []domain.Transaction{},
				Warnings:     []string{},
				Errors:       []string{err.Error()},
			}}
		}

		return ImportCompleteMsg{Result: result}
	}
}

// refreshTransactions updates the transactions list from the portfolio
func (m *TransactionsModel) refreshTransactions() {
	// TODO: Portfolio doesn't store transaction history by design
	// Need to implement transaction storage or use alternative approach
	// For now, using empty slice as placeholder
	m.transactions = []domain.Transaction{}
	
	// Sort by date (newest first)
	sort.Slice(m.transactions, func(i, j int) bool {
		return m.transactions[i].Date.After(m.transactions[j].Date)
	})

	// Update available symbols
	symbolMap := make(map[string]bool)
	for _, txn := range m.transactions {
		symbolMap[txn.StockSymbol] = true
	}
	
	m.symbols = make([]string, 0, len(symbolMap))
	for symbol := range symbolMap {
		m.symbols = append(m.symbols, symbol)
	}
	sort.Strings(m.symbols)

	// Apply current filter
	m.applyFilter()

	// Adjust selected index if needed
	if m.selectedIndex >= len(m.filteredTxns) {
		m.selectedIndex = len(m.filteredTxns) - 1
	}
	if m.selectedIndex < 0 {
		m.selectedIndex = 0
	}
}

// applyFilter applies the current filter to transactions
func (m *TransactionsModel) applyFilter() {
	if m.filter.All || m.filter.Symbol == "" {
		m.filteredTxns = m.transactions
		return
	}

	m.filteredTxns = make([]domain.Transaction, 0)
	for _, txn := range m.transactions {
		if m.filter.Symbol == "" || txn.StockSymbol == m.filter.Symbol {
			m.filteredTxns = append(m.filteredTxns, txn)
		}
	}
}

// countTransactionsForSymbol counts transactions for a specific symbol
func (m *TransactionsModel) countTransactionsForSymbol(symbol string) int {
	count := 0
	for _, txn := range m.transactions {
		if txn.StockSymbol == symbol {
			count++
		}
	}
	return count
}