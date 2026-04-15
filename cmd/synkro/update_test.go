package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1, v2 string
		want   int
	}{
		{"1.0.0", "1.0.0", 0},
		{"2.0.0", "1.0.0", 1},
		{"1.0.0", "2.0.0", -1},
		{"1.1.0", "1.0.0", 1},
		{"1.0.1", "1.0.0", 1},
		{"1.0.0", "1.0.1", -1},
		{"0.1.0", "0.1.0", 0},
		{"10.0.0", "9.9.9", 1},
		{"v1.0.0", "v1.0.0", 0},
		{"v2.0.0", "v1.0.0", 1},
		{"1.0.0-beta", "1.0.0", 0},
	}

	for _, tt := range tests {
		got := compareVersions(tt.v1, tt.v2)
		assert.Equal(t, tt.want, got, "compareVersions(%q, %q)", tt.v1, tt.v2)
	}
}

func TestGetPlatform(t *testing.T) {
	p := getPlatform()
	assert.NotEmpty(t, p.os)
	assert.NotEmpty(t, p.arch)
	assert.NotEmpty(t, p.binaryName)
}

func TestGetSynkroPath(t *testing.T) {
	path := getSynkroPath()
	assert.NotEmpty(t, path)
	assert.Contains(t, path, "synkro")
}
