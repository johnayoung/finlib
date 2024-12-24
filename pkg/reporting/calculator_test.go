package reporting

import (
	"context"
	"testing"
	"time"

	"github.com/johnayoung/finlib/pkg/account"
	"github.com/johnayoung/finlib/pkg/money"
	"github.com/johnayoung/finlib/pkg/storage"
	"github.com/johnayoung/finlib/pkg/transaction"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations
type mockAccountRepository struct {
	mock.Mock
}

func (m *mockAccountRepository) Create(ctx context.Context, entity interface{}) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *mockAccountRepository) Read(ctx context.Context, id string, entity interface{}) error {
	args := m.Called(ctx, id, entity)
	if acc, ok := args.Get(0).(*account.Account); ok && acc != nil {
		*(entity.(*account.Account)) = *acc
	}
	return args.Error(1)
}

func (m *mockAccountRepository) Update(ctx context.Context, entity interface{}) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *mockAccountRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockAccountRepository) Query(ctx context.Context, query interface{}, results interface{}) error {
	args := m.Called(ctx, query, results)
	return args.Error(0)
}

func (m *mockAccountRepository) Count(ctx context.Context, query interface{}) (int64, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(int64), args.Error(1)
}

type mockTransactionProcessor struct {
	mock.Mock
}

func (m *mockTransactionProcessor) ValidateTransaction(ctx context.Context, tx *transaction.Transaction) (*transaction.ValidationResult, error) {
	args := m.Called(ctx, tx)
	return args.Get(0).(*transaction.ValidationResult), args.Error(1)
}

func (m *mockTransactionProcessor) ProcessTransaction(ctx context.Context, tx *transaction.Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func (m *mockTransactionProcessor) ProcessTransactionBatch(ctx context.Context, txs []*transaction.Transaction) error {
	args := m.Called(ctx, txs)
	return args.Error(0)
}

func (m *mockTransactionProcessor) VoidTransaction(ctx context.Context, txID string, reason string) error {
	args := m.Called(ctx, txID, reason)
	return args.Error(0)
}

func (m *mockTransactionProcessor) ReverseTransaction(ctx context.Context, txID string, reason string) error {
	args := m.Called(ctx, txID, reason)
	return args.Error(0)
}

func (m *mockTransactionProcessor) GetTransaction(ctx context.Context, txID string) (*transaction.Transaction, error) {
	args := m.Called(ctx, txID)
	return args.Get(0).(*transaction.Transaction), args.Error(1)
}

func (m *mockTransactionProcessor) GetTransactionSummary(ctx context.Context, tx *transaction.Transaction) (*transaction.TransactionSummary, error) {
	args := m.Called(ctx, tx)
	return args.Get(0).(*transaction.TransactionSummary), args.Error(1)
}

type mockTransactionRepository struct {
	mock.Mock
}

func (m *mockTransactionRepository) Create(ctx context.Context, entity interface{}) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *mockTransactionRepository) Read(ctx context.Context, id string, entity interface{}) error {
	args := m.Called(ctx, id, entity)
	return args.Error(0)
}

func (m *mockTransactionRepository) Update(ctx context.Context, entity interface{}) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *mockTransactionRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockTransactionRepository) Query(ctx context.Context, query storage.Query, results interface{}) error {
	args := m.Called(ctx, query, results)
	if txs, ok := args.Get(0).([]*transaction.Transaction); ok && txs != nil {
		*(results.(*[]*transaction.Transaction)) = txs
	}
	return args.Error(0)
}

func (m *mockTransactionRepository) Count(ctx context.Context, query storage.Query) (int64, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(int64), args.Error(1)
}

func TestNewReportCalculator(t *testing.T) {
	accountStore := &mockAccountRepository{}
	transactionProc := &mockTransactionProcessor{}
	transactionStore := &mockTransactionRepository{}
	
	calculator := NewReportCalculator(accountStore, transactionProc, transactionStore)
	assert.NotNil(t, calculator)
}

