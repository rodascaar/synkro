package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSynkroError_ErrorWithWrapped(t *testing.T) {
	inner := errors.New("file not found")
	se := &SynkroError{
		Code:    "TEST_CODE",
		Message: "test message",
		Err:     inner,
	}

	msg := se.Error()
	assert.Contains(t, msg, "test message")
	assert.Contains(t, msg, "file not found")
	assert.Contains(t, msg, "TEST_CODE")
}

func TestSynkroError_ErrorWithoutWrapped(t *testing.T) {
	se := &SynkroError{
		Code:    "SIMPLE",
		Message: "simple error",
	}

	msg := se.Error()
	assert.Contains(t, msg, "simple error")
	assert.Contains(t, msg, "SIMPLE")
	assert.NotContains(t, msg, "nil")
}

func TestSynkroError_Unwrap(t *testing.T) {
	inner := errors.New("inner error")
	se := &SynkroError{Err: inner}

	assert.Equal(t, inner, se.Unwrap())
	assert.True(t, errors.Is(se, inner))
}

func TestSynkroError_UnwrapNil(t *testing.T) {
	se := &SynkroError{}
	assert.Nil(t, se.Unwrap())
}

func TestDisplayError_SynkroError(t *testing.T) {
	se := &SynkroError{
		Code:    "DB_NOT_FOUND",
		Message: "Database not found",
		Help:    "Run: synkro init",
	}

	DisplayError(se)
}

func TestDisplayError_GenericError(t *testing.T) {
	generic := errors.New("something went wrong")
	DisplayError(generic)
}

func TestPredefinedErrors(t *testing.T) {
	assert.NotNil(t, ErrDatabaseNotFound)
	assert.Equal(t, "DB_NOT_FOUND", ErrDatabaseNotFound.Code)
	assert.Contains(t, ErrDatabaseNotFound.Message, "Database not found")

	assert.NotNil(t, ErrDatabaseLocked)
	assert.Equal(t, "DB_LOCKED", ErrDatabaseLocked.Code)

	assert.NotNil(t, ErrMCPNotConfigured)
	assert.Equal(t, "MCP_NOT_CONFIGURED", ErrMCPNotConfigured.Code)

	assert.NotNil(t, ErrFTS5NotAvailable)
	assert.Equal(t, "FTS5_NOT_AVAILABLE", ErrFTS5NotAvailable.Code)

	assert.NotNil(t, ErrTerminalTooSmall)
	assert.Equal(t, "TERMINAL_TOO_SMALL", ErrTerminalTooSmall.Code)
}
