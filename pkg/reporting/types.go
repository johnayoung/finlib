// pkg/reporting/types.go

// Package reporting provides comprehensive financial reporting capabilities including
// balance sheets, income statements, cash flow statements, and custom reports.
// It supports various output formats, comparative reporting, and customizable
// calculations while maintaining strict financial accuracy.
package reporting

import (
	"context"
	"time"

	"github.com/johnayoung/finlib/pkg/account"
	"github.com/johnayoung/finlib/pkg/money"
	"github.com/shopspring/decimal"
)

// ReportType defines the standard types of financial reports available in the system.
// Custom report types can be defined for specific business needs.
type ReportType string

const (
	BalanceSheet     ReportType = "BALANCE_SHEET"     // Statement of financial position
	IncomeStatement  ReportType = "INCOME_STATEMENT"  // Profit and loss statement
	CashFlow         ReportType = "CASH_FLOW"         // Statement of cash flows
	GeneralLedger    ReportType = "GENERAL_LEDGER"    // Detailed transaction history
	TrialBalance     ReportType = "TRIAL_BALANCE"     // Pre-closing trial balance
	AccountStatement ReportType = "ACCOUNT_STATEMENT" // Individual account activity
	Custom           ReportType = "CUSTOM"            // User-defined report type
)

// BalanceType defines how account balances should be calculated for reporting purposes.
type BalanceType string

const (
	OpeningBalance BalanceType = "OPENING" // Balance at start of period
	ClosingBalance BalanceType = "CLOSING" // Balance at end of period
	AverageBalance BalanceType = "AVERAGE" // Average balance over period
)

// PeriodHandling defines how time periods should be handled in calculations and reporting.
type PeriodHandling struct {
	BalanceType    BalanceType            // How to calculate the balance
	Adjustment     PeriodAdjustment       // Any adjustments to apply to the period
	IncludeClosing bool                   // Whether to include closing entries
	Custom         map[string]interface{} // Custom period handling logic
}

// ReportPeriod represents a time period for financial reporting, supporting
// comparative analysis through the Previous field.
type ReportPeriod struct {
	Start    time.Time     // Start of reporting period
	End      time.Time     // End of reporting period
	Previous *ReportPeriod // Optional previous period for comparisons
}

// ReportOptions defines configuration options for report generation, allowing
// customization of output format, currency handling, and other parameters.
type ReportOptions struct {
	Period        ReportPeriod           // Time period for the report
	Currency      string                 // Currency for the report
	ShowCents     bool                   // Whether to include cents/decimal places
	Format        string                 // Report format (e.g., CSV, JSON)
	FormatOptions map[string]interface{} // Additional formatting options
	Parameters    map[string]interface{} // Custom parameters for specialized reports
}

// AccountSelector defines criteria for selecting accounts to include in reports
// or calculations.
type AccountSelector struct {
	Types      []account.AccountType // Account types to include
	Codes      []string              // Account codes to include
	Categories []string              // Account categories to include
	Tags       []string              // Account tags to include
	Expression string                // Custom selection logic
}

// AccountFilter defines criteria for filtering accounts based on various
// attributes and conditions.
type AccountFilter struct {
	Field       string          // The field to filter on
	Operator    string          // "EQUALS", "CONTAINS", "STARTS_WITH", etc.
	Value       interface{}     // The value to compare against
	SubFilters  []AccountFilter // For complex filters
	Combination string          // How to combine sub-filters ("AND", "OR")
}

// CalculationRule defines how values should be computed in a report, supporting
// both standard calculations and custom expressions.
type CalculationRule struct {
	ID          string                 // Rule identifier
	Name        string                 // Human-readable name
	Description string                 // Rule description
	Type        string                 // "SUM", "AVERAGE", "RATIO", "CUSTOM"
	Expression  string                 // For custom calculations
	Accounts    AccountSelector        // Account selection criteria
	Period      PeriodHandling         // Time period handling
	Parameters  map[string]interface{} // Additional parameters
}

// PeriodAdjustment defines how to adjust time periods for calculations,
// supporting various types of period shifts and adjustments.
type PeriodAdjustment struct {
	Type       string // "SHIFT", "EXTEND", "CUSTOM"
	Unit       string // "DAY", "MONTH", "YEAR"
	Amount     int    // The amount to adjust by
	Expression string // Custom adjustment logic
}

// ReportLine represents a single line in a report, supporting hierarchical
// structure and comparative amounts.
type ReportLine struct {
	AccountID      string                 // Account identifier
	AccountCode    string                 // Account code
	AccountName    string                 // Account name
	Amount         money.Money            // Primary amount
	PreviousAmount *money.Money           // Comparative amount
	Details        map[string]interface{} // Calculation details
	Level          int                    // Hierarchical level
	ParentID       string                 // Parent line identifier
	Children       []*ReportLine          // Child lines
}

// Report represents a generated financial report, containing both the report
// content and metadata about its generation.
type Report struct {
	ID          string                 // Report identifier
	Type        ReportType             // Type of report
	Title       string                 // Report title
	Period      ReportPeriod           // Time period
	Currency    string                 // Report currency
	GeneratedAt time.Time              // Generation timestamp
	GeneratedBy string                 // User or process that generated the report
	Lines       []*ReportLine          // Report content
	Totals      map[string]money.Money // Section/report totals
	Metadata    map[string]interface{} // Additional metadata
}

