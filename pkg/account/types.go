package account

import (
	"time"
	"github.com/johnayoung/finlib/pkg/money"
)

// AccountType represents the classification of an account
type AccountType string

const (
	Asset     AccountType = "ASSET"
	Liability AccountType = "LIABILITY"
	Equity    AccountType = "EQUITY"
	Revenue   AccountType = "REVENUE"
	Expense   AccountType = "EXPENSE"
)

// AccountStatus represents the status of an account
type AccountStatus string

const (
	Active   AccountStatus = "ACTIVE"
	Inactive AccountStatus = "INACTIVE"
	Closed   AccountStatus = "CLOSED"
	Frozen   AccountStatus = "FROZEN"
)

// Account represents a financial account in the system
type Account struct {
	// Unique identifier for the account
	ID string
	// Account code used for reporting and categorization
	Code string
	// Human-readable name of the account
	Name string
	// Type of account (Asset, Liability, etc.)
	Type AccountType
	// Status of the account
	Status AccountStatus
	// Optional parent account ID for hierarchical structures
	ParentID *string
	// When the account was created
	Created time.Time
	// Last modification timestamp
	LastModified time.Time
	// Additional metadata for extensibility
	MetaData map[string]interface{}
	// Balance of the account
	Balance *money.Money
}

// Status represents the current state of an account
type Status struct {
	// Whether the account is active
	Active bool
	// Whether the account is locked for modifications
	Locked bool
	// Reason for current status
	StatusReason string
	// When the status was last updated
	LastUpdated time.Time
}

// ValidationRule represents a rule that must be satisfied for account operations
type ValidationRule struct {
	// Unique identifier for the rule
	ID string
	// Human-readable description of the rule
	Description string
	// Type of rule (e.g., "balance", "transaction")
	Type string
	// Whether rule violation blocks operations
	Blocking bool
}

// Balance represents the current balance of an account
type Balance struct {
	// Account ID this balance belongs to
	AccountID string
	// Timestamp of the balance
	AsOf time.Time
	// Actual balance amount and currency
	Amount string
	Currency string
	// Last transaction ID that affected this balance
	LastTransactionID string
}
