package forms

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"ntx/internal/portfolio/models"
	"ntx/internal/portfolio/services"
	"ntx/internal/ui/themes"
	"ntx/internal/validation"
)

// TransactionForm represents a transaction entry form
type TransactionForm struct {
	Fields          []*FormField
	CurrentField    int
	Theme           themes.Theme
	PortfolioID     int64
	OnSubmit        func(services.ExecuteTransactionRequest) error
	OnCancel        func()
	validator       *validation.TransactionValidator
	validationLevel validation.ValidationLevel
}

// NewTransactionForm creates a new transaction form
func NewTransactionForm(portfolioID int64, theme themes.Theme) *TransactionForm {
	form := &TransactionForm{
		PortfolioID:     portfolioID,
		Theme:           theme,
		CurrentField:    0,
		validator:       validation.NewTransactionValidator(),
		validationLevel: validation.ValidationLenient, // Default to lenient for better UX
	}

	// Create form fields
	form.Fields = []*FormField{
		NewFormField("Symbol", "e.g., NABIL", FieldTypeText, theme).SetRequired(),
		NewFormField("Type", "buy/sell", FieldTypeSelect, theme).SetRequired().SetOptions([]string{"buy", "sell"}),
		NewFormField("Quantity", "Number of shares", FieldTypeNumber, theme).SetRequired(),
		NewFormField("Price (Rs.)", "Price per share", FieldTypeNumber, theme).SetRequired(),
		NewFormField("Commission (Rs.)", "Broker commission", FieldTypeNumber, theme),
		NewFormField("Tax (Rs.)", "Tax amount", FieldTypeNumber, theme),
		NewFormField("Date", "YYYY-MM-DD", FieldTypeDate, theme).SetRequired(),
		NewFormField("Notes", "Optional notes", FieldTypeText, theme),
	}

	// Set default values
	form.Fields[1].SetValue("buy") // Default to buy
	form.Fields[6].SetValue(time.Now().Format("2006-01-02")) // Default to today

	// Focus first field
	if len(form.Fields) > 0 {
		form.Fields[0].Focus()
	}

	return form
}

// Init implements tea.Model
func (f *TransactionForm) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (f *TransactionForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			f.nextField()
			return f, nil
		case "shift+tab":
			f.prevField()
			return f, nil
		case "enter":
			if f.isValid() {
				req, err := f.buildRequest()
				if err == nil && f.OnSubmit != nil {
					f.OnSubmit(req)
				}
			}
			return f, nil
		case "esc":
			if f.OnCancel != nil {
				f.OnCancel()
			}
			return f, nil
		default:
			// Forward to current field
			if f.CurrentField < len(f.Fields) {
				f.Fields[f.CurrentField].Update(msg)
			}
		}
	}

	return f, nil
}

// View implements tea.Model
func (f *TransactionForm) View() string {
	var fieldViews []string

	for i, field := range f.Fields {
		if i == f.CurrentField {
			field.Focus()
		} else {
			field.Blur()
		}
		fieldViews = append(fieldViews, field.View())
	}

	// Add instructions
	instructions := lipgloss.NewStyle().
		Foreground(f.Theme.Muted()).
		Italic(true).
		Render("Tab: Next field | Shift+Tab: Previous | Enter: Submit | Esc: Cancel")

	// Add submit button style
	submitStyle := lipgloss.NewStyle().
		Background(f.Theme.Primary()).
		Foreground(f.Theme.Background()).
		Bold(true).
		Padding(0, 2).
		Margin(1, 0)

	submitText := "Submit Transaction"
	if !f.isValid() {
		submitStyle = submitStyle.
			Background(f.Theme.Muted()).
			Foreground(f.Theme.Foreground())
		submitText = "Fix errors to submit"
	}

	submit := submitStyle.Render(submitText)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinVertical(lipgloss.Left, fieldViews...),
		"",
		submit,
		"",
		instructions,
	)
}

// nextField moves to the next form field
func (f *TransactionForm) nextField() {
	if f.CurrentField < len(f.Fields)-1 {
		f.CurrentField++
	} else {
		f.CurrentField = 0
	}
}

// prevField moves to the previous form field
func (f *TransactionForm) prevField() {
	if f.CurrentField > 0 {
		f.CurrentField--
	} else {
		f.CurrentField = len(f.Fields) - 1
	}
}

// isValid checks if all required fields are valid using NEPSE validation
func (f *TransactionForm) isValid() bool {
	// First check basic field validation
	for _, field := range f.Fields {
		field.Validate()
		if !field.Valid {
			return false
		}
	}

	// Then check NEPSE-specific validation if form is complete enough
	if f.hasBasicRequiredFields() {
		req, err := f.buildRequest()
		if err != nil {
			return false
		}

		result := f.validator.ValidateTransaction(req, f.validationLevel)
		return result.IsValid()
	}

	return true
}

// hasBasicRequiredFields checks if the minimum required fields are filled
func (f *TransactionForm) hasBasicRequiredFields() bool {
	// Symbol, type, quantity, price are minimum required for NEPSE validation
	return f.Fields[0].Value != "" && // Symbol
		f.Fields[1].Value != "" && // Type  
		f.Fields[2].Value != "" && // Quantity
		f.Fields[3].Value != ""    // Price
}

// buildRequest creates an ExecuteTransactionRequest from form data
func (f *TransactionForm) buildRequest() (services.ExecuteTransactionRequest, error) {
	req := services.ExecuteTransactionRequest{
		PortfolioID: f.PortfolioID,
	}

	// Parse form fields
	req.Symbol = strings.ToUpper(strings.TrimSpace(f.Fields[0].Value))
	req.TransactionType = strings.ToLower(strings.TrimSpace(f.Fields[1].Value))

	// Parse quantity
	quantity, err := strconv.ParseInt(f.Fields[2].Value, 10, 64)
	if err != nil {
		return req, fmt.Errorf("invalid quantity: %w", err)
	}
	req.Quantity = quantity

	// Parse price
	price, err := strconv.ParseFloat(f.Fields[3].Value, 64)
	if err != nil {
		return req, fmt.Errorf("invalid price: %w", err)
	}
	req.Price = models.NewMoney(price)

	// Parse commission (optional)
	if f.Fields[4].Value != "" {
		commission, err := strconv.ParseFloat(f.Fields[4].Value, 64)
		if err != nil {
			return req, fmt.Errorf("invalid commission: %w", err)
		}
		req.Commission = models.NewMoney(commission)
	}

	// Parse tax (optional)
	if f.Fields[5].Value != "" {
		tax, err := strconv.ParseFloat(f.Fields[5].Value, 64)
		if err != nil {
			return req, fmt.Errorf("invalid tax: %w", err)
		}
		req.Tax = models.NewMoney(tax)
	}

	// Parse date
	date, err := time.Parse("2006-01-02", f.Fields[6].Value)
	if err != nil {
		return req, fmt.Errorf("invalid date: %w", err)
	}
	req.TransactionDate = date

	// Notes (optional)
	if f.Fields[7].Value != "" {
		notes := strings.TrimSpace(f.Fields[7].Value)
		req.Notes = &notes
	}

	return req, nil
}