// ReportDefinition defines the structure and calculations for a report,
// providing a template that can be reused for generating reports.
type ReportDefinition struct {
	ID          string                 // Definition identifier
	Type        ReportType             // Report type
	Name        string                 // Definition name
	Description string                 // Definition description
	Sections    []ReportSection        // Report sections
	Rules       []CalculationRule      // Calculation rules
	Validations []ValidationRule       // Validation rules
	Format      FormatSpec             // Format specifications
	Extensions  map[string]interface{} // Plugin support
}

// ReportSection defines a section within a report, grouping related
// accounts and calculations.
type ReportSection struct {
	ID           string                // Section identifier
	Title        string                // Section title
	Description  string                // Section description
	AccountTypes []account.AccountType // Account types to include
	Filters      []AccountFilter       // Account filters
	Calculations []Calculation         // Calculations to perform
	Format       SectionFormat         // Section formatting
}

// Calculation defines how to compute values for a report section.
type Calculation struct {
	ID               string           // Calculation identifier
	Type             string           // "SUM", "AVERAGE", "CUSTOM"
	Expression       string           // Custom calculation expression
	AccountSelector  AccountSelector  // Account selection criteria
	PeriodAdjustment PeriodAdjustment // Period adjustment rules
	BalanceType      BalanceType      // Balance calculation type
}

// ValidationRule defines rules for validating report data and calculations.
type ValidationRule struct {
	ID          string // Rule identifier
	Description string // Rule description
	Expression  string // Validation expression
	Severity    string // "ERROR", "WARNING", "INFO"
}

// FormatSpec defines how a report should be formatted, including number
// formatting and column specifications.
type FormatSpec struct {
	DecimalPlaces  int          // Number of decimal places
	NegativeFormat string       // How to format negative numbers
	ShowZeroValues bool         // Whether to show zero values
	Columns        []ColumnSpec // Column specifications
}

// ColumnSpec defines the formatting and behavior of a report column.
type ColumnSpec struct {
	ID     string // Column identifier
	Title  string // Column title
	Type   string // "AMOUNT", "TEXT", "DATE", etc.
	Format string // Format string
	Width  int    // Column width
}

// SectionFormat defines the formatting options for a report section.
type SectionFormat struct {
	ShowHeader  bool   // Whether to show section header
	HeaderStyle string // Header formatting style
	ShowTotals  bool   // Whether to show section totals
	TotalStyle  string // Total line formatting style
	IndentLevel int    // Section indentation level
	LineStyle   string // Line item formatting style
}

// BalanceChange represents changes in account balances over a period.
type BalanceChange struct {
	OpeningBalance money.Money
	ClosingBalance money.Money
	NetChange      money.Money
	Movements      []BalanceMovement
}

// BalanceMovement represents a single change in an account balance.
type BalanceMovement struct {
	Date        time.Time
	Amount      money.Money
	Type        string // "DEBIT", "CREDIT"
	Description string
	Reference   string
}

// RatioDefinition defines how to calculate a financial ratio.
type RatioDefinition struct {
	ID          string      // Ratio identifier
	Name        string      // Ratio name
	Description string      // Ratio description
	Numerator   Calculation // Numerator calculation
	Denominator Calculation // Denominator calculation
	Scale       int32       // Number of decimal places
}

// ReportGenerator defines the interface for generating financial reports.
type ReportGenerator interface {
	// GenerateReport creates a report based on the definition and options
	GenerateReport(ctx context.Context, def *ReportDefinition, opts ReportOptions) (*Report, error)

	// ValidateDefinition checks if a report definition is valid
	ValidateDefinition(ctx context.Context, def *ReportDefinition) error

	// GetReportTypes returns available report types
	GetReportTypes(ctx context.Context) ([]ReportType, error)

	// SaveDefinition stores a report definition
	SaveDefinition(ctx context.Context, def *ReportDefinition) error

	// LoadDefinition retrieves a stored report definition
	LoadDefinition(ctx context.Context, id string) (*ReportDefinition, error)
}

// ReportFormatter handles the formatting of reports into various output formats.
type ReportFormatter interface {
	// FormatReport formats a report according to specified options
	FormatReport(ctx context.Context, report *Report, format string, opts map[string]interface{}) ([]byte, error)

	// GetSupportedFormats returns available output formats
	GetSupportedFormats(ctx context.Context) ([]string, error)
}

// ReportCalculator performs financial calculations for reports.
type ReportCalculator interface {
	// CalculateBalance computes account balances for reporting
	CalculateBalance(ctx context.Context, accountID string, period ReportPeriod) (money.Money, error)

	// CalculateChanges computes changes over a period
	CalculateChanges(ctx context.Context, accountID string, period ReportPeriod) (*BalanceChange, error)

	// CalculateRatio computes financial ratios
	CalculateRatio(ctx context.Context, ratio RatioDefinition, period ReportPeriod) (decimal.Decimal, error)
}
