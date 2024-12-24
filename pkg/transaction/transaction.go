package transaction

import (
	"time"

	"github.com/johnayoung/finlib/pkg/money"
)

// EntryType represents the type of transaction entry (debit or credit)
type EntryType string

const (
	Debit  EntryType = "DEBIT"
	Credit EntryType = "CREDIT"
)

// Reverse returns the opposite entry type
func (e EntryType) Reverse() EntryType {
	if e == Debit {
		return Credit
	}
	return Debit
}

// TransactionType represents the type of transaction
type TransactionType string

const (
	Journal  TransactionType = "JOURNAL"
	Transfer TransactionType = "TRANSFER"
	Reversal TransactionType = "REVERSAL"
)

// TransactionStatus represents the current status of a transaction
type TransactionStatus string

const (
	Draft   TransactionStatus = "DRAFT"
	Pending TransactionStatus = "PENDING"
	Posted  TransactionStatus = "POSTED"
	Voided  TransactionStatus = "VOIDED"
)

// Entry represents a single entry in a transaction
type Entry struct {
	AccountID   string      `json:"account_id"`
	Amount      money.Money `json:"amount"`
	Type        EntryType   `json:"type"`
	Description string      `json:"description"`
}

// Transaction represents a financial transaction
type Transaction struct {
	ID           string            `json:"id"`
	Type         TransactionType   `json:"type"`
	Status       TransactionStatus `json:"status"`
	Date         time.Time         `json:"date"`
	Description  string            `json:"description"`
	Entries      []Entry           `json:"entries"`
	CreatedBy    string            `json:"created_by"`
	Created      time.Time         `json:"created"`
	LastModified time.Time         `json:"last_modified"`
	PostedAt     *time.Time        `json:"posted_at,omitempty"`
	VoidedAt     *time.Time        `json:"voided_at,omitempty"`
	VoidReason   string            `json:"void_reason,omitempty"`
	ReversedAt   *time.Time        `json:"reversed_at,omitempty"`
	ReversalID   string            `json:"reversal_id,omitempty"`
	ReversedFrom string            `json:"reversed_from,omitempty"`
}

// ValidationError represents a single validation error
type ValidationError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Field   string                 `json:"field,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ValidationResult represents the result of transaction validation
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []ValidationError `json:"warnings,omitempty"`
}

// TransactionSummary provides a summary of transaction totals
type TransactionSummary struct {
	TotalDebits      money.Money `json:"total_debits"`
	TotalCredits     money.Money `json:"total_credits"`
	NetAmount        money.Money `json:"net_amount"`
	EntryCount       int         `json:"entry_count"`
	AffectedAccounts []string    `json:"affected_accounts"`
}
