package reporting

import (
	"context"
)

// ReportStorage defines the interface for storing and retrieving report definitions
type ReportStorage interface {
	// SaveDefinition stores a report definition
	SaveDefinition(ctx context.Context, def *ReportDefinition) error

	// LoadDefinition retrieves a stored report definition
	LoadDefinition(ctx context.Context, id string) (*ReportDefinition, error)

	// ListDefinitions retrieves all stored report definitions
	ListDefinitions(ctx context.Context) ([]*ReportDefinition, error)

	// DeleteDefinition removes a stored report definition
	DeleteDefinition(ctx context.Context, id string) error
}
