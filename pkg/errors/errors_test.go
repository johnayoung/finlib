package errors

import (
	"errors"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
)

func TestFinancialError(t *testing.T) {
	t.Run("Error Creation", func(t *testing.T) {
		now := time.Now()
		err := &FinancialError{
			ID:        "err123",
			Code:      "INVALID_AMOUNT",
			Message:   "Invalid transaction amount",
			Details:   "Amount must be positive",
			Category:  ValidationError,
			Severity:  Error,
			Timestamp: now,
			Source:    "TransactionService",
			Retryable: false,
		}

		assert.Equal(t, "err123", err.ID)
		assert.Equal(t, "INVALID_AMOUNT", err.Code)
		assert.Equal(t, "Invalid transaction amount", err.Message)
		assert.Equal(t, "Amount must be positive", err.Details)
		assert.Equal(t, ValidationError, err.Category)
		assert.Equal(t, Error, err.Severity)
		assert.Equal(t, now, err.Timestamp)
		assert.Equal(t, "TransactionService", err.Source)
		assert.False(t, err.Retryable)
	})

	t.Run("Error Interface", func(t *testing.T) {
		err := &FinancialError{
			Category: ValidationError,
			Code:     "TEST_ERROR",
			Message:  "Test error message",
			Details:  "Test error details",
		}

		expectedMsg := "[VALIDATION-TEST_ERROR] Test error message: Test error details"
		assert.Equal(t, expectedMsg, err.Error())
	})

	t.Run("Error Wrapping", func(t *testing.T) {
		originalErr := errors.New("original error")
		wrappedErr := Wrap(
			originalErr,
			"Something went wrong",
			TechnicalError,
			Error,
		)

		assert.Equal(t, "Something went wrong", wrappedErr.Message)
		assert.Equal(t, TechnicalError, wrappedErr.Category)
		assert.Equal(t, Error, wrappedErr.Severity)
		assert.Equal(t, originalErr, wrappedErr.Cause)
	})

	t.Run("Error Is", func(t *testing.T) {
		err1 := &FinancialError{
			Code: "TEST_ERROR",
		}
		err2 := &FinancialError{
			Code: "TEST_ERROR",
		}
		err3 := &FinancialError{
			Code: "OTHER_ERROR",
		}

		assert.True(t, err1.Is(err2))
		assert.False(t, err1.Is(err3))
	})

	t.Run("Error Unwrap", func(t *testing.T) {
		cause := errors.New("cause error")
		err := &FinancialError{
			Message: "wrapper error",
			Cause:   cause,
		}

		assert.Equal(t, cause, err.Unwrap())
	})
}

func TestErrorCategories(t *testing.T) {
	t.Run("Error Categories", func(t *testing.T) {
		assert.Equal(t, ErrorCategory("VALIDATION"), ValidationError)
		assert.Equal(t, ErrorCategory("BUSINESS"), BusinessError)
		assert.Equal(t, ErrorCategory("TECHNICAL"), TechnicalError)
		assert.Equal(t, ErrorCategory("SECURITY"), SecurityError)
		assert.Equal(t, ErrorCategory("CONCURRENCY"), ConcurrencyError)
	})
}

func TestErrorSeverities(t *testing.T) {
	t.Run("Error Severities", func(t *testing.T) {
		assert.Equal(t, ErrorSeverity("INFO"), Info)
		assert.Equal(t, ErrorSeverity("WARNING"), Warning)
		assert.Equal(t, ErrorSeverity("ERROR"), Error)
		assert.Equal(t, ErrorSeverity("CRITICAL"), Critical)
		assert.Equal(t, ErrorSeverity("FATAL"), Fatal)
	})
}
