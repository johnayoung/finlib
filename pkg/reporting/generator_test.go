package reporting

import (
	"context"
	"testing"
	"time"

	"github.com/johnayoung/finlib/pkg/account"
	"github.com/johnayoung/finlib/pkg/money"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for dependencies
type mockReportCalculator struct {
	mock.Mock
}

func (m *mockReportCalculator) CalculateBalance(ctx context.Context, accountID string, period ReportPeriod) (money.Money, error) {
	args := m.Called(ctx, accountID, period)
	return args.Get(0).(money.Money), args.Error(1)
}

func (m *mockReportCalculator) CalculateChanges(ctx context.Context, accountID string, period ReportPeriod) (*BalanceChange, error) {
	args := m.Called(ctx, accountID, period)
	return args.Get(0).(*BalanceChange), args.Error(1)
}

func (m *mockReportCalculator) CalculateRatio(ctx context.Context, ratio RatioDefinition, period ReportPeriod) (decimal.Decimal, error) {
	args := m.Called(ctx, ratio, period)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

type mockReportStorage struct {
	mock.Mock
}

func (m *mockReportStorage) SaveDefinition(ctx context.Context, def *ReportDefinition) error {
	args := m.Called(ctx, def)
	return args.Error(0)
}

func (m *mockReportStorage) LoadDefinition(ctx context.Context, id string) (*ReportDefinition, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*ReportDefinition), args.Error(1)
}

func (m *mockReportStorage) ListDefinitions(ctx context.Context) ([]*ReportDefinition, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*ReportDefinition), args.Error(1)
}

func (m *mockReportStorage) DeleteDefinition(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Test cases
func TestNewReportGenerator(t *testing.T) {
	calculator := &mockReportCalculator{}
	storage := &mockReportStorage{}

	generator := NewReportGenerator(calculator, storage)
	assert.NotNil(t, generator, "NewReportGenerator should return a non-nil generator")
}

func TestValidateDefinition(t *testing.T) {
	calculator := &mockReportCalculator{}
	storage := &mockReportStorage{}
	generator := NewReportGenerator(calculator, storage)

	tests := []struct {
		name    string
		def     *ReportDefinition
		wantErr bool
	}{
		{
			name:    "nil definition",
			def:     nil,
			wantErr: true,
		},
		{
			name: "empty report type",
			def: &ReportDefinition{
				Name: "Test Report",
				Sections: []ReportSection{
					{
						ID:           "section1",
						Title:        "Section 1",
						AccountTypes: []account.AccountType{account.Asset},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty name",
			def: &ReportDefinition{
				Type: BalanceSheet,
				Sections: []ReportSection{
					{
						ID:           "section1",
						Title:        "Section 1",
						AccountTypes: []account.AccountType{account.Asset},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "no sections",
			def: &ReportDefinition{
				Type: BalanceSheet,
				Name: "Test Report",
			},
			wantErr: true,
		},
		{
			name: "valid definition",
			def: &ReportDefinition{
				Type: BalanceSheet,
				Name: "Test Report",
				Sections: []ReportSection{
					{
						ID:           "section1",
						Title:        "Section 1",
						AccountTypes: []account.AccountType{account.Asset},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := generator.ValidateDefinition(context.Background(), tt.def)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGenerateReport(t *testing.T) {
	calculator := &mockReportCalculator{}
	storage := &mockReportStorage{}
	generator := NewReportGenerator(calculator, storage)

	ctx := context.Background()
	testTime := time.Date(2024, 12, 24, 10, 0, 0, 0, time.UTC)

	// Setup test data
	def := &ReportDefinition{
		Type: BalanceSheet,
		Name: "Test Balance Sheet",
		Sections: []ReportSection{
			{
				ID:           "assets",
				Title:        "Assets",
				AccountTypes: []account.AccountType{account.Asset},
			},
		},
	}

	opts := ReportOptions{
		Period: ReportPeriod{
			Start: testTime.AddDate(0, -1, 0),
			End:   testTime,
		},
		Currency: "USD",
	}

	// Execute test
	report, err := generator.GenerateReport(ctx, def, opts)

	// Since getAccountsForSection is not implemented, we expect an error
	assert.Error(t, err)
	assert.Nil(t, report)
	assert.Contains(t, err.Error(), "getAccountsForSection not implemented")
}

func TestGetReportTypes(t *testing.T) {
	calculator := &mockReportCalculator{}
	storage := &mockReportStorage{}
	generator := NewReportGenerator(calculator, storage)

	types, err := generator.GetReportTypes(context.Background())
	assert.NoError(t, err)
	assert.NotEmpty(t, types)
	assert.Contains(t, types, BalanceSheet)
	assert.Contains(t, types, IncomeStatement)
	assert.Contains(t, types, CashFlow)
}

func TestSaveAndLoadDefinition(t *testing.T) {
	calculator := &mockReportCalculator{}
	storage := &mockReportStorage{}
	generator := NewReportGenerator(calculator, storage)

	ctx := context.Background()
	def := &ReportDefinition{
		Type: BalanceSheet,
		Name: "Test Report",
		Sections: []ReportSection{
			{
				ID:           "section1",
				Title:        "Section 1",
				AccountTypes: []account.AccountType{account.Asset},
			},
		},
	}

	// Setup mock expectations for save
	storage.On("SaveDefinition", ctx, def).Return(nil)

	// Test save
	err := generator.SaveDefinition(ctx, def)
	assert.NoError(t, err)
	storage.AssertExpectations(t)

	// Setup mock expectations for load
	storage.On("LoadDefinition", ctx, "test-id").Return(def, nil)

	// Test load
	loadedDef, err := generator.LoadDefinition(ctx, "test-id")
	assert.NoError(t, err)
	assert.Equal(t, def, loadedDef)
	storage.AssertExpectations(t)
}

func TestProcessSection(t *testing.T) {
	calculator := &mockReportCalculator{}
	storage := &mockReportStorage{}
	generator := NewReportGenerator(calculator, storage)

	ctx := context.Background()
	testTime := time.Date(2024, 12, 24, 10, 0, 0, 0, time.UTC)

	section := &ReportSection{
		ID:           "assets",
		Title:        "Assets",
		AccountTypes: []account.AccountType{account.Asset},
	}

	opts := ReportOptions{
		Period: ReportPeriod{
			Start: testTime.AddDate(0, -1, 0),
			End:   testTime,
		},
		Currency: "USD",
	}

	report := &Report{
		ID:       "test-report",
		Type:     BalanceSheet,
		Lines:    make([]*ReportLine, 0),
		Totals:   make(map[string]money.Money),
		Metadata: make(map[string]interface{}),
	}

	// Test error case when processing section
	err := generator.(*defaultReportGenerator).processSection(ctx, report, section, opts)
	assert.Error(t, err) // Should error because getAccountsForSection is not implemented
}
