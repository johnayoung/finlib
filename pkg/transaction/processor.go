package transaction

import (
	"context"
	"fmt"
	"time"

	"github.com/johnayoung/finlib/pkg/money"
	"github.com/johnayoung/finlib/pkg/storage"
)

// Common validation error codes
const (
	ErrCodeInvalidStatus       = "INVALID_STATUS"
	ErrCodeUnbalanced          = "UNBALANCED_TRANSACTION"
	ErrCodeInsufficientEntries = "INSUFFICIENT_ENTRIES"
	ErrCodeMixedCurrencies     = "MIXED_CURRENCIES"
	ErrCodeInvalidAmount       = "INVALID_AMOUNT"
	ErrCodeDuplicateAccount    = "DUPLICATE_ACCOUNT"
)

// TransactionProcessor handles the processing of financial transactions
type TransactionProcessor interface {
	// ValidateTransaction performs comprehensive validation of a transaction
	ValidateTransaction(ctx context.Context, tx *Transaction) (*ValidationResult, error)

	// ProcessTransaction processes a single transaction
	ProcessTransaction(ctx context.Context, tx *Transaction) error

	// ProcessTransactionBatch processes multiple transactions atomically
	ProcessTransactionBatch(ctx context.Context, txs []*Transaction) error

	// VoidTransaction voids a previously posted transaction
	VoidTransaction(ctx context.Context, txID string, reason string) error

	// ReverseTransaction creates and processes a reversal of a transaction
	ReverseTransaction(ctx context.Context, txID string, reason string) error

	// GetTransaction retrieves a transaction by ID
	GetTransaction(ctx context.Context, txID string) (*Transaction, error)

	// GetTransactionSummary calculates transaction totals
	GetTransactionSummary(ctx context.Context, tx *Transaction) (*TransactionSummary, error)
}

// Validator provides transaction validation logic
type Validator interface {
	// Validate performs the validation check
	Validate(ctx context.Context, tx *Transaction) (*ValidationResult, error)
}

// BasicValidator implements core transaction validation rules
type BasicValidator struct{}

// Validate implements the Validator interface
func (v *BasicValidator) Validate(ctx context.Context, tx *Transaction) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:    true,
		Errors:   make([]ValidationError, 0),
		Warnings: make([]ValidationError, 0),
	}

	// Check minimum entry requirement
	if len(tx.Entries) < 2 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrCodeInsufficientEntries,
			Message: "Transaction must have at least two entries",
		})
	}

	// Validate entry amounts and calculate totals
	var totalDebits, totalCredits money.Money
	seenAccounts := make(map[string]bool)
	var currency string

	for i, entry := range tx.Entries {
		// Check for zero amounts
		if entry.Amount.IsZero() {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrCodeInvalidAmount,
				Message: "Entry amount cannot be zero",
				Field:   fmt.Sprintf("Entries[%d].Amount", i),
			})
		}

		// Check for negative amounts
		if entry.Amount.IsNegative() {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrCodeInvalidAmount,
				Message: "Entry amount cannot be negative",
				Field:   fmt.Sprintf("Entries[%d].Amount", i),
			})
		}

		// Check for currency consistency
		if i == 0 {
			currency = entry.Amount.Currency
			totalDebits = entry.Amount
			totalCredits = entry.Amount
		} else if entry.Amount.Currency != currency {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrCodeMixedCurrencies,
				Message: "All entries must use the same currency",
				Field:   fmt.Sprintf("Entries[%d].Amount.Currency", i),
			})
		}

		// Track account usage
		if seenAccounts[entry.AccountID] {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrCodeDuplicateAccount,
				Message: "Account used multiple times in transaction",
				Field:   fmt.Sprintf("Entries[%d].AccountID", i),
			})
		}
		seenAccounts[entry.AccountID] = true

		// Update totals
		if entry.Type == Debit {
			totalDebits = money.Money{
				Amount:   totalDebits.Amount.Add(entry.Amount.Amount),
				Currency: currency,
			}
		} else {
			totalCredits = money.Money{
				Amount:   totalCredits.Amount.Add(entry.Amount.Amount),
				Currency: currency,
			}
		}
	}

	// Check if debits equal credits
	if !totalDebits.Amount.Equal(totalCredits.Amount) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrCodeUnbalanced,
			Message: "Total debits must equal total credits",
			Details: map[string]interface{}{
				"totalDebits":  totalDebits,
				"totalCredits": totalCredits,
			},
		})
	}

	return result, nil
}

