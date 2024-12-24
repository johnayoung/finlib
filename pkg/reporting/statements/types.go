package statements

import (
	"time"
	"github.com/johnayoung/finlib/pkg/money"
)

// StatementType represents the type of financial statement
type StatementType string

const (
	BalanceSheet    StatementType = "BALANCE_SHEET"
	IncomeStatement StatementType = "INCOME_STATEMENT"
	CashFlow        StatementType = "CASH_FLOW"
)

// LineItem represents a single line in a financial statement
type LineItem struct {
	// Label for the line item
	Label string `json:"label"`
	// Amount for the line item
	Amount money.Money `json:"amount"`
	// Account IDs that make up this line item
	AccountIDs []string `json:"account_ids"`
	// Optional sub-items for hierarchical statements
	SubItems []LineItem `json:"sub_items,omitempty"`
	// Optional metadata for custom attributes
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// StatementSection represents a section of a financial statement
type StatementSection struct {
	// Title of the section
	Title string `json:"title"`
	// Line items in this section
	Items []LineItem `json:"items"`
	// Total for this section
	Total money.Money `json:"total"`
}

// Statement represents a financial statement
type Statement struct {
	// Type of statement
	Type StatementType `json:"type"`
	// Title of the statement
	Title string `json:"title"`
	// Entity name
	Entity string `json:"entity"`
	// Statement date or period end
	AsOf time.Time `json:"as_of"`
	// Start of period (for income statement and cash flow)
	PeriodStart *time.Time `json:"period_start,omitempty"`
	// Sections of the statement
	Sections []StatementSection `json:"sections"`
	// Currency of the statement
	Currency string `json:"currency"`
	// Optional comparative period data
	ComparativePeriod *Statement `json:"comparative_period,omitempty"`
	// Additional metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// StatementOptions represents options for generating statements
type StatementOptions struct {
	// Include comparative period
	IncludeComparative bool
	// Comparative period length in months
	ComparativePeriodMonths int
	// Detail level (summary or detailed)
	DetailLevel string
	// Currency to display in
	Currency string
	// Custom account groupings
	AccountGroupings map[string][]string
	// Custom formatting options
	FormatOptions map[string]interface{}
}

// CashFlowCategory represents the category of cash flow
type CashFlowCategory string

const (
	Operating   CashFlowCategory = "OPERATING"
	Investing   CashFlowCategory = "INVESTING"
	Financing   CashFlowCategory = "FINANCING"
	Unclassified CashFlowCategory = "UNCLASSIFIED"
)

// CashFlowMethod represents the method used to calculate operating cash flow
type CashFlowMethod string

const (
	Direct   CashFlowMethod = "DIRECT"   // Shows operating cash receipts and payments
	Indirect CashFlowMethod = "INDIRECT" // Starts with net income and adjusts for non-cash items
)

// CashFlowClassification defines how transactions are classified for cash flow purposes
type CashFlowClassification struct {
	// Account ID this classification applies to
	AccountID string `json:"account_id"`
	// Category of cash flow this account belongs to
	Category CashFlowCategory `json:"category"`
	// Optional subcategory for more detailed reporting
	Subcategory string `json:"subcategory,omitempty"`
	// Whether to include this in working capital calculations
	IsWorkingCapital bool `json:"is_working_capital,omitempty"`
}

// CashFlowOptions represents options specific to cash flow statement generation
type CashFlowOptions struct {
	// Method to use for calculating operating cash flow
	Method CashFlowMethod `json:"method"`
	// Custom account classifications
	Classifications []CashFlowClassification `json:"classifications,omitempty"`
	// Whether to include non-cash transactions
	IncludeNonCash bool `json:"include_non_cash,omitempty"`
	// Whether to show working capital changes separately
	ShowWorkingCapital bool `json:"show_working_capital,omitempty"`
}
