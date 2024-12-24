package validation

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// BasicValidationEngine provides a simple implementation of ValidationEngine
type BasicValidationEngine struct {
	validators []Validator
	mu        sync.RWMutex
}

// NewBasicValidationEngine creates a new BasicValidationEngine
func NewBasicValidationEngine() *BasicValidationEngine {
	return &BasicValidationEngine{
		validators: make([]Validator, 0),
	}
}

// RegisterValidator adds a new validator to the engine
func (e *BasicValidationEngine) RegisterValidator(validator Validator) error {
	if validator == nil {
		return fmt.Errorf("validator cannot be nil")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// Add validator and sort by priority
	e.validators = append(e.validators, validator)
	sort.Slice(e.validators, func(i, j int) bool {
		return e.validators[i].Priority() < e.validators[j].Priority()
	})

	return nil
}

// Validate runs all applicable validators against an object
func (e *BasicValidationEngine) Validate(ctx context.Context, obj interface{}) ([]ValidationResult, error) {
	if obj == nil {
		return nil, fmt.Errorf("cannot validate nil object")
	}

	e.mu.RLock()
	validators := make([]Validator, len(e.validators))
	copy(validators, e.validators)
	e.mu.RUnlock()

	var allResults []ValidationResult
	var hasErrors bool

	// Run each validator in priority order
	for _, validator := range validators {
		results, err := validator.Validate(ctx, obj)
		if err != nil {
			return nil, fmt.Errorf("validator error: %w", err)
		}

		allResults = append(allResults, results...)

		// Check for error severity results
		for _, result := range results {
			if result.Severity == Error {
				hasErrors = true
			}
		}
	}

	// If we have any error severity results, return them as a ValidationError
	if hasErrors {
		return allResults, NewValidationError(allResults)
	}

	return allResults, nil
}

// GetValidators returns all registered validators
func (e *BasicValidationEngine) GetValidators() []Validator {
	e.mu.RLock()
	defer e.mu.RUnlock()

	validators := make([]Validator, len(e.validators))
	copy(validators, e.validators)
	return validators
}
