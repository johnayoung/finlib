package statements

import (
	"context"
	"testing"
	"time"

	"github.com/johnayoung/finlib/pkg/account"
	"github.com/johnayoung/finlib/pkg/money"
	"github.com/johnayoung/finlib/pkg/reporting"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations
type mockAccountRepository struct {
	mock.Mock
}

func (m *mockAccountRepository) Query(ctx context.Context, query interface{}, results interface{}) error {
	args := m.Called(ctx, query, results)
	if accounts, ok := args.Get(0).([]*account.Account); ok && results != nil {
		// Copy mock results to the output parameter
		*(results.(*[]*account.Account)) = accounts
	}
	if args.Get(1) != nil {
		return args.Get(1).(error)
	}
	return nil
}

func (m *mockAccountRepository) Create(ctx context.Context, entity interface{}) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *mockAccountRepository) Read(ctx context.Context, id string, entity interface{}) error {
	args := m.Called(ctx, id, entity)
	return args.Error(0)
}

func (m *mockAccountRepository) Update(ctx context.Context, entity interface{}) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *mockAccountRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type mockReportCalculator struct {
	mock.Mock
}

func (m *mockReportCalculator) CalculateBalance(ctx context.Context, accountID string, period reporting.ReportPeriod) (money.Money, error) {
	args := m.Called(ctx, accountID, period)
	if args.Get(0) == nil {
		return money.Money{}, args.Error(1)
	}
	return args.Get(0).(money.Money), args.Error(1)
}

func (m *mockReportCalculator) CalculateChanges(ctx context.Context, accountID string, period reporting.ReportPeriod) (*reporting.BalanceChange, error) {
	args := m.Called(ctx, accountID, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reporting.BalanceChange), args.Error(1)
}

func (m *mockReportCalculator) CalculateRatio(ctx context.Context, ratio reporting.RatioDefinition, period reporting.ReportPeriod) (decimal.Decimal, error) {
	args := m.Called(ctx, ratio, period)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func TestGenerateBalanceSheet(t *testing.T) {
	// Setup
	ctx := context.Background()
	calculator := new(mockReportCalculator)
	accounts := new(mockAccountRepository)
	generator := NewGenerator(calculator, accounts)

	// Mock data
	asOf := time.Date(2024, 12, 24, 0, 0, 0, 0, time.UTC)
	opts := StatementOptions{
		Currency:    "USD",
		DetailLevel: "detailed",
	}

	// Mock accounts
	mockAssets := []*account.Account{
		{ID: "1001", Name: "Cash", Type: account.Asset},
		{ID: "1002", Name: "Accounts Receivable", Type: account.Asset},
	}
	mockLiabilities := []*account.Account{
		{ID: "2001", Name: "Accounts Payable", Type: account.Liability},
	}
	mockEquity := []*account.Account{
		{ID: "3001", Name: "Retained Earnings", Type: account.Equity},
	}

	// Setup expectations
	accounts.On("Query", ctx, account.Account{Type: account.Asset}, &[]*account.Account{}).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*[]*account.Account)
			*result = mockAssets
		}).Return(mockAssets, nil)
	accounts.On("Query", ctx, account.Account{Type: account.Liability}, &[]*account.Account{}).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*[]*account.Account)
			*result = mockLiabilities
		}).Return(mockLiabilities, nil)
	accounts.On("Query", ctx, account.Account{Type: account.Equity}, &[]*account.Account{}).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*[]*account.Account)
			*result = mockEquity
		}).Return(mockEquity, nil)

	// Mock balances
	calculator.On("CalculateBalance", ctx, "1001", reporting.ReportPeriod{End: asOf}).
		Return(money.Money{Amount: decimal.NewFromInt(1000), Currency: "USD"}, nil)
	calculator.On("CalculateBalance", ctx, "1002", reporting.ReportPeriod{End: asOf}).
		Return(money.Money{Amount: decimal.NewFromInt(500), Currency: "USD"}, nil)
	calculator.On("CalculateBalance", ctx, "2001", reporting.ReportPeriod{End: asOf}).
		Return(money.Money{Amount: decimal.NewFromInt(300), Currency: "USD"}, nil)
	calculator.On("CalculateBalance", ctx, "3001", reporting.ReportPeriod{End: asOf}).
		Return(money.Money{Amount: decimal.NewFromInt(1200), Currency: "USD"}, nil)

	// Execute
	stmt, err := generator.GenerateBalanceSheet(ctx, asOf, opts)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, stmt)
	assert.Equal(t, BalanceSheet, stmt.Type)
	assert.Equal(t, "Balance Sheet", stmt.Title)
	assert.Equal(t, asOf, stmt.AsOf)
	assert.Equal(t, "USD", stmt.Currency)

	// Verify sections
	assert.Equal(t, 3, len(stmt.Sections))

	// Assets Section
	assetsSection := stmt.Sections[0]
	assert.Equal(t, "Assets", assetsSection.Title)
	assert.Equal(t, 2, len(assetsSection.Items))
	assert.Equal(t, "Cash", assetsSection.Items[0].Label)
	assert.Equal(t, decimal.NewFromInt(1000), assetsSection.Items[0].Amount.Amount)
	assert.Equal(t, "Accounts Receivable", assetsSection.Items[1].Label)
	assert.Equal(t, decimal.NewFromInt(500), assetsSection.Items[1].Amount.Amount)
	assert.Equal(t, decimal.NewFromInt(1500), assetsSection.Total.Amount) // 1000 + 500

	// Liabilities Section
	liabilitiesSection := stmt.Sections[1]
	assert.Equal(t, "Liabilities", liabilitiesSection.Title)
	assert.Equal(t, 1, len(liabilitiesSection.Items))
	assert.Equal(t, "Accounts Payable", liabilitiesSection.Items[0].Label)
	assert.Equal(t, decimal.NewFromInt(300), liabilitiesSection.Items[0].Amount.Amount)
	assert.Equal(t, decimal.NewFromInt(300), liabilitiesSection.Total.Amount)

	// Equity Section
	equitySection := stmt.Sections[2]
	assert.Equal(t, "Equity", equitySection.Title)
	assert.Equal(t, 1, len(equitySection.Items))
	assert.Equal(t, "Retained Earnings", equitySection.Items[0].Label)
	assert.Equal(t, decimal.NewFromInt(1200), equitySection.Items[0].Amount.Amount)
	assert.Equal(t, decimal.NewFromInt(1200), equitySection.Total.Amount)

	// Verify mock expectations
	accounts.AssertExpectations(t)
	calculator.AssertExpectations(t)
}

