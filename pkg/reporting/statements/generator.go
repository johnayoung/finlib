package statements

import (
	"context"
	"fmt"
	"time"

	"github.com/johnayoung/finlib/pkg/account"
	"github.com/johnayoung/finlib/pkg/money"
	"github.com/johnayoung/finlib/pkg/reporting"
	"github.com/shopspring/decimal"
)

// Generator handles the generation of financial statements
type Generator struct {
	calculator reporting.ReportCalculator
	accounts   account.Repository
}

// NewGenerator creates a new statement generator
func NewGenerator(calculator reporting.ReportCalculator, accounts account.Repository) *Generator {
	return &Generator{
		calculator: calculator,
		accounts:   accounts,
	}
}

// GenerateBalanceSheet creates a balance sheet statement
func (g *Generator) GenerateBalanceSheet(ctx context.Context, asOf time.Time, opts StatementOptions) (*Statement, error) {
	// Create base statement
	stmt := &Statement{
		Type:     BalanceSheet,
		Title:    "Balance Sheet",
		AsOf:     asOf,
		Currency: opts.Currency,
		Sections: make([]StatementSection, 0),
	}

	// Generate assets section
	assetSection, err := g.generateBalanceSheetSection(ctx, "Assets", account.Asset, asOf, opts)
	if err != nil {
		return nil, fmt.Errorf("error generating assets section: %w", err)
	}
	stmt.Sections = append(stmt.Sections, assetSection)

	// Generate liabilities section
	liabSection, err := g.generateBalanceSheetSection(ctx, "Liabilities", account.Liability, asOf, opts)
	if err != nil {
		return nil, fmt.Errorf("error generating liabilities section: %w", err)
	}
	stmt.Sections = append(stmt.Sections, liabSection)

	// Generate equity section
	equitySection, err := g.generateBalanceSheetSection(ctx, "Equity", account.Equity, asOf, opts)
	if err != nil {
		return nil, fmt.Errorf("error generating equity section: %w", err)
	}
	stmt.Sections = append(stmt.Sections, equitySection)

	// Add comparative period if requested
	if opts.IncludeComparative {
		comparativePeriod := asOf.AddDate(0, -opts.ComparativePeriodMonths, 0)
		comparative, err := g.GenerateBalanceSheet(ctx, comparativePeriod, StatementOptions{
			Currency:    opts.Currency,
			DetailLevel: opts.DetailLevel,
		})
		if err != nil {
			return nil, fmt.Errorf("error generating comparative balance sheet: %w", err)
		}
		stmt.ComparativePeriod = comparative
	}

	return stmt, nil
}

// GenerateIncomeStatement creates an income statement
func (g *Generator) GenerateIncomeStatement(ctx context.Context, periodStart, periodEnd time.Time, opts StatementOptions) (*Statement, error) {
	stmt := &Statement{
		Type:        IncomeStatement,
		Title:       "Income Statement",
		AsOf:        periodEnd,
		PeriodStart: &periodStart,
		Currency:    opts.Currency,
		Sections:    make([]StatementSection, 0),
	}

	// Generate revenue section
	period := reporting.ReportPeriod{Start: periodStart, End: periodEnd}
	revenueSection, err := g.generateIncomeStatementSection(ctx, "Revenue", account.Revenue, period, opts)
	if err != nil {
		return nil, fmt.Errorf("error generating revenue section: %w", err)
	}
	stmt.Sections = append(stmt.Sections, revenueSection)

	// Generate expense section
	expenseSection, err := g.generateIncomeStatementSection(ctx, "Expenses", account.Expense, period, opts)
	if err != nil {
		return nil, fmt.Errorf("error generating expense section: %w", err)
	}
	stmt.Sections = append(stmt.Sections, expenseSection)

	// Add comparative period if requested
	if opts.IncludeComparative {
		periodLength := periodEnd.Sub(periodStart)
		comparativeEnd := periodStart
		comparativeStart := comparativeEnd.Add(-periodLength)
		comparative, err := g.GenerateIncomeStatement(ctx, comparativeStart, comparativeEnd, StatementOptions{
			Currency:    opts.Currency,
			DetailLevel: opts.DetailLevel,
		})
		if err != nil {
			return nil, fmt.Errorf("error generating comparative income statement: %w", err)
		}
		stmt.ComparativePeriod = comparative
	}

	return stmt, nil
}