// BasicTransactionProcessor provides a simple implementation of TransactionProcessor
type BasicTransactionProcessor struct {
	validator Validator
	repo      storage.Repository
}

// NewBasicTransactionProcessor creates a new BasicTransactionProcessor
func NewBasicTransactionProcessor(repo storage.Repository) *BasicTransactionProcessor {
	return &BasicTransactionProcessor{
		validator: &BasicValidator{},
		repo:      repo,
	}
}

// ValidateTransaction implements TransactionProcessor.ValidateTransaction
func (p *BasicTransactionProcessor) ValidateTransaction(ctx context.Context, tx *Transaction) (*ValidationResult, error) {
	return p.validator.Validate(ctx, tx)
}

// ProcessTransaction implements TransactionProcessor.ProcessTransaction
func (p *BasicTransactionProcessor) ProcessTransaction(ctx context.Context, tx *Transaction) error {
	// Validate the transaction
	result, err := p.ValidateTransaction(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to validate transaction: %w", err)
	}
	if !result.Valid {
		return fmt.Errorf("transaction validation failed: %v", result.Errors)
	}

	// Check if transaction can be processed
	if tx.Status != Draft && tx.Status != Pending {
		return fmt.Errorf("transaction must be in Draft or Pending status to process")
	}

	// Update transaction status and timestamps
	now := time.Now()
	tx.Status = Posted
	tx.PostedAt = &now
	tx.LastModified = now

	// Store the transaction
	err = p.repo.Update(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to store transaction: %w", err)
	}

	return nil
}

// ProcessTransactionBatch implements TransactionProcessor.ProcessTransactionBatch
func (p *BasicTransactionProcessor) ProcessTransactionBatch(ctx context.Context, txs []*Transaction) error {
	if len(txs) == 0 {
		return nil
	}

	// Pre-validate all transactions
	for _, tx := range txs {
		result, err := p.ValidateTransaction(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to validate transaction %s: %w", tx.ID, err)
		}
		if !result.Valid {
			return fmt.Errorf("transaction %s validation failed: %v", tx.ID, result.Errors)
		}

		// Check if transaction can be processed
		if tx.Status != Draft && tx.Status != Pending {
			return fmt.Errorf("transaction %s must be in Draft or Pending status to process", tx.ID)
		}
	}

	// Update all transaction statuses and timestamps
	now := time.Now()
	for _, tx := range txs {
		tx.Status = Posted
		tx.PostedAt = &now
		tx.LastModified = now
	}

	// Store all transactions
	// Note: The atomicity of the batch operation depends on the repository implementation
	for _, tx := range txs {
		err := p.repo.Update(ctx, tx)
		if err != nil {
			// If any transaction fails, we should roll back the entire batch
			// This is a simplified rollback - in a real system, we'd need proper transaction support
			for _, rtx := range txs {
				if rtx.Status == Posted {
					rtx.Status = Draft
					rtx.PostedAt = nil
					rtx.LastModified = time.Now()
					_ = p.repo.Update(ctx, rtx)
				}
			}
			return fmt.Errorf("failed to store transaction %s: %w", tx.ID, err)
		}
	}

	return nil
}

// GetTransaction implements TransactionProcessor.GetTransaction
func (p *BasicTransactionProcessor) GetTransaction(ctx context.Context, txID string) (*Transaction, error) {
	var tx Transaction
	err := p.repo.Read(ctx, txID, &tx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve transaction: %w", err)
	}
	return &tx, nil
}

