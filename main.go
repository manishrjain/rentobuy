package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

var reader = bufio.NewReader(os.Stdin)
var defaultIndex = 0
var savedDefaults []string

// Global arrays for monthly costs
var monthlyBuyingCosts []float64
var monthlyRentingCosts []float64

const inputsFile = ".rentobuy_inputs.json"

func main() {
	// Load previous inputs
	savedDefaults = loadInputs()
	defaultIndex = 0
	var inputs []string

	// Get purchase price
	purchasePrice, inputStr, err := getFloatInputWithDefault("Enter purchase price: $")
	inputs = append(inputs, inputStr)
	if err != nil {
		fmt.Println("Invalid purchase price")
		return
	}

	// Get downpayment
	downpayment, inputStr, err := getFloatInputWithDefault("Enter downpayment: $")
	inputs = append(inputs, inputStr)
	if err != nil {
		fmt.Println("Invalid downpayment")
		return
	}

	// Calculate loan amount
	loanAmount := purchasePrice - downpayment

	if loanAmount <= 0 {
		fmt.Println("No loan needed. Purchase can be made with downpayment.")
		return
	}

	// Get loan rate
	annualRate, inputStr, err := getFloatInputWithDefault("Enter loan rate percentage (e.g., 6.5 for 6.5%): ")
	inputs = append(inputs, inputStr)
	if err != nil {
		fmt.Println("Invalid loan rate")
		return
	}

	// Get loan duration
	totalMonths, inputStr, err := getStringInputAndParseWithDefault("Enter loan duration (e.g., 5y6m for 5 years 6 months): ", parseDuration)
	inputs = append(inputs, inputStr)
	if err != nil {
		fmt.Println("Invalid duration format:", err)
		return
	}

	// Get annual expenses
	fmt.Println("\n--- Annual Fixed Expenses ---")
	annualInsurance, inputStr, err := getFloatInputWithDefault("Enter annual insurance: $")
	inputs = append(inputs, inputStr)
	if err != nil {
		fmt.Println("Invalid insurance amount")
		return
	}

	annualTaxes, inputStr, err := getFloatInputWithDefault("Enter annual taxes: $")
	inputs = append(inputs, inputStr)
	if err != nil {
		fmt.Println("Invalid taxes amount")
		return
	}

	totalAnnualExpenses := annualInsurance + annualTaxes

	// Get monthly expenses
	fmt.Println("\n--- Monthly Fixed Expenses ---")
	monthlyExpenses, inputStr, err := getFloatInputWithDefault("Enter monthly fixed expenses (HOA, utilities, etc.): $")
	inputs = append(inputs, inputStr)
	if err != nil {
		fmt.Println("Invalid monthly expenses")
		return
	}

	// Get asset appreciation rate
	fmt.Println("\n--- Asset Appreciation ---")
	appreciationRate, inputStr, err := getFloatInputWithDefault("Enter annual appreciation rate (e.g., 3 for 3%, -2 for -2% depreciation): ")
	inputs = append(inputs, inputStr)
	if err != nil {
		fmt.Println("Invalid appreciation rate")
		return
	}

	// Get renting parameters
	fmt.Println("\n--- Renting Parameters ---")
	rentDeposit, inputStr, err := getFloatInputWithDefault("Enter rental deposit: $")
	inputs = append(inputs, inputStr)
	if err != nil {
		fmt.Println("Invalid deposit amount")
		return
	}

	monthlyRent, inputStr, err := getFloatInputWithDefault("Enter monthly rent: $")
	inputs = append(inputs, inputStr)
	if err != nil {
		fmt.Println("Invalid monthly rent")
		return
	}

	annualRentCosts, inputStr, err := getFloatInputWithDefault("Enter annual rent costs: $")
	inputs = append(inputs, inputStr)
	if err != nil {
		fmt.Println("Invalid annual rent costs")
		return
	}

	otherAnnualCosts, inputStr, err := getFloatInputWithDefault("Enter other annual costs: $")
	inputs = append(inputs, inputStr)
	if err != nil {
		fmt.Println("Invalid other annual costs")
		return
	}

	investmentReturnRate, inputStr, err := getFloatInputWithDefault("Enter investment return rate (e.g., 7 for 7%): ")
	inputs = append(inputs, inputStr)
	if err != nil {
		fmt.Println("Invalid investment return rate")
		return
	}

	// Save inputs for next time
	saveInputs(inputs)

	// Calculate monthly payment for buying
	monthlyRate := annualRate / 100 / 12
	monthlyLoanPayment := calculateMonthlyPayment(loanAmount, monthlyRate, totalMonths)
	monthlyRecurringExpenses := (totalAnnualExpenses / 12) + monthlyExpenses
	totalMonthlyBuyingCost := monthlyLoanPayment + monthlyRecurringExpenses

	// Calculate monthly cost for renting
	monthlyRentingExpenses := (annualRentCosts / 12) + (otherAnnualCosts / 12)
	totalMonthlyRentingCost := monthlyRent + monthlyRentingExpenses

	// Populate global cost arrays for projections (360 months = 30 years max)
	populateMonthlyCosts(360, monthlyLoanPayment, monthlyRecurringExpenses, totalMonths, totalMonthlyRentingCost)

	// Display results
	fmt.Println("\n=== Buying Details ===")
	fmt.Printf("Loan amount: %s\n", formatCurrency(loanAmount))
	fmt.Printf("Annual interest rate: %.2f%%\n", annualRate)
	fmt.Printf("Loan duration: %s months\n", formatNumber(totalMonths))
	fmt.Printf("\nMonthly loan payment: %s\n", formatCurrency(monthlyLoanPayment))
	fmt.Printf("Monthly recurring expenses: %s\n", formatCurrency(monthlyRecurringExpenses))
	fmt.Printf("Total monthly buying cost: %s\n", formatCurrency(totalMonthlyBuyingCost))

	fmt.Println("\n=== Renting Details ===")
	fmt.Printf("Rental deposit: %s\n", formatCurrency(rentDeposit))
	fmt.Printf("Monthly rent: %s\n", formatCurrency(monthlyRent))
	fmt.Printf("Monthly rent expenses: %s\n", formatCurrency(monthlyRentingExpenses))
	fmt.Printf("Total monthly renting cost: %s\n", formatCurrency(totalMonthlyRentingCost))

	// Display projections
	fmt.Println("\n=== Net Worth Projections: Buy vs Rent ===")
	displayComparisonTable(purchasePrice, downpayment, appreciationRate, totalMonths,
		rentDeposit, investmentReturnRate)
}

