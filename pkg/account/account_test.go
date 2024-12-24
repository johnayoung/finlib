package account

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
)

func TestAccountTypes(t *testing.T) {
	t.Run("Account Type Constants", func(t *testing.T) {
		assert.Equal(t, AccountType("ASSET"), Asset)
		assert.Equal(t, AccountType("LIABILITY"), Liability)
		assert.Equal(t, AccountType("EQUITY"), Equity)
		assert.Equal(t, AccountType("REVENUE"), Revenue)
		assert.Equal(t, AccountType("EXPENSE"), Expense)
	})
}

func TestAccount(t *testing.T) {
	t.Run("Account Creation", func(t *testing.T) {
		now := time.Now()
		parentID := "parent123"
		account := Account{
			ID:           "acc123",
			Code:         "1000",
			Name:         "Cash Account",
			Type:         Asset,
			ParentID:     &parentID,
			Created:      now,
			LastModified: now,
			MetaData: map[string]interface{}{
				"department": "Treasury",
			},
		}

		assert.Equal(t, "acc123", account.ID)
		assert.Equal(t, "1000", account.Code)
		assert.Equal(t, "Cash Account", account.Name)
		assert.Equal(t, Asset, account.Type)
		assert.Equal(t, &parentID, account.ParentID)
		assert.Equal(t, now, account.Created)
		assert.Equal(t, now, account.LastModified)
		assert.Equal(t, "Treasury", account.MetaData["department"])
	})
}

func TestStatus(t *testing.T) {
	t.Run("Status Creation", func(t *testing.T) {
		now := time.Now()
		status := Status{
			Active:       true,
			Locked:       false,
			StatusReason: "Account activated",
			LastUpdated:  now,
		}

		assert.True(t, status.Active)
		assert.False(t, status.Locked)
		assert.Equal(t, "Account activated", status.StatusReason)
		assert.Equal(t, now, status.LastUpdated)
	})
}

func TestValidationRule(t *testing.T) {
	t.Run("Validation Rule Creation", func(t *testing.T) {
		rule := ValidationRule{
			ID:          "rule123",
			Description: "Minimum balance rule",
			Type:        "balance",
			Blocking:    true,
		}

		assert.Equal(t, "rule123", rule.ID)
		assert.Equal(t, "Minimum balance rule", rule.Description)
		assert.Equal(t, "balance", rule.Type)
		assert.True(t, rule.Blocking)
	})
}

func TestBalance(t *testing.T) {
	t.Run("Balance Creation", func(t *testing.T) {
		now := time.Now()
		balance := Balance{
			AccountID:         "acc123",
			AsOf:             now,
			Amount:           "1000.50",
			Currency:         "USD",
			LastTransactionID: "tx789",
		}

		assert.Equal(t, "acc123", balance.AccountID)
		assert.Equal(t, now, balance.AsOf)
		assert.Equal(t, "1000.50", balance.Amount)
		assert.Equal(t, "USD", balance.Currency)
		assert.Equal(t, "tx789", balance.LastTransactionID)
	})
}
