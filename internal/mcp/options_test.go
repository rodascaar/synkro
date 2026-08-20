package mcp

import (
	"testing"
)

func TestNewServer_Options(t *testing.T) {
	s := NewServer(nil, nil, nil, nil,
		WithVersion("9.9.9"),
		WithEmbeddingType("onnx"),
		WithConflictThreshold(0.5),
	)

	if s.serverVersion != "9.9.9" {
		t.Fatalf("expected version 9.9.9, got %q", s.serverVersion)
	}
	if s.embeddingType != "onnx" {
		t.Fatalf("expected embedding type onnx, got %q", s.embeddingType)
	}
	if s.conflictThreshold != 0.5 {
		t.Fatalf("expected conflict threshold 0.5, got %v", s.conflictThreshold)
	}
}

func TestNewServer_Defaults(t *testing.T) {
	s := NewServer(nil, nil, nil, nil)

	if s.serverVersion != "1.0" {
		t.Fatalf("expected default version 1.0, got %q", s.serverVersion)
	}
	if s.conflictThreshold != defaultConflictThreshold {
		t.Fatalf("expected default conflict threshold %v, got %v", defaultConflictThreshold, s.conflictThreshold)
	}
}

func TestNewServer_InvalidThreshold(t *testing.T) {
	s := NewServer(nil, nil, nil, nil, WithConflictThreshold(1.5))

	if s.conflictThreshold != defaultConflictThreshold {
		t.Fatalf("out-of-range threshold should be ignored, got %v", s.conflictThreshold)
	}
}