// GenerateCashFlow creates a cash flow statement
func (g *Generator) GenerateCashFlow(ctx context.Context, periodStart, periodEnd time.Time, opts StatementOptions) (*Statement, error) {
	stmt := &Statement{
		Type:        CashFlow,
		Title:       "Statement of Cash Flows",
		AsOf:        periodEnd,
		PeriodStart: &periodStart,
		Currency:    opts.Currency,
		Sections:    make([]StatementSection, 0),
	}

	// Parse cash flow specific options
	cfOpts := CashFlowOptions{
		Method: Indirect, // Default to indirect method
	}
	if opts.FormatOptions != nil {
		if method, ok := opts.FormatOptions["method"].(string); ok {
			cfOpts.Method = CashFlowMethod(method)
		}
		if classifications, ok := opts.FormatOptions["classifications"].([]CashFlowClassification); ok {
			cfOpts.Classifications = classifications
		}
		if includeNonCash, ok := opts.FormatOptions["include_non_cash"].(bool); ok {
			cfOpts.IncludeNonCash = includeNonCash
		}
		if showWorkingCapital, ok := opts.FormatOptions["show_working_capital"].(bool); ok {
			cfOpts.ShowWorkingCapital = showWorkingCapital
		}
	}

	period := reporting.ReportPeriod{Start: periodStart, End: periodEnd}

	if cfOpts.Method == Indirect {
		// Generate operating activities section using indirect method
		operatingSection, err := g.generateOperatingCashFlowIndirect(ctx, period, opts)
		if err != nil {
			return nil, fmt.Errorf("error generating operating activities section: %w", err)
		}
		stmt.Sections = append(stmt.Sections, operatingSection)
	} else {
		// Generate operating activities section using direct method
		operatingSection, err := g.generateOperatingCashFlowDirect(ctx, period, opts)
		if err != nil {
			return nil, fmt.Errorf("error generating operating activities section: %w", err)
		}
		stmt.Sections = append(stmt.Sections, operatingSection)
	}

	// Generate investing activities section
	investingSection, err := g.generateInvestingCashFlow(ctx, period, opts)
	if err != nil {
		return nil, fmt.Errorf("error generating investing activities section: %w", err)
	}
	stmt.Sections = append(stmt.Sections, investingSection)

	// Generate financing activities section
	financingSection, err := g.generateFinancingCashFlow(ctx, period, opts)
	if err != nil {
		return nil, fmt.Errorf("error generating financing activities section: %w", err)
	}
	stmt.Sections = append(stmt.Sections, financingSection)

	// Add comparative period if requested
	if opts.IncludeComparative {
		periodLength := periodEnd.Sub(periodStart)
		comparativeEnd := periodStart
		comparativeStart := comparativeEnd.Add(-periodLength)
		comparative, err := g.GenerateCashFlow(ctx, comparativeStart, comparativeEnd, StatementOptions{
			Currency:    opts.Currency,
			DetailLevel: opts.DetailLevel,
		})
		if err != nil {
			return nil, fmt.Errorf("error generating comparative cash flow statement: %w", err)
		}
		stmt.ComparativePeriod = comparative
	}

	return stmt, nil
}

// Helper functions

func (g *Generator) generateBalanceSheetSection(ctx context.Context, title string, accountType account.AccountType, asOf time.Time, opts StatementOptions) (StatementSection, error) {
	section := StatementSection{
		Title: title,
		Items: make([]LineItem, 0),
	}

	// Get all accounts of this type
	accounts := make([]*account.Account, 0)
	if err := g.accounts.Query(ctx, account.Account{Type: accountType}, &accounts); err != nil {
		return section, fmt.Errorf("error querying accounts: %w", err)
	}

	// Calculate balance for each account
	total := decimal.Zero
	for _, acc := range accounts {
		balance, err := g.calculator.CalculateBalance(ctx, acc.ID, reporting.ReportPeriod{End: asOf})
		if err != nil {
			return section, fmt.Errorf("error calculating balance for account %s: %w", acc.ID, err)
		}

		if !balance.Amount.IsZero() || opts.DetailLevel == "detailed" {
			item := LineItem{
				Label:      acc.Name,
				Amount:     balance,
				AccountIDs: []string{acc.ID},
			}
			section.Items = append(section.Items, item)
			total = total.Add(balance.Amount)
		}
	}

	section.Total = money.Money{Amount: total, Currency: opts.Currency}
	return section, nil
}

func (g *Generator) generateIncomeStatementSection(ctx context.Context, title string, accountType account.AccountType, period reporting.ReportPeriod, opts StatementOptions) (StatementSection, error) {
	section := StatementSection{
		Title: title,
		Items: make([]LineItem, 0),
	}

	// Get all accounts of this type
	accounts := make([]*account.Account, 0)
	if err := g.accounts.Query(ctx, account.Account{Type: accountType}, &accounts); err != nil {
		return section, fmt.Errorf("error querying accounts: %w", err)
	}

	// Calculate changes for each account
	total := decimal.Zero
	for _, acc := range accounts {
		changes, err := g.calculator.CalculateChanges(ctx, acc.ID, period)
		if err != nil {
			return section, fmt.Errorf("error calculating changes for account %s: %w", acc.ID, err)
		}

		if !changes.NetChange.Amount.IsZero() || opts.DetailLevel == "detailed" {
			item := LineItem{
				Label:      acc.Name,
				Amount:     changes.NetChange,
				AccountIDs: []string{acc.ID},
			}
			section.Items = append(section.Items, item)
			total = total.Add(changes.NetChange.Amount)
		}
	}

	section.Total = money.Money{Amount: total, Currency: opts.Currency}
	return section, nil
}