func TestCalculateBalance(t *testing.T) {
	// Setup
	ctx := context.Background()
	accountStore := &mockAccountRepository{}
	transactionProc := &mockTransactionProcessor{}
	transactionStore := &mockTransactionRepository{}
	calculator := NewReportCalculator(accountStore, transactionProc, transactionStore)

	testTime := time.Date(2024, 12, 24, 10, 0, 0, 0, time.UTC)
	testAccount := &account.Account{
		ID:      "ACC001",
		Type:    account.Asset,
		Balance: &money.Money{Amount: decimal.NewFromInt(1000), Currency: "USD"},
	}

	period := ReportPeriod{
		Start: testTime.AddDate(0, -1, 0),
		End:   testTime,
	}

	// Create test transactions
	transactions := []*transaction.Transaction{
		{
			ID:     "TXN001",
			Type:   transaction.Journal,
			Status: transaction.Posted,
			Date:   testTime.AddDate(0, -1, 1),
			Entries: []transaction.Entry{
				{
					AccountID: "ACC001",
					Amount:    money.Money{Amount: decimal.NewFromInt(500), Currency: "USD"},
					Type:     transaction.Debit,
				},
			},
		},
		{
			ID:     "TXN002",
			Type:   transaction.Journal,
			Status: transaction.Posted,
			Date:   testTime.AddDate(0, 0, -1),
			Entries: []transaction.Entry{
				{
					AccountID: "ACC001",
					Amount:    money.Money{Amount: decimal.NewFromInt(300), Currency: "USD"},
					Type:     transaction.Credit,
				},
			},
		},
	}

	// Setup mock expectations
	accountStore.On("Read", mock.Anything, "ACC001", mock.Anything).
		Run(func(args mock.Arguments) {
			acc := args.Get(2).(*account.Account)
			*acc = *testAccount
		}).
		Return(testAccount, nil)

	transactionStore.On("Query", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*[]*transaction.Transaction)
			*result = transactions
		}).
		Return(nil)

	// Execute test
	balance, err := calculator.CalculateBalance(ctx, "ACC001", period)

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, "USD", balance.Currency)
	assert.Equal(t, decimal.NewFromInt(200), balance.Amount) // 500 (debit) - 300 (credit) = 200
}

func TestCalculateChanges(t *testing.T) {
	// Setup
	ctx := context.Background()
	accountStore := &mockAccountRepository{}
	transactionProc := &mockTransactionProcessor{}
	transactionStore := &mockTransactionRepository{}
	calculator := NewReportCalculator(accountStore, transactionProc, transactionStore)

	testTime := time.Date(2024, 12, 24, 10, 0, 0, 0, time.UTC)
	testAccount := &account.Account{
		ID:      "ACC001",
		Type:    account.Asset,
		Balance: &money.Money{Amount: decimal.NewFromInt(1000), Currency: "USD"},
	}

	period := ReportPeriod{
		Start: testTime.AddDate(0, -1, 0),
		End:   testTime,
	}

	// Create test transactions
	transactions := []*transaction.Transaction{
		{
			ID:     "TXN001",
			Type:   transaction.Journal,
			Status: transaction.Posted,
			Date:   testTime.AddDate(0, -1, 1),
			Entries: []transaction.Entry{
				{
					AccountID: "ACC001",
					Amount:    money.Money{Amount: decimal.NewFromInt(500), Currency: "USD"},
					Type:     transaction.Debit,
				},
			},
		},
	}

	// Setup mock expectations
	accountStore.On("Read", mock.Anything, "ACC001", mock.Anything).
		Run(func(args mock.Arguments) {
			acc := args.Get(2).(*account.Account)
			*acc = *testAccount
		}).
		Return(testAccount, nil)

	transactionStore.On("Query", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*[]*transaction.Transaction)
			*result = transactions
		}).
		Return(nil)

	// Execute test
	changes, err := calculator.CalculateChanges(ctx, "ACC001", period)

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, changes)
	assert.Equal(t, 1, len(changes.Movements))
	assert.Equal(t, decimal.NewFromInt(500), changes.Movements[0].Amount.Amount)
}

