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

func TestSetGraph(t *testing.T) {
	SetGraph(nil)
	assert.Nil(t, globalGraph)
}

func TestSetGlobalRepo(t *testing.T) {
	SetGlobalRepo(nil)
	assert.Nil(t, globalRepo)
}

func TestSetSessionTracker(t *testing.T) {
	SetSessionTracker(nil)
	assert.Nil(t, globalSessionTracker)
}

func TestSetContextPruner(t *testing.T) {
	SetContextPruner(nil)
	assert.Nil(t, globalContextPruner)
}

func TestAddRelationInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   AddRelationInput
		wantErr bool
	}{
		{"valid", AddRelationInput{SourceID: "a", TargetID: "b", Type: "related_to"}, false},
		{"missing source", AddRelationInput{TargetID: "b", Type: "related_to"}, true},
		{"missing target", AddRelationInput{SourceID: "a", Type: "related_to"}, true},
		{"invalid type", AddRelationInput{SourceID: "a", TargetID: "b", Type: "invalid"}, true},
		{"valid extends", AddRelationInput{SourceID: "a", TargetID: "b", Type: "extends"}, false},
		{"valid depends_on", AddRelationInput{SourceID: "a", TargetID: "b", Type: "depends_on"}, false},
		{"valid part_of", AddRelationInput{SourceID: "a", TargetID: "b", Type: "part_of"}, false},
		{"valid conflicts_with", AddRelationInput{SourceID: "a", TargetID: "b", Type: "conflicts_with"}, false},
		{"valid example_of", AddRelationInput{SourceID: "a", TargetID: "b", Type: "example_of"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetGraph(nil)
			var buf BufferWriter
			err := AddRelationHandler(tt.input, &buf)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, buf.String(), "graph not available")
			}
		})
	}
}

func TestGetRelationsHandler_NoGraph(t *testing.T) {
	SetGraph(nil)
	var buf BufferWriter
	err := GetRelationsHandler(GetRelationsInput{MemoryID: "test"}, &buf)
	assert.Error(t, err)
	assert.Contains(t, buf.String(), "graph not available")
}

func TestGetRelationsHandler_NoMemoryID(t *testing.T) {
	SetGraph(nil)
	var buf BufferWriter
	err := GetRelationsHandler(GetRelationsInput{}, &buf)
	assert.Error(t, err)
	assert.Contains(t, buf.String(), "graph not available")
}

func TestDeleteRelationHandler_NoGraph(t *testing.T) {
	SetGraph(nil)
	var buf BufferWriter
	err := DeleteRelationHandler(DeleteRelationInput{SourceID: "a", TargetID: "b"}, &buf)
	assert.Error(t, err)
	assert.Contains(t, buf.String(), "graph not available")
}

func TestFindPathHandler_NoGraph(t *testing.T) {
	SetGraph(nil)
	var buf BufferWriter
	err := FindPathHandler(FindPathInput{FromID: "a", ToID: "b"}, &buf)
	assert.Error(t, err)
	assert.Contains(t, buf.String(), "graph not available")
}