// getInputWithDefault prompts the user and reads string input with default value support
// Automatically uses the next default value from savedDefaults
func getInputWithDefault(prompt string) string {
	defaultValue := getDefault()

	if defaultValue != "" {
		fmt.Printf("%s[%s] ", prompt, defaultValue)
	} else {
		fmt.Print(prompt)
	}

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	// If empty and we have a default, use the default
	if input == "" && defaultValue != "" {
		input = defaultValue
	}

	return input
}

// getDefault returns the next default value and increments the counter
func getDefault() string {
	if defaultIndex < len(savedDefaults) {
		val := savedDefaults[defaultIndex]
		defaultIndex++
		return val
	}
	defaultIndex++
	return ""
}

// getFloatInputWithDefault prompts with a default and converts to float
func getFloatInputWithDefault(prompt string) (float64, string, error) {
	input := getInputWithDefault(prompt)
	value, err := parseAmount(input)
	return value, input, err
}

// getStringInputAndParseWithDefault prompts with a default and applies parser
func getStringInputAndParseWithDefault(prompt string, parser func(string) (int, error)) (int, string, error) {
	input := getInputWithDefault(prompt)
	value, err := parser(input)
	return value, input, err
}

// loadInputs loads previously saved inputs from file
func loadInputs() []string {
	data, err := os.ReadFile(inputsFile)
	if err != nil {
		return []string{}
	}

	var inputs []string
	err = json.Unmarshal(data, &inputs)
	if err != nil {
		return []string{}
	}

	return inputs
}