func TestCalculateRatio(t *testing.T) {
	// Setup
	ctx := context.Background()
	accountStore := &mockAccountRepository{}
	transactionProc := &mockTransactionProcessor{}
	transactionStore := &mockTransactionRepository{}
	calculator := NewReportCalculator(accountStore, transactionProc, transactionStore)

	testTime := time.Date(2024, 12, 24, 10, 0, 0, 0, time.UTC)
	period := ReportPeriod{
		Start: testTime.AddDate(0, -1, 0),
		End:   testTime,
	}

	// Create test accounts
	assetAccount := &account.Account{
		ID:      "ASSET001",
		Type:    account.Asset,
		Balance: &money.Money{Amount: decimal.NewFromInt(1000), Currency: "USD"},
	}

	liabilityAccount := &account.Account{
		ID:      "LIAB001",
		Type:    account.Liability,
		Balance: &money.Money{Amount: decimal.NewFromInt(500), Currency: "USD"},
	}

	// Create test transactions for asset account
	assetTransactions := []*transaction.Transaction{
		{
			ID:     "TXN001",
			Type:   transaction.Journal,
			Status: transaction.Posted,
			Date:   testTime.AddDate(0, -1, 1),
			Entries: []transaction.Entry{
				{
					AccountID: "ASSET001",
					Amount:    money.Money{Amount: decimal.NewFromInt(1000), Currency: "USD"},
					Type:     transaction.Debit,
				},
			},
		},
	}

	// Create test transactions for liability account
	liabilityTransactions := []*transaction.Transaction{
		{
			ID:     "TXN002",
			Type:   transaction.Journal,
			Status: transaction.Posted,
			Date:   testTime.AddDate(0, -1, 1),
			Entries: []transaction.Entry{
				{
					AccountID: "LIAB001",
					Amount:    money.Money{Amount: decimal.NewFromInt(500), Currency: "USD"},
					Type:     transaction.Credit,
				},
			},
		},
	}

	// Create test ratio
	ratio := RatioDefinition{
		ID:          "CURRENT_RATIO",
		Name:        "Current Ratio",
		Description: "Current Assets / Current Liabilities",
		Scale:       2,
		Numerator: Calculation{
			ID:   "CURRENT_ASSETS",
			Type: "BALANCE",
			AccountSelector: AccountSelector{
				Types: []account.AccountType{account.Asset},
			},
		},
		Denominator: Calculation{
			ID:   "CURRENT_LIABILITIES",
			Type: "BALANCE",
			AccountSelector: AccountSelector{
				Types: []account.AccountType{account.Liability},
			},
		},
	}

	// Setup mock expectations
	accountStore.On("Query", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*[]*account.Account)
			query := args.Get(1).(storage.Query)
			if query.Filters[0].Value.([]account.AccountType)[0] == account.Asset {
				*result = []*account.Account{assetAccount}
			} else {
				*result = []*account.Account{liabilityAccount}
			}
		}).
		Return(nil)

	accountStore.On("Read", mock.Anything, "ASSET001", mock.Anything).
		Run(func(args mock.Arguments) {
			acc := args.Get(2).(*account.Account)
			*acc = *assetAccount
		}).
		Return(assetAccount, nil)

	accountStore.On("Read", mock.Anything, "LIAB001", mock.Anything).
		Run(func(args mock.Arguments) {
			acc := args.Get(2).(*account.Account)
			*acc = *liabilityAccount
		}).
		Return(liabilityAccount, nil)

	transactionStore.On("Query", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			query := args.Get(1).(storage.Query)
			result := args.Get(2).(*[]*transaction.Transaction)
			
			// Check which account we're querying for
			for _, filter := range query.Filters {
				if filter.Field == "entries.account_id" {
					accountID := filter.Value.(string)
					if accountID == "ASSET001" {
						*result = assetTransactions
					} else if accountID == "LIAB001" {
						*result = liabilityTransactions
					}
					break
				}
			}
		}).
		Return(nil)

	// Execute test
	result, err := calculator.CalculateRatio(ctx, ratio, period)

	// Verify results
	assert.NoError(t, err)
	assert.True(t, decimal.NewFromInt(2).Equal(result)) // 1000/500 = 2.00
}