func (g *Generator) generateOperatingCashFlowIndirect(ctx context.Context, period reporting.ReportPeriod, opts StatementOptions) (StatementSection, error) {
	section := StatementSection{
		Title: "Operating Activities",
		Items: make([]LineItem, 0),
	}

	// Start with net income
	revenueAccounts := make([]*account.Account, 0)
	expenseAccounts := make([]*account.Account, 0)
	if err := g.accounts.Query(ctx, account.Account{Type: account.Revenue}, &revenueAccounts); err != nil {
		return section, fmt.Errorf("error querying revenue accounts: %w", err)
	}
	if err := g.accounts.Query(ctx, account.Account{Type: account.Expense}, &expenseAccounts); err != nil {
		return section, fmt.Errorf("error querying expense accounts: %w", err)
	}

	// Calculate total revenue
	revenue := decimal.Zero
	for _, acc := range revenueAccounts {
		changes, err := g.calculator.CalculateChanges(ctx, acc.ID, period)
		if err != nil {
			return section, fmt.Errorf("error calculating revenue changes: %w", err)
		}
		revenue = revenue.Add(changes.NetChange.Amount)
	}

	// Calculate total expenses
	expenses := decimal.Zero
	for _, acc := range expenseAccounts {
		changes, err := g.calculator.CalculateChanges(ctx, acc.ID, period)
		if err != nil {
			return section, fmt.Errorf("error calculating expense changes: %w", err)
		}
		expenses = expenses.Add(changes.NetChange.Amount)
	}

	// Net income = revenue - expenses
	netIncome := revenue.Sub(expenses)
	section.Items = append(section.Items, LineItem{
		Label:  "Net Income",
		Amount: money.Money{Amount: netIncome, Currency: "USD"},
	})

	// Add back non-cash expenses
	nonCashAdjustments, err := g.calculateNonCashAdjustments(ctx, period)
	if err != nil {
		return section, fmt.Errorf("error calculating non-cash adjustments: %w", err)
	}
	section.Items = append(section.Items, nonCashAdjustments...)

	// Add changes in working capital
	workingCapitalChanges, err := g.calculateWorkingCapitalChanges(ctx, period)
	if err != nil {
		return section, fmt.Errorf("error calculating working capital changes: %w", err)
	}
	section.Items = append(section.Items, workingCapitalChanges...)

	// Calculate section total
	total := decimal.Zero
	for _, item := range section.Items {
		total = total.Add(item.Amount.Amount)
	}
	section.Total = money.Money{Amount: total, Currency: opts.Currency}

	return section, nil
}

func (g *Generator) generateOperatingCashFlowDirect(ctx context.Context, period reporting.ReportPeriod, opts StatementOptions) (StatementSection, error) {
	section := StatementSection{
		Title: "Operating Activities",
		Items: make([]LineItem, 0),
	}

	// Calculate cash receipts from customers
	receipts, err := g.calculateCashReceipts(ctx, period)
	if err != nil {
		return section, fmt.Errorf("error calculating cash receipts: %w", err)
	}
	section.Items = append(section.Items, receipts...)

	// Calculate cash payments
	payments, err := g.calculateCashPayments(ctx, period)
	if err != nil {
		return section, fmt.Errorf("error calculating cash payments: %w", err)
	}
	section.Items = append(section.Items, payments...)

	// Calculate section total
	total := decimal.Zero
	for _, item := range section.Items {
		total = total.Add(item.Amount.Amount)
	}
	section.Total = money.Money{Amount: total, Currency: opts.Currency}

	return section, nil
}

func (g *Generator) generateInvestingCashFlow(ctx context.Context, period reporting.ReportPeriod, opts StatementOptions) (StatementSection, error) {
	section := StatementSection{
		Title: "Investing Activities",
		Items: make([]LineItem, 0),
	}

	// Get all accounts classified as investing activities
	accounts := make([]*account.Account, 0)
	if err := g.accounts.Query(ctx, account.Account{Type: account.Asset}, &accounts); err != nil {
		return section, fmt.Errorf("error querying investing accounts: %w", err)
	}

	// Calculate changes for each investing account
	total := decimal.Zero
	for _, acc := range accounts {
		// TODO: Add logic to determine if this is an investing account
		changes, err := g.calculator.CalculateChanges(ctx, acc.ID, period)
		if err != nil {
			return section, fmt.Errorf("error calculating changes for account %s: %w", acc.ID, err)
		}

		if !changes.NetChange.Amount.IsZero() || opts.DetailLevel == "detailed" {
			item := LineItem{
				Label:      acc.Name,
				Amount:     changes.NetChange,
				AccountIDs: []string{acc.ID},
			}
			section.Items = append(section.Items, item)
			total = total.Add(changes.NetChange.Amount)
		}
	}

	section.Total = money.Money{Amount: total, Currency: opts.Currency}
	return section, nil
}

