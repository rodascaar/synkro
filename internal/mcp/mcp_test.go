package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfidenceLevel(t *testing.T) {
	assert.Equal(t, "high", getConfidenceLevel(0.9))
	assert.Equal(t, "high", getConfidenceLevel(0.8))
	assert.Equal(t, "medium", getConfidenceLevel(0.5))
	assert.Equal(t, "medium", getConfidenceLevel(0.6))
	assert.Equal(t, "low", getConfidenceLevel(0.3))
	assert.Equal(t, "low", getConfidenceLevel(0.0))
}

func TestBufferWriter(t *testing.T) {
	b := &BufferWriter{}

	n, err := b.Write([]byte("hello"))
	assert.NoError(t, err)
	assert.Equal(t, 5, n)

	n, err = b.Write([]byte(" world"))
	assert.NoError(t, err)
	assert.Equal(t, 6, n)

	assert.Equal(t, "hello world", b.String())
}

func TestBufferWriter_Empty(t *testing.T) {
	b := &BufferWriter{}
	assert.Equal(t, "", b.String())
}

func TestErrorResult(t *testing.T) {
	result := errorResult(assert.AnError)
	assert.True(t, result.IsError)
	assert.Len(t, result.Content, 1)
}
