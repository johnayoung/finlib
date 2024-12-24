package transaction

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/johnayoung/finlib/pkg/money"
	"github.com/johnayoung/finlib/pkg/storage"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of storage.Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, entity interface{}) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockRepository) Read(ctx context.Context, id string, entity interface{}) error {
	args := m.Called(ctx, id, entity)
	return args.Error(0)
}

func (m *MockRepository) Update(ctx context.Context, entity interface{}) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) Query(ctx context.Context, query storage.Query, results interface{}) error {
	args := m.Called(ctx, query, results)
	return args.Error(0)
}

func (m *MockRepository) Count(ctx context.Context, query storage.Query) (int64, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(int64), args.Error(1)
}

// NewTestTransaction creates a valid transaction for testing
func NewTestTransaction() *Transaction {
	return &Transaction{
		ID:          "TX001",
		Type:        Journal,
		Status:      Draft,
		Date:        time.Now(),
		Description: "Test Transaction",
		Entries: []Entry{
			{
				AccountID:    "ACC001",
				Amount:      money.Money{Amount: decimal.NewFromInt(100), Currency: "USD"},
				Type:        Debit,
				Description: "Debit Entry",
			},
			{
				AccountID:    "ACC002",
				Amount:      money.Money{Amount: decimal.NewFromInt(100), Currency: "USD"},
				Type:        Credit,
				Description: "Credit Entry",
			},
		},
		CreatedBy: "test-user",
		Created:   time.Now(),
	}
}

func TestBasicValidator_Validate(t *testing.T) {
	tests := []struct {
		name          string
		transaction   *Transaction
		wantValid     bool
		wantErrorCode string
	}{
		{
			name:        "valid transaction",
			transaction: NewTestTransaction(),
			wantValid:   true,
		},
		{
			name: "insufficient entries",
			transaction: &Transaction{
				ID:     "TX002",
				Status: Draft,
				Entries: []Entry{
					{
						AccountID: "ACC001",
						Amount:    money.Money{Amount: decimal.NewFromInt(100), Currency: "USD"},
						Type:      Debit,
					},
				},
			},
			wantValid:     false,
			wantErrorCode: ErrCodeInsufficientEntries,
		},
		{
			name: "unbalanced transaction",
			transaction: &Transaction{
				ID:     "TX003",
				Status: Draft,
				Entries: []Entry{
					{
						AccountID: "ACC001",
						Amount:    money.Money{Amount: decimal.NewFromInt(100), Currency: "USD"},
						Type:      Debit,
					},
					{
						AccountID: "ACC002",
						Amount:    money.Money{Amount: decimal.NewFromInt(50), Currency: "USD"},
						Type:      Credit,
					},
				},
			},
			wantValid:     false,
			wantErrorCode: ErrCodeUnbalanced,
		},
		{
			name: "mixed currencies",
			transaction: &Transaction{
				ID:     "TX004",
				Status: Draft,
				Entries: []Entry{
					{
						AccountID: "ACC001",
						Amount:    money.Money{Amount: decimal.NewFromInt(100), Currency: "USD"},
						Type:      Debit,
					},
					{
						AccountID: "ACC002",
						Amount:    money.Money{Amount: decimal.NewFromInt(100), Currency: "EUR"},
						Type:      Credit,
					},
				},
			},
			wantValid:     false,
			wantErrorCode: ErrCodeMixedCurrencies,
		},
		{
			name: "duplicate account",
			transaction: &Transaction{
				ID:     "TX005",
				Status: Draft,
				Entries: []Entry{
					{
						AccountID: "ACC001",
						Amount:    money.Money{Amount: decimal.NewFromInt(100), Currency: "USD"},
						Type:      Debit,
					},
					{
						AccountID: "ACC001",
						Amount:    money.Money{Amount: decimal.NewFromInt(100), Currency: "USD"},
						Type:      Credit,
					},
				},
			},
			wantValid:     false,
			wantErrorCode: ErrCodeDuplicateAccount,
		},
	}

	validator := &BasicValidator{}
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.Validate(ctx, tt.transaction)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantValid, result.Valid)

			if !tt.wantValid {
				assert.Greater(t, len(result.Errors), 0)
				assert.Equal(t, tt.wantErrorCode, result.Errors[0].Code)
			}
		})
	}
}

