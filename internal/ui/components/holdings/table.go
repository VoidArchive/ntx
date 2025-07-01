/*
NTX Portfolio Management TUI - Holdings Table Component

Single source of truth table renderer with consistent layout calculations,
perfect alignment, and btop-style integrated borders for professional
financial data visualization.

Clean design eliminates alignment issues by using one layout engine
for all table components (headers, separators, data rows, totals).
*/

package holdings

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Table provides single-source-of-truth table rendering
// All table components use the same layout calculations for perfect alignment
type Table struct {
	// Layout specifications calculated once and reused
	TotalWidth   int            // Complete table width including borders
	ContentWidth int            // Available width for content (excluding borders)
	Columns      []ColumnLayout // Column specifications with exact dimensions

	// Border and separator characters for consistent formatting
	BorderLeft   string // Left border: "│ "
	BorderRight  string // Right border: " │"
	ColSeparator string // Column separator: " │ "
	RowSeparator string // Horizontal separator: "─"
	Intersection string // Border intersection: "┼"

	// Component references
	Display *HoldingsDisplay
}

// ColumnLayout defines exact layout for a single column
// All rendering functions use these exact specifications
type ColumnLayout struct {
	Title      string // Column header text
	Width      int    // Exact width in characters
	RightAlign bool   // Text alignment preference
	Visible    bool   // Column visibility
}

// NewTable creates table with calculated layout
// Performs all dimension calculations once as single source of truth
func NewTable(hd *HoldingsDisplay) *Table {
	table := &Table{
		Display: hd,
	}

	table.calculateLayout()
	return table
}

// calculateLayout performs all table layout calculations
// Single source of truth ensures perfect alignment across all components
func (t *Table) calculateLayout() {
	// Get terminal dimensions
	terminalWidth := t.Display.getTableWidth()

	// Define border elements for precise layout calculation
	t.BorderLeft = "│"
	t.BorderRight = "│"
	t.ColSeparator = "│"
	t.RowSeparator = "─"
	t.Intersection = "┼"

	// Calculate available content width
	// Actual structure: │ col1 │ col2 │ col3 │ (spaces handled by padding)
	// Each column gets 1 space on each side, separators are just │
	borderWidth := 2 // left │ + right │
	// The table should consume the full terminal width. We therefore
	// allocate the remaining space *entirely* to the content area and
	// remove the previously hard-coded two-character side margin.
	t.TotalWidth = terminalWidth
	t.ContentWidth = terminalWidth - borderWidth

	// Define column structure based on layout config
	layoutConfig := t.Display.getLayoutConfig()
	columnDefs := []struct {
		title      string
		rightAlign bool
		visible    bool
	}{
		{"Symbol", false, true},
		{"Qty", true, true},
		{"Cost", true, true},
		{"LTP", true, true},
		{"Value", true, true},
		{"Day P/L", true, layoutConfig.ShowDayPL},
		{"Total P/L", true, true},
		{"%Chg", true, true},
		{"RSI", true, layoutConfig.ShowRSI},
	}

	// Count visible columns
	visibleCount := 0
	for _, def := range columnDefs {
		if def.visible {
			visibleCount++
		}
	}

	// Calculate separator space - each separator is just │ (1 char)
	separatorWidth := 1 // Just │ character
	totalSeparatorSpace := 0
	if visibleCount > 1 {
		totalSeparatorSpace = (visibleCount - 1) * separatorWidth
	}

	// Calculate equal column width and distribute remainder
	availableForColumns := t.ContentWidth - totalSeparatorSpace
	baseWidth := availableForColumns / visibleCount
	remainder := availableForColumns % visibleCount

	// Apply minimum width constraint
	minWidth := 8
	if baseWidth < minWidth {
		baseWidth = minWidth
		// Recalculate total width to account for minimum width constraint
		actualColumnsWidth := visibleCount * baseWidth
		actualTotalWidth := actualColumnsWidth + totalSeparatorSpace + 2 // +2 for left/right borders
		t.TotalWidth = actualTotalWidth
		t.ContentWidth = actualTotalWidth - 2
		// Clear remainder since we're using uniform column widths now
		remainder = 0
	}

	// Build column layout specifications with remainder distribution
	t.Columns = []ColumnLayout{}
	columnIndex := 0
	for _, def := range columnDefs {
		if def.visible {
			// Distribute remainder pixels to first few columns
			width := baseWidth
			if columnIndex < remainder {
				width++
			}

			t.Columns = append(t.Columns, ColumnLayout{
				Title:      def.title,
				Width:      width,
				RightAlign: def.rightAlign,
				Visible:    true,
			})
			columnIndex++
		}
	}
}

