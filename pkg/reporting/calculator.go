package reporting

import (
	"context"
	"fmt"
	"time"

	"github.com/johnayoung/finlib/pkg/account"
	"github.com/johnayoung/finlib/pkg/money"
	"github.com/johnayoung/finlib/pkg/storage"
	"github.com/johnayoung/finlib/pkg/transaction"
	"github.com/shopspring/decimal"
)

// defaultReportCalculator implements the ReportCalculator interface
type defaultReportCalculator struct {
	accountStore     account.Repository
	transactionProc  transaction.TransactionProcessor
	transactionStore storage.Repository
}

// NewReportCalculator creates a new instance of the report calculator
func NewReportCalculator(
	accountStore account.Repository,
	transactionProc transaction.TransactionProcessor,
	transactionStore storage.Repository,
) ReportCalculator {
	return &defaultReportCalculator{
		accountStore:     accountStore,
		transactionProc:  transactionProc,
		transactionStore: transactionStore,
	}
}

// CalculateBalance computes account balances for reporting
func (c *defaultReportCalculator) CalculateBalance(ctx context.Context, accountID string, period ReportPeriod) (money.Money, error) {
	// Get the account
	var acc account.Account
	if err := c.accountStore.Read(ctx, accountID, &acc); err != nil {
		return money.Money{}, fmt.Errorf("error reading account: %w", err)
	}

	// Get transactions for the period
	transactions, err := c.getTransactionsForPeriod(ctx, accountID, period)
	if err != nil {
		return money.Money{}, fmt.Errorf("error getting transactions: %w", err)
	}

	// Calculate balance from transactions
	return c.calculateBalanceFromTransactions(transactions, acc.Type)
}

// CalculateChanges computes changes over a period
func (c *defaultReportCalculator) CalculateChanges(ctx context.Context, accountID string, period ReportPeriod) (*BalanceChange, error) {
	// Get transactions for the period
	transactions, err := c.getTransactionsForPeriod(ctx, accountID, period)
	if err != nil {
		return nil, fmt.Errorf("error getting transactions: %w", err)
	}

	// Calculate opening balance (balance at start of period)
	openingBalance, err := c.getBalanceAtTime(ctx, accountID, period.Start)
	if err != nil {
		return nil, fmt.Errorf("error calculating opening balance: %w", err)
	}

	// Calculate closing balance (balance at end of period)
	closingBalance, err := c.getBalanceAtTime(ctx, accountID, period.End)
	if err != nil {
		return nil, fmt.Errorf("error calculating closing balance: %w", err)
	}

	// Get movements from transactions
	movements := make([]BalanceMovement, 0, len(transactions))
	for _, tx := range transactions {
		for _, entry := range tx.Entries {
			if entry.AccountID == accountID {
				movements = append(movements, BalanceMovement{
					Date:        tx.Date,
					Amount:      entry.Amount,
					Type:        string(entry.Type),
					Description: tx.Description,
					Reference:   tx.ID,
				})
			}
		}
	}

	// Calculate net change
	netChange := money.Money{
		Amount:   closingBalance.Amount.Sub(openingBalance.Amount),
		Currency: closingBalance.Currency,
	}

	return &BalanceChange{
		OpeningBalance: openingBalance,
		ClosingBalance: closingBalance,
		NetChange:      netChange,
		Movements:      movements,
	}, nil
}

// CalculateRatio computes financial ratios
func (c *defaultReportCalculator) CalculateRatio(ctx context.Context, ratio RatioDefinition, period ReportPeriod) (decimal.Decimal, error) {
	// Calculate numerator
	numerator, err := c.calculateValue(ctx, ratio.Numerator, period)
	if err != nil {
		return decimal.Zero, fmt.Errorf("error calculating numerator: %w", err)
	}

	// Calculate denominator
	denominator, err := c.calculateValue(ctx, ratio.Denominator, period)
	if err != nil {
		return decimal.Zero, fmt.Errorf("error calculating denominator: %w", err)
	}

	// Check for division by zero
	if denominator.IsZero() {
		return decimal.Zero, fmt.Errorf("division by zero: denominator is zero")
	}

	// Calculate ratio with proper scaling
	result := numerator.Div(denominator)
	if ratio.Scale > 0 {
		result = result.Round(ratio.Scale)
	}

	return result, nil
}

// Helper functions

