package money

import (
	"errors"
	"github.com/shopspring/decimal"
)

var (
	ErrMismatchedCurrencies = errors.New("mismatched currencies")
	ErrInvalidAmount        = errors.New("invalid amount")
	ErrDivisionByZero      = errors.New("division by zero")
)

// Add adds two monetary values of the same currency
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrMismatchedCurrencies
	}
	return Money{
		Amount:   m.Amount.Add(other.Amount),
		Currency: m.Currency,
	}, nil
}

// Subtract subtracts one monetary value from another of the same currency
func (m Money) Subtract(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrMismatchedCurrencies
	}
	return Money{
		Amount:   m.Amount.Sub(other.Amount),
		Currency: m.Currency,
	}, nil
}

// Multiply multiplies a monetary value by a decimal factor
func (m Money) Multiply(factor decimal.Decimal) Money {
	return Money{
		Amount:   m.Amount.Mul(factor),
		Currency: m.Currency,
	}
}

// Divide divides a monetary value by a decimal factor
func (m Money) Divide(factor decimal.Decimal) (Money, error) {
	if factor.IsZero() {
		return Money{}, ErrDivisionByZero
	}
	return Money{
		Amount:   m.Amount.Div(factor),
		Currency: m.Currency,
	}, nil
}

// IsZero returns true if the monetary amount is zero
func (m Money) IsZero() bool {
	return m.Amount.IsZero()
}

// IsNegative returns true if the monetary amount is negative
func (m Money) IsNegative() bool {
	return m.Amount.IsNegative()
}

// IsPositive returns true if the monetary amount is positive
func (m Money) IsPositive() bool {
	return m.Amount.IsPositive()
}

// Abs returns the absolute value of the monetary amount
func (m Money) Abs() Money {
	return Money{
		Amount:   m.Amount.Abs(),
		Currency: m.Currency,
	}
}

// Equal returns true if two monetary values are equal
func (m Money) Equal(other Money) bool {
	return m.Currency == other.Currency && m.Amount.Equal(other.Amount)
}
