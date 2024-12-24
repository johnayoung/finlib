package money

import (
	"testing"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestMoneyOperations(t *testing.T) {
	t.Run("Add", func(t *testing.T) {
		// Test successful addition
		m1 := Money{Amount: decimal.NewFromFloat(100.50), Currency: "USD"}
		m2 := Money{Amount: decimal.NewFromFloat(50.25), Currency: "USD"}
		result, err := m1.Add(m2)
		assert.NoError(t, err)
		assert.True(t, decimal.NewFromFloat(150.75).Equal(result.Amount))
		assert.Equal(t, "USD", result.Currency)

		// Test mismatched currencies
		m3 := Money{Amount: decimal.NewFromFloat(100.50), Currency: "EUR"}
		_, err = m1.Add(m3)
		assert.ErrorIs(t, err, ErrMismatchedCurrencies)
	})

	t.Run("Subtract", func(t *testing.T) {
		m1 := Money{Amount: decimal.NewFromFloat(100.50), Currency: "USD"}
		m2 := Money{Amount: decimal.NewFromFloat(50.25), Currency: "USD"}
		result, err := m1.Subtract(m2)
		assert.NoError(t, err)
		assert.True(t, decimal.NewFromFloat(50.25).Equal(result.Amount))
		assert.Equal(t, "USD", result.Currency)

		// Test mismatched currencies
		m3 := Money{Amount: decimal.NewFromFloat(50.25), Currency: "EUR"}
		_, err = m1.Subtract(m3)
		assert.ErrorIs(t, err, ErrMismatchedCurrencies)
	})

	t.Run("Multiply", func(t *testing.T) {
		m1 := Money{Amount: decimal.NewFromFloat(100.50), Currency: "USD"}
		result := m1.Multiply(decimal.NewFromFloat(2))
		assert.True(t, decimal.NewFromFloat(201.00).Equal(result.Amount))
		assert.Equal(t, "USD", result.Currency)
	})

	t.Run("Divide", func(t *testing.T) {
		m1 := Money{Amount: decimal.NewFromFloat(100.50), Currency: "USD"}
		
		// Test successful division
		result, err := m1.Divide(decimal.NewFromFloat(2))
		assert.NoError(t, err)
		assert.True(t, decimal.NewFromFloat(50.25).Equal(result.Amount))
		assert.Equal(t, "USD", result.Currency)

		// Test division by zero
		_, err = m1.Divide(decimal.Zero)
		assert.ErrorIs(t, err, ErrDivisionByZero)
	})

	t.Run("IsZero", func(t *testing.T) {
		m1 := Money{Amount: decimal.Zero, Currency: "USD"}
		m2 := Money{Amount: decimal.NewFromFloat(100.50), Currency: "USD"}
		assert.True(t, m1.IsZero())
		assert.False(t, m2.IsZero())
	})

	t.Run("IsNegative", func(t *testing.T) {
		m1 := Money{Amount: decimal.NewFromFloat(-100.50), Currency: "USD"}
		m2 := Money{Amount: decimal.NewFromFloat(100.50), Currency: "USD"}
		assert.True(t, m1.IsNegative())
		assert.False(t, m2.IsNegative())
	})

	t.Run("IsPositive", func(t *testing.T) {
		m1 := Money{Amount: decimal.NewFromFloat(100.50), Currency: "USD"}
		m2 := Money{Amount: decimal.NewFromFloat(-100.50), Currency: "USD"}
		assert.True(t, m1.IsPositive())
		assert.False(t, m2.IsPositive())
	})

	t.Run("Abs", func(t *testing.T) {
		m1 := Money{Amount: decimal.NewFromFloat(-100.50), Currency: "USD"}
		result := m1.Abs()
		assert.True(t, decimal.NewFromFloat(100.50).Equal(result.Amount))
		assert.Equal(t, "USD", result.Currency)
	})

	t.Run("Equal", func(t *testing.T) {
		m1 := Money{Amount: decimal.NewFromFloat(100.50), Currency: "USD"}
		m2 := Money{Amount: decimal.NewFromFloat(100.50), Currency: "USD"}
		m3 := Money{Amount: decimal.NewFromFloat(100.50), Currency: "EUR"}
		m4 := Money{Amount: decimal.NewFromFloat(200.00), Currency: "USD"}

		assert.True(t, m1.Equal(m2))
		assert.False(t, m1.Equal(m3))
		assert.False(t, m1.Equal(m4))
	})
}

func TestCurrency(t *testing.T) {
	t.Run("Currency Creation", func(t *testing.T) {
		currency := Currency{
			Code:         "USD",
			Name:         "US Dollar",
			DefaultScale: 2,
			Symbol:       "$",
			SymbolPrefix: true,
			Active:       true,
		}

		assert.Equal(t, "USD", currency.Code)
		assert.Equal(t, "US Dollar", currency.Name)
		assert.Equal(t, uint8(2), currency.DefaultScale)
		assert.Equal(t, "$", currency.Symbol)
		assert.True(t, currency.SymbolPrefix)
		assert.True(t, currency.Active)
	})
}

func TestFormat(t *testing.T) {
	t.Run("Format Creation", func(t *testing.T) {
		format := Format{
			DecimalSeparator:   ".",
			ThousandSeparator: ",",
			DecimalPlaces:     2,
		}

		assert.Equal(t, ".", format.DecimalSeparator)
		assert.Equal(t, ",", format.ThousandSeparator)
		assert.Equal(t, uint8(2), format.DecimalPlaces)
	})
}
