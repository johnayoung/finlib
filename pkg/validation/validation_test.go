package validation

import (
	"context"
	"github.com/shopspring/decimal"
	"github.com/johnayoung/finlib/pkg/money"
	"github.com/johnayoung/finlib/pkg/transaction"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTransactionValidator(t *testing.T) {
	ctx := context.Background()
	validator := NewTransactionValidator()

	t.Run("Valid Transaction", func(t *testing.T) {
		tx := &transaction.Transaction{
			ID:          "TX001",
			Date:        time.Now(),
			Description: "Test Transaction",
			Entries: []transaction.Entry{
				{
					AccountID:   "ACC001",
					Amount:     money.Money{Amount: decimal.NewFromInt(100), Currency: "USD"},
					Type:       transaction.Debit,
					Description: "Debit entry",
				},
				{
					AccountID:   "ACC002",
					Amount:     money.Money{Amount: decimal.NewFromInt(100), Currency: "USD"},
					Type:       transaction.Credit,
					Description: "Credit entry",
				},
			},
		}

		results, err := validator.Validate(ctx, tx)
		assert.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("Unbalanced Transaction", func(t *testing.T) {
		tx := &transaction.Transaction{
			ID:          "TX002",
			Date:        time.Now(),
			Description: "Unbalanced Transaction",
			Entries: []transaction.Entry{
				{
					AccountID:   "ACC001",
					Amount:     money.Money{Amount: decimal.NewFromInt(100), Currency: "USD"},
					Type:       transaction.Debit,
					Description: "Debit entry",
				},
				{
					AccountID:   "ACC002",
					Amount:     money.Money{Amount: decimal.NewFromInt(90), Currency: "USD"},
					Type:       transaction.Credit,
					Description: "Credit entry",
				},
			},
		}

		results, err := validator.Validate(ctx, tx)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "TX_BALANCE", results[0].Code)
		assert.Equal(t, Error, results[0].Severity)
	})

	t.Run("Missing Description", func(t *testing.T) {
		tx := &transaction.Transaction{
			ID:   "TX003",
			Date: time.Now(),
			Entries: []transaction.Entry{
				{
					AccountID:   "ACC001",
					Amount:     money.Money{Amount: decimal.NewFromInt(100), Currency: "USD"},
					Type:       transaction.Debit,
					Description: "Debit entry",
				},
				{
					AccountID:   "ACC002",
					Amount:     money.Money{Amount: decimal.NewFromInt(100), Currency: "USD"},
					Type:       transaction.Credit,
					Description: "Credit entry",
				},
			},
		}

		results, err := validator.Validate(ctx, tx)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "TX_DESCRIPTION", results[0].Code)
		assert.Equal(t, Warning, results[0].Severity)
	})

	t.Run("Invalid Object Type", func(t *testing.T) {
		results, err := validator.Validate(ctx, "not a transaction")
		assert.Error(t, err)
		assert.Nil(t, results)
	})
}

func TestBasicValidationEngine(t *testing.T) {
	ctx := context.Background()
	engine := NewBasicValidationEngine()
	
	t.Run("Register and Run Validator", func(t *testing.T) {
		validator := NewTransactionValidator()
		err := engine.RegisterValidator(validator)
		assert.NoError(t, err)

		tx := &transaction.Transaction{
			ID:          "TX001",
			Date:        time.Now(),
			Description: "Test Transaction",
			Entries: []transaction.Entry{
				{
					AccountID:   "ACC001",
					Amount:     money.Money{Amount: decimal.NewFromInt(100), Currency: "USD"},
					Type:       transaction.Debit,
					Description: "Debit entry",
				},
				{
					AccountID:   "ACC002",
					Amount:     money.Money{Amount: decimal.NewFromInt(100), Currency: "USD"},
					Type:       transaction.Credit,
					Description: "Credit entry",
				},
			},
		}

		results, err := engine.Validate(ctx, tx)
		assert.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("Register Nil Validator", func(t *testing.T) {
		err := engine.RegisterValidator(nil)
		assert.Error(t, err)
	})

	t.Run("Validate Nil Object", func(t *testing.T) {
		results, err := engine.Validate(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, results)
	})
}