func (g *Generator) generateFinancingCashFlow(ctx context.Context, period reporting.ReportPeriod, opts StatementOptions) (StatementSection, error) {
	section := StatementSection{
		Title: "Financing Activities",
		Items: make([]LineItem, 0),
	}

	// Get all accounts classified as financing activities
	accounts := make([]*account.Account, 0)
	if err := g.accounts.Query(ctx, account.Account{Type: account.Liability}, &accounts); err != nil {
		return section, fmt.Errorf("error querying financing accounts: %w", err)
	}

	// Calculate changes for each financing account
	total := decimal.Zero
	for _, acc := range accounts {
		// TODO: Add logic to determine if this is a financing account
		changes, err := g.calculator.CalculateChanges(ctx, acc.ID, period)
		if err != nil {
			return section, fmt.Errorf("error calculating changes for account %s: %w", acc.ID, err)
		}

		if !changes.NetChange.Amount.IsZero() || opts.DetailLevel == "detailed" {
			item := LineItem{
				Label:      acc.Name,
				Amount:     changes.NetChange,
				AccountIDs: []string{acc.ID},
			}
			section.Items = append(section.Items, item)
			total = total.Add(changes.NetChange.Amount)
		}
	}

	section.Total = money.Money{Amount: total, Currency: opts.Currency}
	return section, nil
}

func (g *Generator) calculateNetIncome(ctx context.Context, period reporting.ReportPeriod) (money.Money, error) {
	revenueAccounts := make([]*account.Account, 0)
	expenseAccounts := make([]*account.Account, 0)
	if err := g.accounts.Query(ctx, account.Account{Type: account.Revenue}, &revenueAccounts); err != nil {
		return money.Money{}, fmt.Errorf("error querying revenue accounts: %w", err)
	}
	if err := g.accounts.Query(ctx, account.Account{Type: account.Expense}, &expenseAccounts); err != nil {
		return money.Money{}, fmt.Errorf("error querying expense accounts: %w", err)
	}

	// Calculate total revenue
	revenue := decimal.Zero
	for _, acc := range revenueAccounts {
		changes, err := g.calculator.CalculateChanges(ctx, acc.ID, period)
		if err != nil {
			return money.Money{}, fmt.Errorf("error calculating revenue changes: %w", err)
		}
		revenue = revenue.Add(changes.NetChange.Amount)
	}

	// Calculate total expenses
	expenses := decimal.Zero
	for _, acc := range expenseAccounts {
		changes, err := g.calculator.CalculateChanges(ctx, acc.ID, period)
		if err != nil {
			return money.Money{}, fmt.Errorf("error calculating expense changes: %w", err)
		}
		expenses = expenses.Add(changes.NetChange.Amount)
	}

	// Net income = revenue - expenses
	netIncome := revenue.Sub(expenses)
	return money.Money{Amount: netIncome, Currency: "USD"}, nil
}

func (g *Generator) calculateNonCashAdjustments(ctx context.Context, period reporting.ReportPeriod) ([]LineItem, error) {
	items := make([]LineItem, 0)
	// TODO: Implement non-cash adjustments calculation
	// This should include:
	// - Depreciation and amortization
	// - Stock-based compensation
	// - Gains/losses on asset sales
	// - Other non-cash items
	return items, nil
}

func (g *Generator) calculateWorkingCapitalChanges(ctx context.Context, period reporting.ReportPeriod) ([]LineItem, error) {
	items := make([]LineItem, 0)
	// TODO: Implement working capital changes calculation
	// This should include changes in:
	// - Accounts receivable
	// - Inventory
	// - Accounts payable
	// - Other current assets and liabilities
	return items, nil
}

func (g *Generator) calculateCashReceipts(ctx context.Context, period reporting.ReportPeriod) ([]LineItem, error) {
	items := make([]LineItem, 0)
	// TODO: Implement cash receipts calculation
	// This should include:
	// - Cash received from customers
	// - Interest received
	// - Other operating receipts
	return items, nil
}

func (g *Generator) calculateCashPayments(ctx context.Context, period reporting.ReportPeriod) ([]LineItem, error) {
	items := make([]LineItem, 0)
	// TODO: Implement cash payments calculation
	// This should include:
	// - Payments to suppliers
	// - Payments to employees
	// - Interest paid
	// - Taxes paid
	// - Other operating payments
	return items, nil
}