// saveInputs saves current inputs to file for next run
func saveInputs(inputs []string) {
	data, err := json.Marshal(inputs)
	if err != nil {
		return
	}

	os.WriteFile(inputsFile, data, 0644)
}

// parseAmount parses currency amounts with k, M, B suffixes
// Returns 0 for empty input
// Also handles % sign (strips it out)
func parseAmount(input string) (float64, error) {
	input = strings.ToLower(strings.TrimSpace(input))

	// Handle empty input - default to 0
	if input == "" {
		return 0, nil
	}

	// Remove % sign if present (for percentage inputs like "-10%")
	input = strings.TrimSuffix(input, "%")
	input = strings.TrimSpace(input)

	// Check for suffix
	multiplier := 1.0
	numStr := input

	if strings.HasSuffix(input, "k") {
		multiplier = 1000.0
		numStr = strings.TrimSuffix(input, "k")
	} else if strings.HasSuffix(input, "m") {
		multiplier = 1000000.0
		numStr = strings.TrimSuffix(input, "m")
	} else if strings.HasSuffix(input, "b") {
		multiplier = 1000000000.0
		numStr = strings.TrimSuffix(input, "b")
	}

	// Parse the numeric part
	value, err := strconv.ParseFloat(strings.TrimSpace(numStr), 64)
	if err != nil {
		return 0, err
	}

	return value * multiplier, nil
}

// getStringInputAndParse prompts the user and applies a parser function
func getStringInputAndParse(prompt string, parser func(string) (int, error)) (int, error) {
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return parser(strings.TrimSpace(input))
}

// formatCurrency formats a number as currency with commas and 2 decimal places
func formatCurrency(amount float64) string {
	// Handle negative numbers
	sign := ""
	if amount < 0 {
		sign = "-"
		amount = -amount
	}

	// Format with 2 decimal places
	formatted := fmt.Sprintf("%.2f", amount)
	parts := strings.Split(formatted, ".")

	// Add commas to the integer part
	intPart := parts[0]
	var result strings.Builder
	for i, digit := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(digit)
	}

	return fmt.Sprintf("%s$%s.%s", sign, result.String(), parts[1])
}