func TestGenerateIncomeStatement(t *testing.T) {
	// Setup
	ctx := context.Background()
	calculator := new(mockReportCalculator)
	accounts := new(mockAccountRepository)
	generator := NewGenerator(calculator, accounts)

	// Test data
	periodStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	period := reporting.ReportPeriod{Start: periodStart, End: periodEnd}
	opts := StatementOptions{
		Currency:    "USD",
		DetailLevel: "detailed",
	}

	// Mock accounts
	mockRevenue := []*account.Account{
		{ID: "4001", Name: "Sales Revenue", Type: account.Revenue},
		{ID: "4002", Name: "Service Revenue", Type: account.Revenue},
	}
	mockExpenses := []*account.Account{
		{ID: "5001", Name: "Cost of Goods Sold", Type: account.Expense},
		{ID: "5002", Name: "Operating Expenses", Type: account.Expense},
	}

	// Setup expectations
	accounts.On("Query", ctx, account.Account{Type: account.Revenue}, &[]*account.Account{}).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*[]*account.Account)
			*result = mockRevenue
		}).Return(mockRevenue, nil)
	accounts.On("Query", ctx, account.Account{Type: account.Expense}, &[]*account.Account{}).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*[]*account.Account)
			*result = mockExpenses
		}).Return(mockExpenses, nil)

	// Mock changes
	calculator.On("CalculateChanges", ctx, "4001", period).
		Return(&reporting.BalanceChange{NetChange: money.Money{Amount: decimal.NewFromInt(5000), Currency: "USD"}}, nil)
	calculator.On("CalculateChanges", ctx, "4002", period).
		Return(&reporting.BalanceChange{NetChange: money.Money{Amount: decimal.NewFromInt(3000), Currency: "USD"}}, nil)
	calculator.On("CalculateChanges", ctx, "5001", period).
		Return(&reporting.BalanceChange{NetChange: money.Money{Amount: decimal.NewFromInt(2000), Currency: "USD"}}, nil)
	calculator.On("CalculateChanges", ctx, "5002", period).
		Return(&reporting.BalanceChange{NetChange: money.Money{Amount: decimal.NewFromInt(1500), Currency: "USD"}}, nil)

	// Execute
	stmt, err := generator.GenerateIncomeStatement(ctx, periodStart, periodEnd, opts)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, stmt)
	assert.Equal(t, IncomeStatement, stmt.Type)
	assert.Equal(t, "Income Statement", stmt.Title)
	assert.Equal(t, periodEnd, stmt.AsOf)
	assert.Equal(t, &periodStart, stmt.PeriodStart)
	assert.Equal(t, "USD", stmt.Currency)

	// Verify sections
	assert.Len(t, stmt.Sections, 2)

	// Revenue section
	assert.Equal(t, "Revenue", stmt.Sections[0].Title)
	assert.Len(t, stmt.Sections[0].Items, 2)
	assert.Equal(t, decimal.NewFromInt(8000), stmt.Sections[0].Total.Amount)

	// Expense section
	assert.Equal(t, "Expenses", stmt.Sections[1].Title)
	assert.Len(t, stmt.Sections[1].Items, 2)
	assert.Equal(t, decimal.NewFromInt(3500), stmt.Sections[1].Total.Amount)

	// Verify mock expectations
	accounts.AssertExpectations(t)
	calculator.AssertExpectations(t)
}

