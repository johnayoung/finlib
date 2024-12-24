package event

import (
	"context"
	"time"
)

// Event represents a domain event in the system
type Event struct {
	ID        string
	Type      string
	Timestamp time.Time
	Source    string
	Data      interface{}
	Metadata  map[string]interface{}
}

// Handler processes events
type Handler interface {
	Handle(ctx context.Context, event Event) error
}

// Publisher publishes events
type Publisher interface {
	Publish(ctx context.Context, event Event) error
}

// Bus manages event publishing and subscription
type Bus interface {
	Publisher
	Subscribe(eventType string, handler Handler) error
	Unsubscribe(eventType string, handler Handler) error
}

// Event types
const (
	// Transaction events
	TransactionCreated   = "transaction.created"
	TransactionValidated = "transaction.validated"
	TransactionPending   = "transaction.pending"
	TransactionPosted    = "transaction.posted"
	TransactionFailed    = "transaction.failed"
	TransactionVoided    = "transaction.voided"

	// Account events
	AccountBalanceUpdated = "account.balance.updated"
)

// ValidationEvent contains validation result details
type ValidationEvent struct {
	TransactionID string
	Valid         bool
	Errors        []string
	Warnings      []string
}

// TransactionStatusEvent contains transaction status change details
type TransactionStatusEvent struct {
	TransactionID string
	OldStatus     string
	NewStatus     string
	Reason        string
}

// BalanceUpdateEvent contains balance update details
type BalanceUpdateEvent struct {
	AccountID  string
	OldBalance interface{}
	NewBalance interface{}
	ChangeType string
}
