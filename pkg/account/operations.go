package account

import (
	"context"
	"errors"
)

var (
	ErrAccountNotFound     = errors.New("account not found")
	ErrInvalidAccountType  = errors.New("invalid account type")
	ErrAccountLocked       = errors.New("account is locked")
	ErrInvalidOperation    = errors.New("invalid operation")
	ErrInvalidAccountCode  = errors.New("invalid account code")
)

// AccountManager defines the interface for account operations
type AccountManager interface {
	// CreateAccount creates a new account
	CreateAccount(ctx context.Context, account *Account) error

	// GetAccount retrieves an account by ID
	GetAccount(ctx context.Context, id string) (*Account, error)

	// UpdateAccount updates an existing account
	UpdateAccount(ctx context.Context, account *Account) error

	// DeleteAccount marks an account as deleted
	DeleteAccount(ctx context.Context, id string) error

	// GetAccountStatus retrieves the current status of an account
	GetAccountStatus(ctx context.Context, id string) (*Status, error)

	// SetAccountStatus updates the status of an account
	SetAccountStatus(ctx context.Context, id string, status *Status) error

	// GetAccountBalance retrieves the current balance of an account
	GetAccountBalance(ctx context.Context, id string) (*Balance, error)

	// ValidateAccount performs validation checks on an account
	ValidateAccount(ctx context.Context, account *Account) error

	// ListAccounts retrieves accounts based on filters
	ListAccounts(ctx context.Context, filters map[string]interface{}) ([]*Account, error)
}

// ValidationManager defines the interface for account validation operations
type ValidationManager interface {
	// AddValidationRule adds a new validation rule
	AddValidationRule(ctx context.Context, rule *ValidationRule) error

	// RemoveValidationRule removes a validation rule
	RemoveValidationRule(ctx context.Context, ruleID string) error

	// GetValidationRules retrieves all validation rules for an account type
	GetValidationRules(ctx context.Context, accountType AccountType) ([]*ValidationRule, error)

	// ValidateOperation checks if an operation is allowed
	ValidateOperation(ctx context.Context, accountID string, operation string) error
}
