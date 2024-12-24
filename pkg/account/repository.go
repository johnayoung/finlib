package account

import (
	"context"
)

// Repository defines the interface for account data persistence
type Repository interface {
	// Create creates a new account
	Create(ctx context.Context, entity interface{}) error

	// Read retrieves an account by ID
	Read(ctx context.Context, id string, entity interface{}) error

	// Update updates an existing account
	Update(ctx context.Context, entity interface{}) error

	// Delete deletes an account by ID
	Delete(ctx context.Context, id string) error

	// Query executes a query and returns matching accounts
	Query(ctx context.Context, query interface{}, results interface{}) error
}
