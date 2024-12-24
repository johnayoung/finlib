package validation

import (
	"context"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/johnayoung/finlib/pkg/transaction"
)

// TransactionValidator implements basic transaction validation rules
type TransactionValidator struct {
	rules []ValidationRule
}

// NewTransactionValidator creates a new TransactionValidator
func NewTransactionValidator() *TransactionValidator {
	return &TransactionValidator{
		rules: []ValidationRule{
			{
				ID:          "TX_BALANCE",
				Description: "Transaction must be balanced (debits equal credits)",
				Severity:    Error,
				Category:    "TRANSACTION",
			},
			{
				ID:          "TX_MIN_ENTRIES",
				Description: "Transaction must have at least two entries",
				Severity:    Error,
				Category:    "TRANSACTION",
			},
			{
				ID:          "TX_DESCRIPTION",
				Description: "Transaction must have a description",
				Severity:    Warning,
				Category:    "TRANSACTION",
			},
		},
	}
}

// Validate performs validation on a transaction
func (v *TransactionValidator) Validate(ctx context.Context, obj interface{}) ([]ValidationResult, error) {
	tx, ok := obj.(*transaction.Transaction)
	if !ok {
		return nil, fmt.Errorf("expected *transaction.Transaction, got %T", obj)
	}

	var results []ValidationResult

	// Check minimum entries
	if len(tx.Entries) < 2 {
		results = append(results, ValidationResult{
			Code:     "TX_MIN_ENTRIES",
			Message:  "Transaction must have at least two entries",
			Severity: Error,
			Field:    "Entries",
		})
	}

	// Check description
	if tx.Description == "" {
		results = append(results, ValidationResult{
			Code:     "TX_DESCRIPTION",
			Message:  "Transaction should have a description",
			Severity: Warning,
			Field:    "Description",
		})
	}

	// Check balance
	if len(tx.Entries) > 0 {
		var debits, credits decimal.Decimal

		for _, entry := range tx.Entries {
			switch entry.Type {
			case transaction.Debit:
				debits = debits.Add(entry.Amount.Amount)
			case transaction.Credit:
				credits = credits.Add(entry.Amount.Amount)
			}
		}

		if !debits.Equal(credits) {
			results = append(results, ValidationResult{
				Code:    "TX_BALANCE",
				Message: fmt.Sprintf("Transaction is not balanced: debits=%s, credits=%s", debits, credits),
				Severity: Error,
				Field:   "Entries",
				Metadata: map[string]interface{}{
					"debits":  debits,
					"credits": credits,
				},
			})
		}
	}

	return results, nil
}

// GetRules returns the validation rules
func (v *TransactionValidator) GetRules() []ValidationRule {
	return v.rules
}

// Priority returns the validator priority (lower executes first)
func (v *TransactionValidator) Priority() int {
	return 100
}