// formatNumber formats an integer with commas
func formatNumber(num int) string {
	numStr := strconv.Itoa(num)
	var result strings.Builder

	for i, digit := range numStr {
		if i > 0 && (len(numStr)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(digit)
	}

	return result.String()
}

// parseDuration parses duration strings like "5y6m", "30y", "6m"
func parseDuration(duration string) (int, error) {
	duration = strings.ToLower(duration)
	years := 0
	months := 0

	// Find 'y' for years
	yIndex := strings.Index(duration, "y")
	if yIndex != -1 {
		yearStr := duration[:yIndex]
		var err error
		years, err = strconv.Atoi(yearStr)
		if err != nil {
			return 0, fmt.Errorf("invalid year format")
		}
		duration = duration[yIndex+1:]
	}

	// Find 'm' for months
	mIndex := strings.Index(duration, "m")
	if mIndex != -1 {
		monthStr := duration[:mIndex]
		var err error
		months, err = strconv.Atoi(monthStr)
		if err != nil {
			return 0, fmt.Errorf("invalid month format")
		}
	}

	totalMonths := years*12 + months
	if totalMonths <= 0 {
		return 0, fmt.Errorf("duration must be greater than 0")
	}

	return totalMonths, nil
}

// calculateMonthlyPayment calculates the monthly payment using the amortization formula
// M = P * [r(1+r)^n] / [(1+r)^n - 1]
func calculateMonthlyPayment(principal, monthlyRate float64, months int) float64 {
	if monthlyRate == 0 {
		return principal / float64(months)
	}

	factor := math.Pow(1+monthlyRate, float64(months))
	monthlyPayment := principal * (monthlyRate * factor) / (factor - 1)
	return monthlyPayment
}

// displayComparisonTable displays buy vs rent net worth projections side-by-side
// Uses global monthlyBuyingCosts and monthlyRentingCosts arrays
func displayComparisonTable(purchasePrice, downpayment, appreciationRate float64, loanDuration int,
	rentDeposit, investmentReturnRate float64) {
	// Define standard periods
	standardPeriods := []struct {
		label  string
		months int
	}{
		{"1 year", 12},
		{"3 years", 36},
		{"5 years", 60},
		{"10 years", 120},
		{"20 years", 240},
		{"30 years", 360},
	}

	// Build the final list of periods, inserting loan term if needed
	periods := []struct {
		label  string
		months int
	}{}

	loanTermLabel := fmt.Sprintf("Loan term (%d years)", loanDuration/12)
	if loanDuration%12 != 0 {
		years := loanDuration / 12
		months := loanDuration % 12
		loanTermLabel = fmt.Sprintf("Loan term (%dy %dm)", years, months)
	}

	inserted := false
	for _, period := range standardPeriods {
		// Insert loan term before the first period that's longer
		if !inserted && loanDuration < period.months && loanDuration > 0 {
			periods = append(periods, struct {
				label  string
				months int
			}{loanTermLabel, loanDuration})
			inserted = true
		}

		// Skip if this period matches the loan duration
		if period.months == loanDuration {
			periods = append(periods, struct {
				label  string
				months int
			}{loanTermLabel, loanDuration})
			inserted = true
		} else {
			periods = append(periods, period)
		}
	}

	// If loan term is longer than all standard periods, add it at the end
	if !inserted && loanDuration > 0 {
		periods = append(periods, struct {
			label  string
			months int
		}{loanTermLabel, loanDuration})
	}

	// Print table header
	fmt.Printf("\n%-20s %-25s %-25s %-25s\n", "Period", "Buying Net Worth", "Renting Net Worth", "Difference")
	fmt.Println(strings.Repeat("-", 95))

	// Print each row
	for _, period := range periods {
		_, _, buyingNetWorth := calculateNetWorth(
			period.months, purchasePrice, downpayment, appreciationRate,
		)

		rentingNetWorth := calculateRentingNetWorth(
			period.months, downpayment, rentDeposit, investmentReturnRate,
		)

		difference := buyingNetWorth - rentingNetWorth

		fmt.Printf("%-20s %-25s %-25s %-25s\n",
			period.label,
			formatCurrency(buyingNetWorth),
			formatCurrency(rentingNetWorth),
			formatCurrency(difference),
		)
	}
}

// displayNetWorthTable displays net worth projections in a table format
// Uses global monthlyBuyingCosts array
func displayNetWorthTable(purchasePrice, downpayment, appreciationRate float64, loanDuration int) {
	// Define standard periods
	standardPeriods := []struct {
		label  string
		months int
	}{
		{"1 year", 12},
		{"3 years", 36},
		{"5 years", 60},
		{"10 years", 120},
		{"20 years", 240},
		{"30 years", 360},
	}

	// Build the final list of periods, inserting loan term if needed
	periods := []struct {
		label  string
		months int
	}{}

	loanTermLabel := fmt.Sprintf("Loan term (%d years)", loanDuration/12)
	if loanDuration%12 != 0 {
		years := loanDuration / 12
		months := loanDuration % 12
		loanTermLabel = fmt.Sprintf("Loan term (%dy %dm)", years, months)
	}

	inserted := false
	for _, period := range standardPeriods {
		// Insert loan term before the first period that's longer
		if !inserted && loanDuration < period.months && loanDuration > 0 {
			periods = append(periods, struct {
				label  string
				months int
			}{loanTermLabel, loanDuration})
			inserted = true
		}

		// Skip if this period matches the loan duration
		if period.months == loanDuration {
			periods = append(periods, struct {
				label  string
				months int
			}{loanTermLabel, loanDuration})
			inserted = true
		} else {
			periods = append(periods, period)
		}
	}

	// If loan term is longer than all standard periods, add it at the end
	if !inserted && loanDuration > 0 {
		periods = append(periods, struct {
			label  string
			months int
		}{loanTermLabel, loanDuration})
	}

	// Print table header
	fmt.Printf("\n%-20s %-20s %-20s %-20s\n", "Period", "Asset Value", "Total Expenditure", "Net Worth")
	fmt.Println(strings.Repeat("-", 80))

	// Print each row
	for _, period := range periods {
		assetValue, totalExpenditure, netWorth := calculateNetWorth(
			period.months, purchasePrice, downpayment, appreciationRate,
		)

		fmt.Printf("%-20s %-20s %-20s %-20s\n",
			period.label,
			formatCurrency(assetValue),
			formatCurrency(totalExpenditure),
			formatCurrency(netWorth),
		)
	}
}

// calculateNetWorth calculates the asset value, total expenditure, and net worth for a given time period
// Uses the global monthlyBuyingCosts array
func calculateNetWorth(months int, purchasePrice, downpayment, appreciationRate float64) (float64, float64, float64) {
	// Calculate years for appreciation (asset continues to appreciate beyond loan term)
	years := float64(months) / 12.0

	// Calculate asset value with appreciation/depreciation
	assetValue := purchasePrice * math.Pow(1+(appreciationRate/100), years)

	// Calculate total expenditure by summing monthly costs from array
	totalExpenditure := downpayment
	for i := 0; i < months; i++ {
		totalExpenditure += monthlyBuyingCosts[i]
	}

	// Calculate net worth
	netWorth := assetValue - totalExpenditure

	return assetValue, totalExpenditure, netWorth
}

// populateMonthlyCosts fills global arrays with monthly costs for buying and renting
func populateMonthlyCosts(maxMonths int, monthlyLoanPayment, monthlyRecurringExpenses float64, loanDuration int, monthlyRentingCost float64) {
	monthlyBuyingCosts = make([]float64, maxMonths)
	monthlyRentingCosts = make([]float64, maxMonths)

	for i := 0; i < maxMonths; i++ {
		// Renting cost is constant every month
		monthlyRentingCosts[i] = monthlyRentingCost

		// Buying cost: loan payment stops after loan duration, but recurring expenses continue
		if i < loanDuration {
			monthlyBuyingCosts[i] = monthlyLoanPayment + monthlyRecurringExpenses
		} else {
			// After loan is paid off, only recurring expenses remain
			monthlyBuyingCosts[i] = monthlyRecurringExpenses
		}
	}
}

// calculateRentingNetWorth calculates net worth for the renting scenario
// Uses month-by-month calculation: investment grows from downpayment + monthly savings
func calculateRentingNetWorth(months int, downpayment, rentDeposit, investmentReturnRate float64) float64 {
	// Start with downpayment minus deposit as initial investment
	investmentValue := downpayment - rentDeposit
	monthlyInvestmentRate := investmentReturnRate / 100 / 12

	// For each month: calculate savings, add to investment, grow investment
	for i := 0; i < months; i++ {
		// Monthly savings = buying cost - renting cost
		monthlySavings := monthlyBuyingCosts[i] - monthlyRentingCosts[i]

		// Add savings to investment
		investmentValue += monthlySavings

		// Apply monthly growth
		investmentValue *= (1 + monthlyInvestmentRate)
	}

	// Add back 75% of deposit (recoverable)
	recoverableDeposit := rentDeposit * 0.75

	return investmentValue + recoverableDeposit
}
