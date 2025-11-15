package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FormField represents a single input field in the form
type FormField struct {
	Key      string
	Label    string
	Help     string
	Input    textinput.Model
	Required bool
}

// FormModel is the bubbletea model for the interactive form
type FormModel struct {
	fields       []FormField
	currentField int
	submitted    bool
	values       map[string]string
	err          error
}

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))  // White/bright
	cursorStyle  = focusedStyle.Copy()
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))  // Light grey but still readable
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
)

// NewFormModel creates a new form with all the input fields
func NewFormModel(defaults map[string]string) FormModel {
	fields := []FormField{
		makeField("inflation_rate", "Inflation Rate (%)", "Annual inflation for all recurring costs", defaults),
		makeField("purchase_price", "Purchase Price ($)", "Property purchase price", defaults),
		makeField("downpayment", "Downpayment ($)", "Initial payment amount", defaults),
		makeField("loan_rate", "Loan Rate (%)", "Annual interest rate (e.g., 6.5)", defaults),
		makeField("loan_duration", "Loan Duration", "Loan term (e.g., 5y, 30y)", defaults),
		makeField("annual_insurance", "Annual Insurance ($)", "Yearly insurance cost", defaults),
		makeField("annual_taxes", "Other Annual Costs ($)", "Taxes, HOA fees, etc.", defaults),
		makeField("monthly_expenses", "Monthly Expenses ($)", "Monthly HOA, utilities, etc.", defaults),
		makeField("appreciation_rate", "Appreciation Rate (%)", "Annual property value change (e.g., 3 or -2)", defaults),
		makeField("rent_deposit", "Rental Deposit ($)", "Initial rental deposit", defaults),
		makeField("monthly_rent", "Monthly Rent ($)", "Base monthly rent amount", defaults),
		makeField("annual_rent_costs", "Annual Rent Costs ($)", "Yearly rental-related costs", defaults),
		makeField("other_annual_costs", "Other Annual Costs ($)", "Additional yearly costs for renting", defaults),
		makeField("investment_return_rate", "Investment Return Rate (%)", "Expected return on investments (e.g., 7)", defaults),
	}

	// Focus the first field
	fields[0].Input.Focus()

	return FormModel{
		fields:       fields,
		currentField: 0,
		submitted:    false,
		values:       make(map[string]string),
	}
}

func makeField(key, label, help string, defaults map[string]string) FormField {
	ti := textinput.New()
	ti.Placeholder = "0"
	ti.CharLimit = 32
	ti.Width = 30  // Fixed width to prevent jumping

	if val, ok := defaults[key]; ok {
		ti.SetValue(val)
	}

	return FormField{
		Key:      key,
		Label:    label,
		Help:     help,
		Input:    ti,
		Required: true,
	}
}

func (m FormModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "ctrl+k":
			// Save values and submit
			for _, field := range m.fields {
				m.values[field.Key] = field.Input.Value()
			}
			m.submitted = true
			return m, tea.Quit

		case "up":
			if m.currentField > 0 {
				m.fields[m.currentField].Input.Blur()
				m.currentField--
				m.fields[m.currentField].Input.Focus()
			}

		case "down":
			if m.currentField < len(m.fields)-1 {
				m.fields[m.currentField].Input.Blur()
				m.currentField++
				m.fields[m.currentField].Input.Focus()
			}
		}
	}

	// Update the focused input field
	var cmd tea.Cmd
	m.fields[m.currentField].Input, cmd = m.fields[m.currentField].Input.Update(msg)
	return m, cmd
}

func (m FormModel) View() string {
	if m.submitted {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("┌─────────────────────────────────────────────────────────────────────────────────────────────────┐"))
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("│                              Rent vs Buy Calculator                                             │"))
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("└─────────────────────────────────────────────────────────────────────────────────────────────────┘"))
	b.WriteString("\n\n")

	// Render fields in 2 columns
	for i := 0; i < len(m.fields); i += 2 {
		// Left column field
		leftField := m.fields[i]
		var leftLabel string
		if i == m.currentField {
			leftLabel = focusedStyle.Render("❯ " + leftField.Label)
		} else {
			leftLabel = blurredStyle.Render("  " + leftField.Label)
		}

		// Right column field (if exists)
		var rightLabel, rightInput string
		if i+1 < len(m.fields) {
			rightField := m.fields[i+1]
			if i+1 == m.currentField {
				rightLabel = focusedStyle.Render("❯ " + rightField.Label)
			} else {
				rightLabel = blurredStyle.Render("  " + rightField.Label)
			}
			rightInput = fmt.Sprintf("%-35s", rightField.Input.View())  // Fixed width padding
		}

		// Print labels side by side (padded to 50 chars each)
		b.WriteString(fmt.Sprintf("%-50s  %s\n", leftLabel, rightLabel))

		// Print inputs side by side (with fixed spacing)
		leftInput := fmt.Sprintf("%-35s", leftField.Input.View())  // Fixed width padding
		b.WriteString(fmt.Sprintf("  %s  %s\n", leftInput, rightInput))

		b.WriteString("\n")
	}

	// Show help text for current field at the bottom
	currentField := m.fields[m.currentField]
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  " + currentField.Help))
	b.WriteString("\n\n")

	// Navigation help
	b.WriteString(helpStyle.Render("  ↑/↓: Navigate  Ctrl+K: Calculate  Ctrl+C/Esc: Quit"))
	b.WriteString("\n")

	return b.String()
}

// RunInteractiveForm runs the interactive form and returns the values
func RunInteractiveForm(defaults map[string]string) (map[string]string, error) {
	m := NewFormModel(defaults)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	model := finalModel.(FormModel)
	if !model.submitted {
		return nil, fmt.Errorf("form cancelled")
	}

	return model.values, nil
}
