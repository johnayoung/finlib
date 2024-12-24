package reporting

import (
	"context"
	"fmt"
	"time"

	"github.com/johnayoung/finlib/pkg/account"
	"github.com/johnayoung/finlib/pkg/money"
)

// defaultReportGenerator implements the ReportGenerator interface
type defaultReportGenerator struct {
	calculator ReportCalculator
	storage    ReportStorage
}

// NewReportGenerator creates a new instance of the default report generator
func NewReportGenerator(calculator ReportCalculator, storage ReportStorage) ReportGenerator {
	return &defaultReportGenerator{
		calculator: calculator,
		storage:    storage,
	}
}

// GenerateReport creates a report based on the definition and options
func (g *defaultReportGenerator) GenerateReport(ctx context.Context, def *ReportDefinition, opts ReportOptions) (*Report, error) {
	if err := g.ValidateDefinition(ctx, def); err != nil {
		return nil, fmt.Errorf("invalid report definition: %w", err)
	}

	report := &Report{
		ID:          generateReportID(),
		Type:        def.Type,
		Title:       def.Name,
		Period:      opts.Period,
		Currency:    opts.Currency,
		GeneratedAt: time.Now(),
		Lines:       make([]*ReportLine, 0),
		Totals:      make(map[string]money.Money),
		Metadata:    make(map[string]interface{}),
	}

	// Process each section in the report definition
	for _, section := range def.Sections {
		if err := g.processSection(ctx, report, &section, opts); err != nil {
			return nil, fmt.Errorf("error processing section %s: %w", section.ID, err)
		}
	}

	// Apply any report-level calculations
	if err := g.applyCalculations(ctx, report, def.Rules, opts); err != nil {
		return nil, fmt.Errorf("error applying calculations: %w", err)
	}

	// Validate the generated report
	if err := g.validateReport(ctx, report, def.Validations); err != nil {
		return nil, fmt.Errorf("report validation failed: %w", err)
	}

	return report, nil
}

// ValidateDefinition checks if a report definition is valid
func (g *defaultReportGenerator) ValidateDefinition(ctx context.Context, def *ReportDefinition) error {
	if def == nil {
		return fmt.Errorf("report definition cannot be nil")
	}

	if def.Type == "" {
		return fmt.Errorf("report type is required")
	}

	if def.Name == "" {
		return fmt.Errorf("report name is required")
	}

	if len(def.Sections) == 0 {
		return fmt.Errorf("report must have at least one section")
	}

	// Validate each section
	for _, section := range def.Sections {
		if err := g.validateSection(&section); err != nil {
			return fmt.Errorf("invalid section %s: %w", section.ID, err)
		}
	}

	return nil
}

// GetReportTypes returns available report types
func (g *defaultReportGenerator) GetReportTypes(ctx context.Context) ([]ReportType, error) {
	return []ReportType{
		BalanceSheet,
		IncomeStatement,
		CashFlow,
		GeneralLedger,
		TrialBalance,
		AccountStatement,
		Custom,
	}, nil
}

// SaveDefinition stores a report definition
func (g *defaultReportGenerator) SaveDefinition(ctx context.Context, def *ReportDefinition) error {
	if err := g.ValidateDefinition(ctx, def); err != nil {
		return fmt.Errorf("invalid report definition: %w", err)
	}
	return g.storage.SaveDefinition(ctx, def)
}

// LoadDefinition retrieves a stored report definition
func (g *defaultReportGenerator) LoadDefinition(ctx context.Context, id string) (*ReportDefinition, error) {
	return g.storage.LoadDefinition(ctx, id)
}

// processSection processes a single section of the report
func (g *defaultReportGenerator) processSection(ctx context.Context, report *Report, section *ReportSection, opts ReportOptions) error {
	// Get accounts for this section based on types and filters
	accounts, err := g.getAccountsForSection(ctx, section)
	if err != nil {
		return fmt.Errorf("error getting accounts: %w", err)
	}

	// Process each account and create report lines
	for _, acc := range accounts {
		line, err := g.createReportLine(ctx, acc, section, opts)
		if err != nil {
			return fmt.Errorf("error creating report line for account %s: %w", acc.ID, err)
		}
		report.Lines = append(report.Lines, line)
	}

	// Apply section-specific calculations
	if err := g.applySectionCalculations(ctx, report, section, opts); err != nil {
		return fmt.Errorf("error applying section calculations: %w", err)
	}

	return nil
}