// GetTransactionSummary implements TransactionProcessor.GetTransactionSummary
func (p *BasicTransactionProcessor) GetTransactionSummary(ctx context.Context, tx *Transaction) (*TransactionSummary, error) {
	summary := &TransactionSummary{
		EntryCount:       len(tx.Entries),
		AffectedAccounts: make([]string, 0, len(tx.Entries)),
	}

	for _, entry := range tx.Entries {
		summary.AffectedAccounts = append(summary.AffectedAccounts, entry.AccountID)
		if entry.Type == Debit {
			summary.TotalDebits = money.Money{
				Amount:   summary.TotalDebits.Amount.Add(entry.Amount.Amount),
				Currency: entry.Amount.Currency,
			}
		} else {
			summary.TotalCredits = money.Money{
				Amount:   summary.TotalCredits.Amount.Add(entry.Amount.Amount),
				Currency: entry.Amount.Currency,
			}
		}
	}

	summary.NetAmount = money.Money{
		Amount:   summary.TotalDebits.Amount.Sub(summary.TotalCredits.Amount),
		Currency: tx.Entries[0].Amount.Currency,
	}

	return summary, nil
}

// VoidTransaction implements TransactionProcessor.VoidTransaction
func (p *BasicTransactionProcessor) VoidTransaction(ctx context.Context, txID string, reason string) error {
	// Retrieve the transaction
	tx, err := p.GetTransaction(ctx, txID)
	if err != nil {
		return fmt.Errorf("failed to retrieve transaction: %w", err)
	}

	// Verify transaction can be voided
	if tx.Status != Posted {
		return fmt.Errorf("only posted transactions can be voided")
	}
	if tx.VoidedAt != nil {
		return fmt.Errorf("transaction is already voided")
	}

	// Update transaction status
	now := time.Now()
	tx.Status = Voided
	tx.VoidedAt = &now
	tx.VoidReason = reason
	tx.LastModified = now

	// Store the updated transaction
	err = p.repo.Update(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to store voided transaction: %w", err)
	}

	return nil
}

// ReverseTransaction implements TransactionProcessor.ReverseTransaction
func (p *BasicTransactionProcessor) ReverseTransaction(ctx context.Context, txID string, reason string) error {
	// Retrieve the original transaction
	origTx, err := p.GetTransaction(ctx, txID)
	if err != nil {
		return fmt.Errorf("failed to retrieve transaction: %w", err)
	}

	// Verify transaction can be reversed
	if origTx.Status != Posted {
		return fmt.Errorf("only posted transactions can be reversed")
	}
	if origTx.ReversedAt != nil {
		return fmt.Errorf("transaction is already reversed")
	}

	// Create reversal transaction
	now := time.Now()
	reversalTx := &Transaction{
		ID:           fmt.Sprintf("REV-%s", origTx.ID), // Prefix with REV for clarity
		Type:         Reversal,
		Status:       Draft,
		Date:         now,
		Description:  fmt.Sprintf("Reversal of %s: %s", origTx.ID, reason),
		Entries:      make([]Entry, len(origTx.Entries)),
		CreatedBy:    origTx.CreatedBy,
		Created:      now,
		LastModified: now,
		ReversedFrom: origTx.ID,
	}

	// Create reversed entries (swap debits and credits)
	for i, entry := range origTx.Entries {
		reversalTx.Entries[i] = Entry{
			AccountID:    entry.AccountID,
			Amount:       entry.Amount,
			Type:        entry.Type.Reverse(), // Swap debit/credit
			Description: fmt.Sprintf("Reversal of: %s", entry.Description),
		}
	}

	// Process the reversal transaction
	err = p.ProcessTransaction(ctx, reversalTx)
	if err != nil {
		return fmt.Errorf("failed to process reversal transaction: %w", err)
	}

	// Update original transaction
	origTx.ReversedAt = &now
	origTx.ReversalID = reversalTx.ID
	origTx.LastModified = now

	// Store the updated original transaction
	err = p.repo.Update(ctx, origTx)
	if err != nil {
		return fmt.Errorf("failed to update original transaction: %w", err)
	}

	return nil
}
