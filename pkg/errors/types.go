package errors

import (
	"fmt"
	"time"
)

// ErrorCategory classifies the type of error
type ErrorCategory string

const (
	ValidationError  ErrorCategory = "VALIDATION"
	BusinessError    ErrorCategory = "BUSINESS"
	TechnicalError   ErrorCategory = "TECHNICAL"
	SecurityError    ErrorCategory = "SECURITY"
	ConcurrencyError ErrorCategory = "CONCURRENCY"
)

// ErrorSeverity indicates the impact level of an error
type ErrorSeverity string

const (
	Info     ErrorSeverity = "INFO"
	Warning  ErrorSeverity = "WARNING"
	Error    ErrorSeverity = "ERROR"
	Critical ErrorSeverity = "CRITICAL"
	Fatal    ErrorSeverity = "FATAL"
)

// FinancialError represents a domain-specific error
type FinancialError struct {
	// Unique error identifier
	ID string
	// Error code for categorization
	Code string
	// Human-readable message
	Message string
	// Detailed error description
	Details string
	// Error category
	Category ErrorCategory
	// Error severity
	Severity ErrorSeverity
	// Time the error occurred
	Timestamp time.Time
	// Component where the error originated
	Source string
	// Whether the operation can be retried
	Retryable bool
	// Original error if wrapped
	Cause error
}

// Error implements the error interface
func (e *FinancialError) Error() string {
	return fmt.Sprintf("[%s-%s] %s: %s", e.Category, e.Code, e.Message, e.Details)
}

// Wrap creates a new FinancialError wrapping an existing error
func Wrap(err error, message string, category ErrorCategory, severity ErrorSeverity) *FinancialError {
	return &FinancialError{
		Message:   message,
		Category:  category,
		Severity:  severity,
		Timestamp: time.Now(),
		Cause:     err,
	}
}

// Is implements error matching for wrapped errors
func (e *FinancialError) Is(target error) bool {
	t, ok := target.(*FinancialError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// Unwrap returns the wrapped error
func (e *FinancialError) Unwrap() error {
	return e.Cause
}