// getLayoutConfig returns layout configuration based on terminal size
func (hd *HoldingsDisplay) getLayoutConfig() struct {
	ShowRSI     bool
	ShowDayPL   bool
	CompactMode bool
} {
	width := hd.TerminalSize.Width
	return struct {
		ShowRSI     bool
		ShowDayPL   bool
		CompactMode bool
	}{
		ShowRSI:     width >= 120,
		ShowDayPL:   true,
		CompactMode: width < 80,
	}
}

// Render generates complete table using unified layout
// All components use same layout source for perfect alignment
func (t *Table) Render() string {
	if len(t.Display.Holdings) == 0 {
		return t.Display.renderEmptyState()
	}

	var output strings.Builder

	// Render all table sections using unified layout system
	output.WriteString(t.RenderTopBorder())
	output.WriteString("\n")
	output.WriteString(t.RenderHeader())
	output.WriteString("\n")
	output.WriteString(t.RenderSeparator())
	output.WriteString("\n")

	// Data rows
	for i, holding := range t.Display.Holdings {
		isCurrentRow := i == t.Display.SelectedRow
		isMultiSelected := t.Display.SelectedItems[i]
		output.WriteString(t.RenderDataRow(holding, isCurrentRow, isMultiSelected))
		output.WriteString("\n")
	}

	// Add empty rows to fill available vertical space
	paddingRows := t.calculatePaddingRows()
	for range paddingRows {
		output.WriteString(t.renderEmptyRow())
		output.WriteString("\n")
	}

	if t.Display.ShowFooter {
		output.WriteString(t.RenderSeparator())
		output.WriteString("\n")
		output.WriteString(t.RenderFooter())
		output.WriteString("\n")
	}

	output.WriteString(t.RenderBottomBorder())

	return output.String()
}

// RenderTopBorder creates top border with title using exact layout and theme colors
func (t *Table) RenderTopBorder() string {
	title := "[2]Holdings"

	if t.TotalWidth < len(title)+10 {
		border := "┌" + strings.Repeat("─", t.TotalWidth-2) + "┐"
		return lipgloss.NewStyle().Foreground(t.Display.Theme.Primary()).Render(border)
	}

	titleSection := "─" + title + "─"
	remainingWidth := t.TotalWidth - len([]rune(titleSection)) - 2
	leftPadding := strings.Repeat("─", remainingWidth)

	border := "┌" + titleSection + leftPadding + "┐"
	return lipgloss.NewStyle().Foreground(t.Display.Theme.Primary()).Render(border)
}

// RenderBottomBorder creates bottom border using exact layout and theme colors
func (t *Table) RenderBottomBorder() string {
	border := "└" + strings.Repeat("─", t.TotalWidth-2) + "┘"
	return lipgloss.NewStyle().Foreground(t.Display.Theme.Primary()).Render(border)
}

// RenderHeader renders column headers with exact alignment
func (t *Table) RenderHeader() string {
	var parts []string

	for _, col := range t.Columns {
		header := col.Title
		if len(header) > col.Width {
			header = header[:col.Width-1] + "-"
		}

		// Apply theme styling
		styledHeader := t.Display.Theme.MutedStyle().Render(header)

		// Pad to exact column width with space padding
		if col.RightAlign {
			header = " " + t.padRight(styledHeader, col.Width-2) + " "
		} else {
			header = " " + t.padLeft(styledHeader, col.Width-2) + " "
		}

		parts = append(parts, header)
	}

	// Apply theme colors to separators and join: │ col1 │ col2 │ col3 │
	coloredSeparator := lipgloss.NewStyle().Foreground(t.Display.Theme.Primary()).Render("│")
	content := strings.Join(parts, coloredSeparator)

	return coloredSeparator + content + coloredSeparator
}