func (c *defaultReportCalculator) getTransactionsForPeriod(ctx context.Context, accountID string, period ReportPeriod) ([]*transaction.Transaction, error) {
	// Create a query to find transactions for the account within the period
	query := storage.Query{
		Filters: []storage.Filter{
			{Field: "entries.account_id", Operator: "=", Value: accountID},
			{Field: "date", Operator: ">=", Value: period.Start},
			{Field: "date", Operator: "<=", Value: period.End},
			{Field: "status", Operator: "=", Value: transaction.Posted},
		},
		Sort: []storage.Sort{
			{Field: "date", Desc: false},
		},
	}

	var transactions []*transaction.Transaction
	if err := c.transactionStore.Query(ctx, query, &transactions); err != nil {
		return nil, fmt.Errorf("error querying transactions: %w", err)
	}

	return transactions, nil
}

func (c *defaultReportCalculator) calculateBalanceFromTransactions(transactions []*transaction.Transaction, accountType account.AccountType) (money.Money, error) {
	if len(transactions) == 0 {
		return money.Money{Amount: decimal.Zero, Currency: "USD"}, nil
	}

	// Use the currency of the first transaction as the balance currency
	currency := transactions[0].Entries[0].Amount.Currency
	balance := decimal.Zero

	for _, tx := range transactions {
		for _, entry := range tx.Entries {
			if entry.Amount.Currency != currency {
				return money.Money{}, fmt.Errorf("mixed currencies in transactions")
			}

			switch entry.Type {
			case transaction.Debit:
				if accountType == account.Asset || accountType == account.Expense {
					balance = balance.Add(entry.Amount.Amount)
				} else {
					balance = balance.Sub(entry.Amount.Amount)
				}
			case transaction.Credit:
				if accountType == account.Asset || accountType == account.Expense {
					balance = balance.Sub(entry.Amount.Amount)
				} else {
					balance = balance.Add(entry.Amount.Amount)
				}
			}
		}
	}

	return money.Money{Amount: balance, Currency: currency}, nil
}

func (c *defaultReportCalculator) getBalanceAtTime(ctx context.Context, accountID string, at time.Time) (money.Money, error) {
	// Get all transactions up to the specified time
	period := ReportPeriod{
		Start: time.Time{}, // Beginning of time
		End:   at,
	}

	transactions, err := c.getTransactionsForPeriod(ctx, accountID, period)
	if err != nil {
		return money.Money{}, fmt.Errorf("error getting transactions: %w", err)
	}

	// Get the account to determine its type
	var acc account.Account
	if err := c.accountStore.Read(ctx, accountID, &acc); err != nil {
		return money.Money{}, fmt.Errorf("error reading account: %w", err)
	}

	return c.calculateBalanceFromTransactions(transactions, acc.Type)
}

func (c *defaultReportCalculator) calculateValue(ctx context.Context, calc Calculation, period ReportPeriod) (decimal.Decimal, error) {
	// Get accounts matching the selector
	accounts, err := c.getAccountsForSelector(ctx, calc.AccountSelector)
	if err != nil {
		return decimal.Zero, fmt.Errorf("error getting accounts: %w", err)
	}

	// Calculate total for all matching accounts
	total := decimal.Zero
	for _, acc := range accounts {
		balance, err := c.CalculateBalance(ctx, acc.ID, period)
		if err != nil {
			return decimal.Zero, fmt.Errorf("error calculating balance for account %s: %w", acc.ID, err)
		}
		total = total.Add(balance.Amount)
	}

	return total, nil
}

func (c *defaultReportCalculator) getAccountsForSelector(ctx context.Context, selector AccountSelector) ([]*account.Account, error) {
	// Create a query based on the selector criteria
	query := storage.Query{
		Filters: make([]storage.Filter, 0),
	}

	// Add type filters
	if len(selector.Types) > 0 {
		query.Filters = append(query.Filters, storage.Filter{
			Field:    "type",
			Operator: "in",
			Value:    selector.Types,
		})
	}

	// Add code filters
	if len(selector.Codes) > 0 {
		query.Filters = append(query.Filters, storage.Filter{
			Field:    "code",
			Operator: "in",
			Value:    selector.Codes,
		})
	}

	// Add category filters
	if len(selector.Categories) > 0 {
		query.Filters = append(query.Filters, storage.Filter{
			Field:    "category",
			Operator: "in",
			Value:    selector.Categories,
		})
	}

	var accounts []*account.Account
	if err := c.accountStore.Query(ctx, query, &accounts); err != nil {
		return nil, fmt.Errorf("error querying accounts: %w", err)
	}

	return accounts, nil
}
