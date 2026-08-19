package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rodascaar/synkro/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildTestBinary(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "synkro-test")

	wd, err := os.Getwd()
	require.NoError(t, err)
	projectRoot := filepath.Dir(filepath.Dir(wd))

	cmd := exec.Command("go", "build", "-o", binPath, filepath.Join(projectRoot, "cmd/synkro"))
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(), "CGO_ENABLED=1")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "build failed: %s", string(output))
	return binPath
}

func TestInitCmd(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	bin := buildTestBinary(t)

	cmd := exec.Command(bin, "init", "--no-tutorial")
	cmd.Env = append(os.Environ(), "SYNKRO_DB_PATH="+dbPath, "SYNKRO_MODEL_TYPE=tfidf")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "init failed: %s", string(output))
	assert.Contains(t, string(output), "Database initialized")
	assert.FileExists(t, dbPath)
}

func TestAddCmd_MissingTitle(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	d, err := os.Create(dbPath)
	require.NoError(t, err)
	require.NoError(t, d.Close())

	bin := buildTestBinary(t)

	cmd := exec.Command(bin, "add", "--type", "note", "--content", "some content")
	cmd.Env = append(os.Environ(), "SYNKRO_DB_PATH="+dbPath, "SYNKRO_MODEL_TYPE=tfidf")
	output, err := cmd.CombinedOutput()

	assert.Error(t, err)
	assert.True(t, strings.Contains(string(output), "title is required") || strings.Contains(string(output), "required"), "unexpected output: %s", string(output))
}

func TestVersionCmd(t *testing.T) {
	bin := buildTestBinary(t)

	cmd := exec.Command(bin, "version")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "version failed: %s", string(output))
	assert.NotEmpty(t, string(output))
}

func TestHealthCmd(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	bin := buildTestBinary(t)

	cmd := exec.Command(bin, "init", "--no-tutorial")
	cmd.Env = append(os.Environ(), "SYNKRO_DB_PATH="+dbPath, "SYNKRO_MODEL_TYPE=tfidf")
	_, err := cmd.CombinedOutput()
	require.NoError(t, err)

	cmd = exec.Command(bin, "health")
	cmd.Env = append(os.Environ(), "SYNKRO_DB_PATH="+dbPath, "SYNKRO_MODEL_TYPE=tfidf")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "health check failed: %s", string(output))
	assert.NotEmpty(t, string(output))
}

func TestConfig_PreferredModelDefault(t *testing.T) {
	cfg := &config.Config{
		DatabasePath:   "memory.db",
		ModelType:      "tfidf",
		ModelDir:       "models",
		PreferredModel: "all-MiniLM-L6-v2",
	}

	assert.Equal(t, "all-MiniLM-L6-v2", cfg.PreferredModel)
	assert.Equal(t, "models", cfg.ModelDir)
}