// RenderSeparator creates horizontal separator with perfect intersections
// Uses exact column widths to ensure alignment with vertical separators
func (t *Table) RenderSeparator() string {
	var parts []string

	// Create horizontal line for each column using exact width
	for _, col := range t.Columns {
		parts = append(parts, strings.Repeat(t.RowSeparator, col.Width))
	}

	// Create intersection pattern that matches the new separator structure
	intersectionChar := lipgloss.NewStyle().Foreground(t.Display.Theme.Primary()).Render("┼")
	content := strings.Join(parts, intersectionChar)

	// Border intersections: ├─────┼─────┼─────┤
	leftEdge := lipgloss.NewStyle().Foreground(t.Display.Theme.Primary()).Render("├")
	rightEdge := lipgloss.NewStyle().Foreground(t.Display.Theme.Primary()).Render("┤")

	return leftEdge + content + rightEdge
}

// RenderDataRow renders single data row with exact column alignment
func (t *Table) RenderDataRow(holding Holding, isCurrentRow bool, isMultiSelected bool) string {
	var parts []string

	for _, col := range t.Columns {
		content := t.getCellContent(holding, col.Title)

		// Add selection indicator for Symbol column
		if col.Title == "Symbol" && isMultiSelected {
			content = "✓" + content
		} else if col.Title == "Symbol" && isCurrentRow {
			content = "►" + content
		}

		// Truncate if too long (accounting for space padding)
		if len(content) > col.Width-2 {
			content = content[:col.Width-3] + "…"
		}

		// Apply theme colors based on content type
		styledContent := t.applyStyling(content, col.Title, holding, isCurrentRow, isMultiSelected)

		// Pad to exact column width with space padding
		if col.RightAlign {
			content = " " + t.padRight(styledContent, col.Width-2) + " "
		} else {
			content = " " + t.padLeft(styledContent, col.Width-2) + " "
		}

		parts = append(parts, content)
	}

	// Apply theme colors to separators: │ col1 │ col2 │ col3 │
	coloredSeparator := lipgloss.NewStyle().Foreground(t.Display.Theme.Primary()).Render("│")
	content := strings.Join(parts, coloredSeparator)

	return coloredSeparator + content + coloredSeparator
}

// RenderTotalRow renders portfolio total row with exact alignment
func (t *Table) RenderTotalRow() string {
	total := t.Display.GetPortfolioTotal()
	var parts []string

	for _, col := range t.Columns {
		var content string

		switch col.Title {
		case "Symbol":
			content = "Total"
		case "Qty", "Cost", "LTP", "RSI":
			content = "—"
		case "Value":
			content = FormatCurrency(total.MarketValue)
		case "Day P/L":
			content = FormatPL(total.DayPL)
		case "Total P/L":
			content = FormatPL(total.TotalPL)
		case "%Chg":
			content = FormatPercent(total.PercentChange)
		default:
			content = "—"
		}

		// Truncate if too long (accounting for space padding)
		if len(content) > col.Width-2 {
			content = content[:col.Width-3] + "…"
		}

		// Apply bold styling for totals
		styledContent := lipgloss.NewStyle().
			Foreground(t.Display.Theme.Primary()).
			Bold(true).
			Render(content)

		// Pad to exact column width with space padding
		if col.RightAlign {
			content = " " + t.padRight(styledContent, col.Width-2) + " "
		} else {
			content = " " + t.padLeft(styledContent, col.Width-2) + " "
		}

		parts = append(parts, content)
	}

	// Apply theme colors to separators: │ col1 │ col2 │ col3 │
	coloredSeparator := lipgloss.NewStyle().Foreground(t.Display.Theme.Primary()).Render("│")
	content := strings.Join(parts, coloredSeparator)

	return coloredSeparator + content + coloredSeparator
}

