package validation

import (
	"context"
	"fmt"
)

// ValidationSeverity indicates the severity of a validation result
type ValidationSeverity string

const (
	Error   ValidationSeverity = "ERROR"
	Warning ValidationSeverity = "WARNING"
	Info    ValidationSeverity = "INFO"
)

// ValidationResult represents the outcome of a validation check
type ValidationResult struct {
	Code      string
	Message   string
	Severity  ValidationSeverity
	Field     string
	Metadata  map[string]interface{}
}

// ValidationRule describes a specific validation rule
type ValidationRule struct {
	ID          string
	Description string
	Severity    ValidationSeverity
	Category    string
}

// Validator defines the interface for implementing validation rules
type Validator interface {
	// Validate performs the validation and returns any violations
	Validate(ctx context.Context, obj interface{}) ([]ValidationResult, error)

	// GetRules returns the rules this validator checks
	GetRules() []ValidationRule

	// Priority determines the order of validator execution
	Priority() int
}

// ValidationEngine coordinates validation across the system
type ValidationEngine interface {
	// RegisterValidator adds a new validator to the engine
	RegisterValidator(validator Validator) error

	// Validate runs all applicable validators against an object
	Validate(ctx context.Context, obj interface{}) ([]ValidationResult, error)

	// GetValidators returns all registered validators
	GetValidators() []Validator
}

// ValidationError represents a validation-specific error
type ValidationError struct {
	Results []ValidationResult
}

func (e *ValidationError) Error() string {
	if len(e.Results) == 0 {
		return "validation failed"
	}
	return fmt.Sprintf("validation failed: %s", e.Results[0].Message)
}

// NewValidationError creates a new ValidationError
func NewValidationError(results []ValidationResult) *ValidationError {
	return &ValidationError{Results: results}
}