func TestGenerateCashFlow(t *testing.T) {
	ctx := context.Background()
	calculator := new(mockReportCalculator)
	accounts := new(mockAccountRepository)
	g := NewGenerator(calculator, accounts)

	// Mock data
	asOf := time.Date(2024, 12, 24, 0, 0, 0, 0, time.UTC)
	periodStart := asOf.AddDate(0, -1, 0)

	// Mock accounts
	mockRevenue := []*account.Account{
		{ID: "4001", Name: "Sales Revenue", Type: account.Revenue},
	}
	mockExpenses := []*account.Account{
		{ID: "5001", Name: "Operating Expenses", Type: account.Expense},
	}
	mockAssets := []*account.Account{
		{ID: "1001", Name: "Equipment", Type: account.Asset},
	}
	mockLiabilities := []*account.Account{
		{ID: "2001", Name: "Bank Loan", Type: account.Liability},
	}

	// Mock changes
	period := reporting.ReportPeriod{Start: periodStart, End: asOf}

	// Setup expectations for net income calculation
	accounts.On("Query", ctx, account.Account{Type: account.Revenue}, &[]*account.Account{}).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*[]*account.Account)
			*result = mockRevenue
		}).Return(mockRevenue, nil)
	accounts.On("Query", ctx, account.Account{Type: account.Expense}, &[]*account.Account{}).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*[]*account.Account)
			*result = mockExpenses
		}).Return(mockExpenses, nil)

	// Setup expectations for investing activities
	accounts.On("Query", ctx, account.Account{Type: account.Asset}, &[]*account.Account{}).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*[]*account.Account)
			*result = mockAssets
		}).Return(mockAssets, nil)

	// Setup expectations for financing activities
	accounts.On("Query", ctx, account.Account{Type: account.Liability}, &[]*account.Account{}).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*[]*account.Account)
			*result = mockLiabilities
		}).Return(mockLiabilities, nil)

	// Mock changes for accounts
	calculator.On("CalculateChanges", ctx, "4001", period).
		Return(&reporting.BalanceChange{
			NetChange: money.Money{Amount: decimal.NewFromInt(100000), Currency: "USD"},
		}, nil)
	calculator.On("CalculateChanges", ctx, "5001", period).
		Return(&reporting.BalanceChange{
			NetChange: money.Money{Amount: decimal.NewFromInt(60000), Currency: "USD"},
		}, nil)
	calculator.On("CalculateChanges", ctx, "1001", period).
		Return(&reporting.BalanceChange{
			NetChange: money.Money{Amount: decimal.NewFromInt(-50000), Currency: "USD"},
		}, nil)
	calculator.On("CalculateChanges", ctx, "2001", period).
		Return(&reporting.BalanceChange{
			NetChange: money.Money{Amount: decimal.NewFromInt(30000), Currency: "USD"},
		}, nil)

	// Generate cash flow statement
	stmt, err := g.GenerateCashFlow(ctx, periodStart, asOf, StatementOptions{
		Currency:    "USD",
		DetailLevel: "detailed",
		FormatOptions: map[string]interface{}{
			"method": string(Indirect),
		},
	})

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, stmt)
	assert.Equal(t, CashFlow, stmt.Type)
	assert.Equal(t, "Statement of Cash Flows", stmt.Title)
	assert.Equal(t, asOf, stmt.AsOf)
	assert.Equal(t, &periodStart, stmt.PeriodStart)
	assert.Equal(t, "USD", stmt.Currency)

	// Verify sections
	assert.Equal(t, 3, len(stmt.Sections))

	// Operating Activities Section
	operatingSection := stmt.Sections[0]
	assert.Equal(t, "Operating Activities", operatingSection.Title)
	assert.True(t, len(operatingSection.Items) >= 1) // At least net income should be present
	assert.Equal(t, "Net Income", operatingSection.Items[0].Label)
	assert.Equal(t, decimal.NewFromInt(40000), operatingSection.Items[0].Amount.Amount) // 100000 - 60000

	// Investing Activities Section
	investingSection := stmt.Sections[1]
	assert.Equal(t, "Investing Activities", investingSection.Title)
	assert.Equal(t, 1, len(investingSection.Items))
	assert.Equal(t, "Equipment", investingSection.Items[0].Label)
	assert.Equal(t, decimal.NewFromInt(-50000), investingSection.Items[0].Amount.Amount)

	// Financing Activities Section
	financingSection := stmt.Sections[2]
	assert.Equal(t, "Financing Activities", financingSection.Title)
	assert.Equal(t, 1, len(financingSection.Items))
	assert.Equal(t, "Bank Loan", financingSection.Items[0].Label)
	assert.Equal(t, decimal.NewFromInt(30000), financingSection.Items[0].Amount.Amount)

	// Verify mock expectations
	accounts.AssertExpectations(t)
	calculator.AssertExpectations(t)
}