// RenderFooter renders footer with shortcuts and status using exact layout
func (t *Table) RenderFooter() string {
	shortcuts := t.Display.getShortcutsText()
	status := t.Display.getStatusText()

	// Calculate footer content layout using visual width for Unicode characters
	shortcutsWidth := len([]rune(shortcuts))
	statusWidth := len([]rune(status))

	var content string
	if shortcutsWidth+statusWidth+2 <= t.ContentWidth {
		spacing := t.ContentWidth - shortcutsWidth - statusWidth
		spacer := strings.Repeat(" ", spacing)
		content = shortcuts + spacer + status
	} else {
		if shortcutsWidth <= t.ContentWidth {
			content = shortcuts + strings.Repeat(" ", t.ContentWidth-shortcutsWidth)
		} else {
			// Truncate based on visual width, not byte length
			shortcutsRunes := []rune(shortcuts)
			if len(shortcutsRunes) > t.ContentWidth-1 {
				content = string(shortcutsRunes[:t.ContentWidth-1]) + "…"
			} else {
				content = shortcuts + "…"
			}
		}
	}

	// Ensure exact content width using visual width for Unicode characters
	contentVisualWidth := len([]rune(content))
	if contentVisualWidth < t.ContentWidth {
		content = content + strings.Repeat(" ", t.ContentWidth-contentVisualWidth)
	}

	// Apply theme colors to separators for footer spanning full width
	// NOTE: We no longer wrap `content` with an additional leading/trailing
	// space because the table already includes padding inside each column.
	// Keeping those spaces caused the footer to overflow by two characters,
	// making the right border mis-aligned with the top/bottom borders.

	coloredSeparator := lipgloss.NewStyle().Foreground(t.Display.Theme.Primary()).Render("│")
	styledContent := t.Display.Theme.MutedStyle().Render(content)

	return coloredSeparator + styledContent + coloredSeparator
}

// getCellContent extracts content for specific column from holding data
func (t *Table) getCellContent(holding Holding, columnTitle string) string {
	switch columnTitle {
	case "Symbol":
		return holding.Symbol
	case "Qty":
		return strconv.FormatInt(holding.Quantity, 10)
	case "Cost":
		return FormatCurrency(holding.AvgCost)
	case "LTP":
		return FormatCurrency(holding.CurrentLTP)
	case "Value":
		return FormatCurrency(holding.MarketValue)
	case "Day P/L":
		return FormatPL(holding.DayPL)
	case "Total P/L":
		return FormatPL(holding.TotalPL)
	case "%Chg":
		return FormatPercent(holding.PercentChange)
	case "RSI":
		return fmt.Sprintf("%.0f", holding.RSI)
	default:
		return "—"
	}
}

// applyStyling applies theme-based styling to cell content
func (t *Table) applyStyling(content, columnTitle string, holding Holding, isCurrentRow bool, isMultiSelected bool) string {
	var style lipgloss.Style

	if isCurrentRow {
		// Current row highlighting (cursor position)
		style = lipgloss.NewStyle().
			Background(t.Display.Theme.Primary()).
			Foreground(t.Display.Theme.Background())
	} else if isMultiSelected {
		// Multi-selection highlighting
		style = lipgloss.NewStyle().
			Background(t.Display.Theme.Muted()).
			Foreground(t.Display.Theme.Background())
	} else {
		// Apply column-specific styling
		switch columnTitle {
		case "Day P/L":
			// Use sophisticated gradient coloring for Day P/L
			color := t.Display.GetPLColor(holding.DayPL)
			style = lipgloss.NewStyle().Foreground(color)
		case "Total P/L":
			// Use sophisticated gradient coloring for Total P/L
			color := t.Display.GetPLColor(holding.TotalPL)
			style = lipgloss.NewStyle().Foreground(color)
		case "%Chg":
			// Use sophisticated gradient coloring for percentage change
			// Convert percentage to paisa equivalent for color calculation
			percentInPaisa := int64(holding.PercentChange * 100)
			color := t.Display.GetPLColor(percentInPaisa)
			style = lipgloss.NewStyle().Foreground(color)
		default:
			style = lipgloss.NewStyle().Foreground(t.Display.Theme.Foreground())
		}
	}

	return style.Render(content)
}

// padLeft pads string to exact width with left alignment
func (t *Table) padLeft(s string, width int) string {
	// Get visual width (ignoring ANSI codes)
	visualWidth := lipgloss.Width(s)
	if visualWidth >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visualWidth)
}

// padRight pads string to exact width with right alignment
func (t *Table) padRight(s string, width int) string {
	// Get visual width (ignoring ANSI codes)
	visualWidth := lipgloss.Width(s)
	if visualWidth >= width {
		return s
	}
	return strings.Repeat(" ", width-visualWidth) + s
}

