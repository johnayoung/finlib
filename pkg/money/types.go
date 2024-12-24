package money

import (
	"github.com/shopspring/decimal"
)

// Money represents a monetary value in a specific currency
type Money struct {
	Amount   decimal.Decimal
	Currency string
}

// Currency represents a currency definition in the system
type Currency struct {
	// ISO 4217 code (e.g., "USD", "EUR")
	Code string
	// Official currency name
	Name string
	// Number of decimal places typically used
	DefaultScale uint8
	// Symbol used for display (e.g., "$", "€")
	Symbol string
	// Whether the symbol appears before the amount
	SymbolPrefix bool
	// Whether this currency is still active
	Active bool
}

// Format represents currency formatting options
type Format struct {
	// Decimal separator (e.g., "." or ",")
	DecimalSeparator string
	// Thousand separator (e.g., "," or ".")
	ThousandSeparator string
	// Number of decimal places to display
	DecimalPlaces uint8
}
