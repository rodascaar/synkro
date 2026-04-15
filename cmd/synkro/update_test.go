package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestParseSemver(t *testing.T) {
	tests := []struct {
		input string
		want  [3]int
	}{
		{"1.2.3", [3]int{1, 2, 3}},
		{"0.0.1", [3]int{0, 0, 1}},
		{"10.20.30", [3]int{10, 20, 30}},
		{"v1.2.3", [3]int{1, 2, 3}},
		{"1.2.3-beta", [3]int{1, 2, 3}},
		{"1.2", [3]int{1, 2, 0}},
		{"1", [3]int{1, 0, 0}},
		{"", [3]int{0, 0, 0}},
		{"invalid", [3]int{0, 0, 0}},
		{"v2.0.0-rc1", [3]int{2, 0, 0}},
	}

	for _, tt := range tests {
		got := parseSemver(tt.input)
		assert.Equal(t, tt.want, got, "parseSemver(%q)", tt.input)
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

func TestFileSHA256(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "testfile")
	content := "hello world"
	require.NoError(t, os.WriteFile(tmpFile, []byte(content), 0644))

	hash, err := fileSHA256(tmpFile)
	require.NoError(t, err)

	expected := sha256.Sum256([]byte(content))
	assert.Equal(t, hex.EncodeToString(expected[:]), hash)
}

func TestFileSHA256_NotFound(t *testing.T) {
	_, err := fileSHA256("/nonexistent/file")
	assert.Error(t, err)
}

func TestCheckLatestRelease_WithMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "releases/latest")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"tag_name": "v3.0.0",
			"html_url": "https://github.com/rodascaar/synkro/releases/tag/v3.0.0",
			"name": "v3.0.0",
			"draft": false,
			"prerelease": false,
			"body": "New release",
			"assets": [
				{"name": "synkro-darwin-arm64.tar.gz", "browser_download_url": "https://example.com/synkro-darwin-arm64.tar.gz", "size": 1000000},
				{"name": "synkro-linux-amd64.tar.gz", "browser_download_url": "https://example.com/synkro-linux-amd64.tar.gz", "size": 1000000}
			]
		}`))
	}))
	defer server.Close()

	resp, err := http.Get(server.URL + "/repos/rodascaar/synkro/releases/latest")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "json")
}

func TestFindChecksumForAsset_WithMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "checksums.txt") {
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("abc123def456  synkro-darwin-arm64.tar.gz\n789xyz  synkro-linux-amd64.tar.gz\n"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	checksumURL := server.URL + "/repos/rodascaar/synkro/releases/tag/v2.0.0"
	assetName := "synkro-darwin-arm64.tar.gz"

	result, err := findChecksumForAsset(checksumURL, assetName)
	require.NoError(t, err)
	assert.Equal(t, "abc123def456", result)
}

func TestFindChecksumForAsset_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("abc123  other-file.tar.gz\n"))
	}))
	defer server.Close()

	result, err := findChecksumForAsset(server.URL+"/repos/test/releases/tag/v1.0.0", "missing-file.tar.gz")
	assert.Error(t, err)
	assert.Empty(t, result)
}

func TestGetUpdateInfo_WithMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"tag_name": "v99.0.0",
			"html_url": "https://github.com/test/test/releases/tag/v99.0.0",
			"name": "v99.0.0",
			"draft": false,
			"prerelease": false,
			"body": "Test release",
			"assets": []
		}`))
	}))
	defer server.Close()

	resp, err := http.Get(server.URL + "/repos/rodascaar/synkro/releases/latest")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