// GetActualTableWidth returns the actual table width after layout calculation
// This may differ from terminal width due to minimum column width constraints
func (hd *HoldingsDisplay) GetActualTableWidth() int {
	table := NewTable(hd)
	// Return calculated width (may be larger than terminal due to minimum constraints)
	return table.TotalWidth
}

// calculatePaddingRows determines how many empty rows to add for full height usage
func (t *Table) calculatePaddingRows() int {
	terminalHeight := t.Display.TerminalSize.Height

	// Calculate fixed elements height:
	// Overview widget: 3 lines (top border + content + bottom border)
	// Spacing between overview and table: 2 lines
	// Table fixed elements: top border + header + separator + bottom border = 4 lines
	// Optional footer: separator + footer = 2 lines (if enabled)
	overviewHeight := 3
	spacingHeight := 2
	tableFixedHeight := 4
	if t.Display.ShowFooter {
		tableFixedHeight += 2 // separator + footer
	}

	usedHeight := overviewHeight + spacingHeight + tableFixedHeight
	currentDataRows := len(t.Display.Holdings)
	availableDataRows := terminalHeight - usedHeight

	// Ensure we don't go negative
	if availableDataRows <= currentDataRows {
		return 0
	}

	return availableDataRows - currentDataRows
}

// renderEmptyRow renders a single empty row for padding
func (t *Table) renderEmptyRow() string {
	var parts []string

	for _, col := range t.Columns {
		// Create empty content padded to column width
		content := " " + strings.Repeat(" ", col.Width-2) + " "
		parts = append(parts, content)
	}

	// Apply theme colors to separators
	coloredSeparator := lipgloss.NewStyle().Foreground(t.Display.Theme.Primary()).Render("│")
	content := strings.Join(parts, coloredSeparator)

	return coloredSeparator + content + coloredSeparator
}

// renderEmptyState shows message when no holdings exist with theme colors
func (hd *HoldingsDisplay) renderEmptyState() string {
	width := hd.getTableWidth()
	message := "No holdings to display. Press 'a' to add a transaction."

	topBorder := "┌─[2]Holdings─" + strings.Repeat("─", width-11) + "┐"
	emptyLine := "│" + strings.Repeat(" ", width-2) + "│"

	padding := max((width-len(message)-2)/2, 0)
	messageSpacing := strings.Repeat(" ", padding)
	messageLine := "│" + messageSpacing + message +
		strings.Repeat(" ", width-len(message)-padding-2) + "│"

	bottomBorder := "└" + strings.Repeat("─", width-2) + "┘"

	// Apply theme colors to all borders
	borderStyle := lipgloss.NewStyle().Foreground(hd.Theme.Primary())
	styledTopBorder := borderStyle.Render(topBorder)
	styledEmptyLine := borderStyle.Render(emptyLine)
	styledMessageLine := borderStyle.Render(messageLine)
	styledBottomBorder := borderStyle.Render(bottomBorder)

	return styledTopBorder + "\n" + styledEmptyLine + "\n" + styledMessageLine + "\n" +
		styledEmptyLine + "\n" + styledBottomBorder
}

// getTableWidth calculates total table width based on terminal size
func (hd *HoldingsDisplay) getTableWidth() int {
	if hd.TerminalSize.Width < 60 {
		return 60
	}
	return hd.TerminalSize.Width
}

// getShortcutsText returns responsive shortcuts text
func (hd *HoldingsDisplay) getShortcutsText() string {
	width := hd.TerminalSize.Width

	switch {
	case width >= 120:
		return "↑↓:navigate  s:sort  a:add  d:details  r:refresh  t:theme  h:help  q:quit"
	case width >= 80:
		return "↑↓:nav  s:sort  a:add  d:details  r:refresh  t:theme  q:quit"
	default:
		return "↑↓:nav  s:sort  a:add  q:quit"
	}
}

// getStatusText returns current status information
func (hd *HoldingsDisplay) getStatusText() string {
	timestamp := hd.LastUpdate.Format("15:04 MST")
	return fmt.Sprintf("Last Update: %s  │  Theme: %s", timestamp, hd.Theme.Name())
}
