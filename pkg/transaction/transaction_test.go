package transaction

import (
	"context"
	"testing"
	"time"

	"github.com/johnayoung/finlib/pkg/money"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestTransactionValidation(t *testing.T) {
	validator := &BasicValidator{}
	ctx := context.Background()

	t.Run("Valid Transaction", func(t *testing.T) {
		tx := createValidTransaction()
		result, err := validator.Validate(ctx, tx)
		assert.NoError(t, err)
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
	})

	t.Run("Insufficient Entries", func(t *testing.T) {
		tx := &Transaction{
			ID:     "TX001",
			Type:   Journal,
			Status: Draft,
			Entries: []Entry{
				{
					AccountID: "ACC001",
					Amount:    money.Money{Amount: decimal.NewFromFloat(100), Currency: "USD"},
					Type:      Debit,
				},
			},
		}

		result, err := validator.Validate(ctx, tx)
		assert.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Equal(t, ErrCodeInsufficientEntries, result.Errors[0].Code)
	})

	t.Run("Unbalanced Transaction", func(t *testing.T) {
		tx := &Transaction{
			ID:     "TX001",
			Type:   Journal,
			Status: Draft,
			Entries: []Entry{
				{
					AccountID: "ACC001",
					Amount:    money.Money{Amount: decimal.NewFromFloat(100), Currency: "USD"},
					Type:      Debit,
				},
				{
					AccountID: "ACC002",
					Amount:    money.Money{Amount: decimal.NewFromFloat(90), Currency: "USD"},
					Type:      Credit,
				},
			},
		}

		result, err := validator.Validate(ctx, tx)
		assert.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Equal(t, ErrCodeUnbalanced, result.Errors[0].Code)
	})

	t.Run("Mixed Currencies", func(t *testing.T) {
		tx := &Transaction{
			ID:     "TX001",
			Type:   Journal,
			Status: Draft,
			Entries: []Entry{
				{
					AccountID: "ACC001",
					Amount:    money.Money{Amount: decimal.NewFromFloat(100), Currency: "USD"},
					Type:      Debit,
				},
				{
					AccountID: "ACC002",
					Amount:    money.Money{Amount: decimal.NewFromFloat(100), Currency: "EUR"},
					Type:      Credit,
				},
			},
		}

		result, err := validator.Validate(ctx, tx)
		assert.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Equal(t, ErrCodeMixedCurrencies, result.Errors[0].Code)
	})

	t.Run("Duplicate Accounts", func(t *testing.T) {
		tx := &Transaction{
			ID:     "TX001",
			Type:   Journal,
			Status: Draft,
			Entries: []Entry{
				{
					AccountID: "ACC001",
					Amount:    money.Money{Amount: decimal.NewFromFloat(100), Currency: "USD"},
					Type:      Debit,
				},
				{
					AccountID: "ACC001",
					Amount:    money.Money{Amount: decimal.NewFromFloat(100), Currency: "USD"},
					Type:      Credit,
				},
			},
		}

		result, err := validator.Validate(ctx, tx)
		assert.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Equal(t, ErrCodeDuplicateAccount, result.Errors[0].Code)
	})

	t.Run("Zero Amount", func(t *testing.T) {
		tx := &Transaction{
			ID:     "TX001",
			Type:   Journal,
			Status: Draft,
			Entries: []Entry{
				{
					AccountID: "ACC001",
					Amount:    money.Money{Amount: decimal.Zero, Currency: "USD"},
					Type:      Debit,
				},
				{
					AccountID: "ACC002",
					Amount:    money.Money{Amount: decimal.Zero, Currency: "USD"},
					Type:      Credit,
				},
			},
		}

		result, err := validator.Validate(ctx, tx)
		assert.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Equal(t, ErrCodeInvalidAmount, result.Errors[0].Code)
	})
}

func createValidTransaction() *Transaction {
	now := time.Now()
	return &Transaction{
		ID:          "TX001",
		Type:        Journal,
		Status:      Draft,
		Date:        now,
		Description: "Test Transaction",
		Entries: []Entry{
			{
				AccountID:   "ACC001",
				Amount:      money.Money{Amount: decimal.NewFromFloat(100), Currency: "USD"},
				Type:        Debit,
				Description: "Debit entry",
			},
			{
				AccountID:   "ACC002",
				Amount:      money.Money{Amount: decimal.NewFromFloat(100), Currency: "USD"},
				Type:        Credit,
				Description: "Credit entry",
			},
		},
		CreatedBy:    "test-user",
		Created:      now,
		LastModified: now,
	}
}