func TestBasicTransactionProcessor_ProcessTransaction(t *testing.T) {
	tests := []struct {
		name        string
		transaction *Transaction
		setupMock   func(*MockRepository)
		wantErr     bool
		errCheck    func(error) bool
	}{
		{
			name:        "successful processing",
			transaction: NewTestTransaction(),
			setupMock: func(repo *MockRepository) {
				repo.On("Update", mock.Anything, mock.AnythingOfType("*transaction.Transaction")).
					Run(func(args mock.Arguments) {
						tx := args.Get(1).(*Transaction)
						assert.Equal(t, Posted, tx.Status)
						assert.NotNil(t, tx.PostedAt)
					}).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "invalid status",
			transaction: func() *Transaction {
				tx := NewTestTransaction()
				tx.Status = Posted
				return tx
			}(),
			setupMock: func(repo *MockRepository) {},
			wantErr:   true,
			errCheck: func(err error) bool {
				return err.Error() == "transaction must be in Draft or Pending status to process"
			},
		},
		{
			name:        "storage error",
			transaction: NewTestTransaction(),
			setupMock: func(repo *MockRepository) {
				repo.On("Update", mock.Anything, mock.AnythingOfType("*transaction.Transaction")).
					Return(assert.AnError)
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return err.Error() == "failed to store transaction: assert.AnError general error for testing"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			tt.setupMock(mockRepo)

			processor := NewBasicTransactionProcessor(mockRepo)
			err := processor.ProcessTransaction(context.Background(), tt.transaction)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCheck != nil {
					assert.True(t, tt.errCheck(err))
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestBasicTransactionProcessor_GetTransactionSummary(t *testing.T) {
	tests := []struct {
		name        string
		transaction *Transaction
		want        *TransactionSummary
	}{
		{
			name:        "basic summary",
			transaction: NewTestTransaction(),
			want: &TransactionSummary{
				TotalDebits:      money.Money{Amount: decimal.NewFromInt(100), Currency: "USD"},
				TotalCredits:     money.Money{Amount: decimal.NewFromInt(100), Currency: "USD"},
				NetAmount:        money.Money{Amount: decimal.Zero, Currency: "USD"},
				EntryCount:       2,
				AffectedAccounts: []string{"ACC001", "ACC002"},
			},
		},
		{
			name: "multiple entries",
			transaction: &Transaction{
				Entries: []Entry{
					{
						AccountID: "ACC001",
						Amount:    money.Money{Amount: decimal.NewFromInt(100), Currency: "USD"},
						Type:      Debit,
					},
					{
						AccountID: "ACC002",
						Amount:    money.Money{Amount: decimal.NewFromInt(50), Currency: "USD"},
						Type:      Credit,
					},
					{
						AccountID: "ACC003",
						Amount:    money.Money{Amount: decimal.NewFromInt(50), Currency: "USD"},
						Type:      Credit,
					},
				},
			},
			want: &TransactionSummary{
				TotalDebits:      money.Money{Amount: decimal.NewFromInt(100), Currency: "USD"},
				TotalCredits:     money.Money{Amount: decimal.NewFromInt(100), Currency: "USD"},
				NetAmount:        money.Money{Amount: decimal.Zero, Currency: "USD"},
				EntryCount:       3,
				AffectedAccounts: []string{"ACC001", "ACC002", "ACC003"},
			},
		},
	}

	processor := NewBasicTransactionProcessor(&MockRepository{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processor.GetTransactionSummary(context.Background(), tt.transaction)
			assert.NoError(t, err)
			assert.Equal(t, tt.want.TotalDebits.Amount.String(), got.TotalDebits.Amount.String())
			assert.Equal(t, tt.want.TotalCredits.Amount.String(), got.TotalCredits.Amount.String())
			assert.Equal(t, tt.want.NetAmount.Amount.String(), got.NetAmount.Amount.String())
			assert.Equal(t, tt.want.EntryCount, got.EntryCount)
			assert.ElementsMatch(t, tt.want.AffectedAccounts, got.AffectedAccounts)
		})
	}
}

func TestBasicTransactionProcessor_ProcessTransactionBatch(t *testing.T) {
	tests := []struct {
		name      string
		txs       []*Transaction
		setupMock func(*MockRepository)
		wantErr   bool
		errCheck  func(error) bool
	}{
		{
			name: "successful batch processing",
			txs: []*Transaction{
				NewTestTransaction(),
				func() *Transaction {
					tx := NewTestTransaction()
					tx.ID = "TX002"
					return tx
				}(),
			},
			setupMock: func(repo *MockRepository) {
				repo.On("Update", mock.Anything, mock.MatchedBy(func(tx *Transaction) bool {
					return tx.ID == "TX001"
				})).Return(nil)
				repo.On("Update", mock.Anything, mock.MatchedBy(func(tx *Transaction) bool {
					return tx.ID == "TX002"
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "empty batch",
			txs:  []*Transaction{},
			setupMock: func(repo *MockRepository) {
				// No calls expected
			},
			wantErr: false,
		},
		{
			name: "invalid transaction in batch",
			txs: []*Transaction{
				NewTestTransaction(),
				func() *Transaction {
					tx := NewTestTransaction()
					tx.ID = "TX002"
					tx.Status = Posted // Invalid status
					return tx
				}(),
			},
			setupMock: func(repo *MockRepository) {
				// No updates should be called due to validation failure
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return err.Error() == "transaction TX002 must be in Draft or Pending status to process"
			},
		},
		{
			name: "storage error with rollback",
			txs: []*Transaction{
				NewTestTransaction(),
				func() *Transaction {
					tx := NewTestTransaction()
					tx.ID = "TX002"
					return tx
				}(),
			},
			setupMock: func(repo *MockRepository) {
				// First transaction succeeds
				repo.On("Update", mock.Anything, mock.MatchedBy(func(tx *Transaction) bool {
					return tx.ID == "TX001"
				})).Return(nil)
				
				// Second transaction fails
				repo.On("Update", mock.Anything, mock.MatchedBy(func(tx *Transaction) bool {
					return tx.ID == "TX002"
				})).Return(assert.AnError)

				// Rollback for first transaction
				repo.On("Update", mock.Anything, mock.MatchedBy(func(tx *Transaction) bool {
					return tx.ID == "TX001" && tx.Status == Draft
				})).Return(nil)
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return err.Error() == "failed to store transaction TX002: assert.AnError general error for testing"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			tt.setupMock(mockRepo)

			processor := NewBasicTransactionProcessor(mockRepo)
			err := processor.ProcessTransactionBatch(context.Background(), tt.txs)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCheck != nil {
					assert.True(t, tt.errCheck(err))
				}
			} else {
				assert.NoError(t, err)
				if len(tt.txs) > 0 {
					// Verify all transactions are posted
					for _, tx := range tt.txs {
						assert.Equal(t, Posted, tx.Status)
						assert.NotNil(t, tx.PostedAt)
					}
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestBasicTransactionProcessor_VoidTransaction(t *testing.T) {
	tests := []struct {
		name      string
		txID      string
		reason    string
		setupMock func(*MockRepository)
		wantErr   bool
		errCheck  func(error) bool
	}{
		{
			name:   "successful void",
			txID:   "TX001",
			reason: "Test void",
			setupMock: func(repo *MockRepository) {
				// Setup for GetTransaction
				repo.On("Read", mock.Anything, "TX001", mock.AnythingOfType("*transaction.Transaction")).
					Run(func(args mock.Arguments) {
						tx := args.Get(2).(*Transaction)
						*tx = *NewTestTransaction()
						tx.Status = Posted
					}).
					Return(nil)

				// Setup for Update
				repo.On("Update", mock.Anything, mock.MatchedBy(func(tx *Transaction) bool {
					return tx.ID == "TX001" && tx.Status == Voided && tx.VoidReason == "Test void"
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "transaction not found",
			txID:   "TX999",
			reason: "Test void",
			setupMock: func(repo *MockRepository) {
				repo.On("Read", mock.Anything, "TX999", mock.AnythingOfType("*transaction.Transaction")).
					Return(fmt.Errorf("transaction not found"))
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return err.Error() == "failed to retrieve transaction: failed to retrieve transaction: transaction not found"
			},
		},
		{
			name:   "already voided",
			txID:   "TX001",
			reason: "Test void",
			setupMock: func(repo *MockRepository) {
				repo.On("Read", mock.Anything, "TX001", mock.AnythingOfType("*transaction.Transaction")).
					Run(func(args mock.Arguments) {
						tx := args.Get(2).(*Transaction)
						*tx = *NewTestTransaction()
						tx.Status = Posted
						now := time.Now()
						tx.VoidedAt = &now
					}).
					Return(nil)
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return err.Error() == "transaction is already voided"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			tt.setupMock(mockRepo)

			processor := NewBasicTransactionProcessor(mockRepo)
			err := processor.VoidTransaction(context.Background(), tt.txID, tt.reason)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCheck != nil {
					assert.True(t, tt.errCheck(err))
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestBasicTransactionProcessor_ReverseTransaction(t *testing.T) {
	tests := []struct {
		name      string
		txID      string
		reason    string
		setupMock func(*MockRepository)
		wantErr   bool
		errCheck  func(error) bool
	}{
		{
			name:   "successful reversal",
			txID:   "TX001",
			reason: "Test reversal",
			setupMock: func(repo *MockRepository) {
				// Setup for GetTransaction
				repo.On("Read", mock.Anything, "TX001", mock.AnythingOfType("*transaction.Transaction")).
					Run(func(args mock.Arguments) {
						tx := args.Get(2).(*Transaction)
						*tx = *NewTestTransaction()
						tx.Status = Posted
					}).
					Return(nil)

				// Setup for ProcessTransaction (creating reversal)
				repo.On("Update", mock.Anything, mock.MatchedBy(func(tx *Transaction) bool {
					return tx.Type == Reversal && tx.ReversedFrom == "TX001"
				})).Return(nil)

				// Setup for updating original transaction
				repo.On("Update", mock.Anything, mock.MatchedBy(func(tx *Transaction) bool {
					return tx.ID == "TX001" && tx.ReversalID != ""
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "transaction not found",
			txID:   "TX999",
			reason: "Test reversal",
			setupMock: func(repo *MockRepository) {
				repo.On("Read", mock.Anything, "TX999", mock.AnythingOfType("*transaction.Transaction")).
					Return(fmt.Errorf("transaction not found"))
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return err.Error() == "failed to retrieve transaction: failed to retrieve transaction: transaction not found"
			},
		},
		{
			name:   "already reversed",
			txID:   "TX001",
			reason: "Test reversal",
			setupMock: func(repo *MockRepository) {
				repo.On("Read", mock.Anything, "TX001", mock.AnythingOfType("*transaction.Transaction")).
					Run(func(args mock.Arguments) {
						tx := args.Get(2).(*Transaction)
						*tx = *NewTestTransaction()
						tx.Status = Posted
						now := time.Now()
						tx.ReversedAt = &now
					}).
					Return(nil)
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return err.Error() == "transaction is already reversed"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			tt.setupMock(mockRepo)

			processor := NewBasicTransactionProcessor(mockRepo)
			err := processor.ReverseTransaction(context.Background(), tt.txID, tt.reason)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCheck != nil {
					assert.True(t, tt.errCheck(err))
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
