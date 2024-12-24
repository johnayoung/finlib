package storage

import (
	"context"
	"time"
)

// Filter represents a query filter condition
type Filter struct {
	Field    string
	Operator string
	Value    interface{}
}

// Sort represents a sort order specification
type Sort struct {
	Field string
	Desc  bool
}

// Pagination represents pagination parameters
type Pagination struct {
	Offset int64
	Limit  int64
}

// Query encapsulates query parameters
type Query struct {
	Filters    []Filter
	Sort       []Sort
	Pagination *Pagination
}

// AuditEntry represents an audit log entry
type AuditEntry struct {
	ID            string
	EntityType    string
	EntityID      string
	Operation     string
	UserID        string
	Timestamp     time.Time
	PreviousState interface{}
	NewState      interface{}
	Metadata      map[string]interface{}
}

// VersionInfo represents entity version information
type VersionInfo struct {
	Version    int64
	ModifiedAt time.Time
	ModifiedBy string
}

// OptimisticLockError indicates a version conflict
type OptimisticLockError struct {
	EntityType      string
	EntityID        string
	CurrentVersion  int64
	ExpectedVersion int64
}

func (e *OptimisticLockError) Error() string {
	return "optimistic lock error: version mismatch"
}

// Transaction represents a database transaction
type Transaction interface {
	// Commit commits the transaction
	Commit(ctx context.Context) error

	// Rollback rolls back the transaction
	Rollback(ctx context.Context) error
}

// TransactionManager manages database transactions
type TransactionManager interface {
	// BeginTransaction starts a new transaction
	BeginTransaction(ctx context.Context) (Transaction, error)

	// WithTransaction executes a function within a transaction
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// Repository represents the base repository interface
type Repository interface {
	// Create creates a new entity
	Create(ctx context.Context, entity interface{}) error

	// Read retrieves an entity by ID
	Read(ctx context.Context, id string, entity interface{}) error

	// Update updates an existing entity
	Update(ctx context.Context, entity interface{}) error

	// Delete deletes an entity by ID
	Delete(ctx context.Context, id string) error

	// Query executes a query and returns matching entities
	Query(ctx context.Context, query Query, results interface{}) error

	// Count returns the number of entities matching a query
	Count(ctx context.Context, query Query) (int64, error)
}

// AuditableRepository adds auditing capabilities
type AuditableRepository interface {
	Repository

	// GetAuditTrail retrieves the audit trail for an entity
	GetAuditTrail(ctx context.Context, entityID string) ([]AuditEntry, error)

	// GetVersionInfo retrieves version information for an entity
	GetVersionInfo(ctx context.Context, entityID string) (*VersionInfo, error)
}

// SearchOptions represents search parameters
type SearchOptions struct {
	Query      string
	Filters    []Filter
	Pagination *Pagination
	Sort       []Sort
}

// SearchableRepository adds search capabilities
type SearchableRepository interface {
	Repository

	// Search performs a full-text search
	Search(ctx context.Context, options SearchOptions, results interface{}) error

	// SearchCount returns the number of search results
	SearchCount(ctx context.Context, options SearchOptions) (int64, error)
}

// BatchOperation represents a batch operation type
type BatchOperation string

const (
	BatchCreate BatchOperation = "CREATE"
	BatchUpdate BatchOperation = "UPDATE"
	BatchDelete BatchOperation = "DELETE"
)

// BatchItem represents an item in a batch operation
type BatchItem struct {
	Operation BatchOperation
	Entity    interface{}
	ID        string
}

// BatchResult represents the result of a batch operation
type BatchResult struct {
	Success bool
	Error   error
	ID      string
}

// BatchRepository adds batch operation capabilities
type BatchRepository interface {
	Repository

	// BatchExecute executes multiple operations in a batch
	BatchExecute(ctx context.Context, items []BatchItem) []BatchResult
}
