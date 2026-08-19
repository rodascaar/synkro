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

func TestWrap_NilError(t *testing.T) {
	assert.Nil(t, Wrap(nil, "CODE", "msg", "help"))
}

func TestWrap_ReturnsSynkroError(t *testing.T) {
	inner := errors.New("inner")
	se := Wrap(inner, "CODE", "message", "help")
	assert.NotNil(t, se)
	assert.Equal(t, "CODE", se.Code)
	assert.Equal(t, inner, se.Err)
}