// validateSection validates a single section definition
func (g *defaultReportGenerator) validateSection(section *ReportSection) error {
	if section.ID == "" {
		return fmt.Errorf("section ID is required")
	}

	if section.Title == "" {
		return fmt.Errorf("section title is required")
	}

	if len(section.AccountTypes) == 0 && len(section.Filters) == 0 {
		return fmt.Errorf("section must specify either account types or filters")
	}

	return nil
}

// validateReport validates the generated report against the specified validation rules
func (g *defaultReportGenerator) validateReport(ctx context.Context, report *Report, rules []ValidationRule) error {
	for _, rule := range rules {
		if err := g.applyValidationRule(ctx, report, &rule); err != nil {
			return fmt.Errorf("validation rule %s failed: %w", rule.ID, err)
		}
	}
	return nil
}

// Helper functions

func generateReportID() string {
	return fmt.Sprintf("RPT_%d", time.Now().UnixNano())
}

// getAccountsForSection retrieves accounts based on section criteria
func (g *defaultReportGenerator) getAccountsForSection(ctx context.Context, section *ReportSection) ([]*account.Account, error) {
	// This is a placeholder implementation that returns an error
	return nil, fmt.Errorf("getAccountsForSection not implemented")
}

// createReportLine creates a report line for an account
func (g *defaultReportGenerator) createReportLine(ctx context.Context, acc *account.Account, section *ReportSection, opts ReportOptions) (*ReportLine, error) {
	// Calculate the balance for the account
	balance, err := g.calculator.CalculateBalance(ctx, acc.ID, opts.Period)
	if err != nil {
		return nil, fmt.Errorf("error calculating balance: %w", err)
	}

	line := &ReportLine{
		AccountID:   acc.ID,
		AccountCode: acc.Code,
		AccountName: acc.Name,
		Amount:      balance,
		Details:     make(map[string]interface{}),
	}

	// If comparative reporting is enabled, calculate previous period
	if opts.Period.Previous != nil {
		prevBalance, err := g.calculator.CalculateBalance(ctx, acc.ID, *opts.Period.Previous)
		if err != nil {
			return nil, fmt.Errorf("error calculating previous balance: %w", err)
		}
		line.PreviousAmount = &prevBalance
	}

	return line, nil
}

// applyCalculations applies report-level calculations
func (g *defaultReportGenerator) applyCalculations(ctx context.Context, report *Report, rules []CalculationRule, opts ReportOptions) error {
	for _, rule := range rules {
		if err := g.applyCalculationRule(ctx, report, &rule, opts); err != nil {
			return fmt.Errorf("error applying calculation rule %s: %w", rule.ID, err)
		}
	}
	return nil
}

// applySectionCalculations applies section-specific calculations
func (g *defaultReportGenerator) applySectionCalculations(ctx context.Context, report *Report, section *ReportSection, opts ReportOptions) error {
	for _, calc := range section.Calculations {
		if err := g.applyCalculation(ctx, report, &calc, opts); err != nil {
			return fmt.Errorf("error applying calculation %s: %w", calc.ID, err)
		}
	}
	return nil
}

// applyCalculationRule applies a single calculation rule
func (g *defaultReportGenerator) applyCalculationRule(ctx context.Context, report *Report, rule *CalculationRule, opts ReportOptions) error {
	// Implementation would depend on the specific calculation types supported
	// This is a placeholder that would need to be implemented
	return nil
}

// applyCalculation applies a single calculation
func (g *defaultReportGenerator) applyCalculation(ctx context.Context, report *Report, calc *Calculation, opts ReportOptions) error {
	// Implementation would depend on the specific calculation types supported
	// This is a placeholder that would need to be implemented
	return nil
}

// applyValidationRule applies a single validation rule
func (g *defaultReportGenerator) applyValidationRule(ctx context.Context, report *Report, rule *ValidationRule) error {
	// Implementation would depend on the specific validation types supported
	// This is a placeholder that would need to be implemented
	return nil
}
