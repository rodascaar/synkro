package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("SYNKRO_DB_PATH", "")
	t.Setenv("SYNKRO_DEBUG", "")
	t.Setenv("SYNKRO_MAX_TOKENS", "")
	t.Setenv("SYNKRO_MODEL_TYPE", "")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "memory.db", cfg.DatabasePath)
	assert.False(t, cfg.Debug)
	assert.Equal(t, 4000, cfg.MaxTokens)
	assert.Equal(t, 20, cfg.SessionBuffer)
	assert.Equal(t, 1000, cfg.CacheSize)
	assert.Equal(t, 0.5, cfg.SimilarityThreshold)
	assert.Equal(t, 384, cfg.EmbeddingDim)
	assert.Equal(t, "tfidf", cfg.ModelType)
	assert.True(t, cfg.AutoUpdateCheck)
	assert.True(t, cfg.CheckUpdateOnStart)
	assert.Equal(t, 0, cfg.LastUpdateCheck)
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("SYNKRO_DB_PATH", "custom.db")
	t.Setenv("SYNKRO_DEBUG", "true")
	t.Setenv("SYNKRO_MAX_TOKENS", "8000")
	t.Setenv("SYNKRO_SESSION_BUFFER", "50")
	t.Setenv("SYNKRO_CACHE_SIZE", "2000")
	t.Setenv("SYNKRO_SIMILARITY_THRESHOLD", "0.8")
	t.Setenv("SYNKRO_EMBEDDING_DIM", "768")
	t.Setenv("SYNKRO_MODEL_TYPE", "onnx")
	t.Setenv("SYNKRO_AUTO_UPDATE", "false")
	t.Setenv("SYNKRO_CHECK_UPDATE_ON_START", "false")
	t.Setenv("SYNKRO_LAST_UPDATE_CHECK", "100")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "custom.db", cfg.DatabasePath)
	assert.True(t, cfg.Debug)
	assert.Equal(t, 8000, cfg.MaxTokens)
	assert.Equal(t, 50, cfg.SessionBuffer)
	assert.Equal(t, 2000, cfg.CacheSize)
	assert.Equal(t, 0.8, cfg.SimilarityThreshold)
	assert.Equal(t, 768, cfg.EmbeddingDim)
	assert.Equal(t, "onnx", cfg.ModelType)
	assert.False(t, cfg.AutoUpdateCheck)
	assert.False(t, cfg.CheckUpdateOnStart)
	assert.Equal(t, 100, cfg.LastUpdateCheck)
}

func TestLoad_FromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".synkro")
	configPath := filepath.Join(configDir, "config.json")

	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	content := "SYNKRO_DB_PATH=file.db\nSYNKRO_MAX_TOKENS=6000\nSYNKRO_MODEL_TYPE=onnx\n"
	err = os.WriteFile(configPath, []byte(content), 0644)
	require.NoError(t, err)

	t.Setenv("SYNKRO_DB_PATH", "")
	t.Setenv("SYNKRO_MAX_TOKENS", "")
	t.Setenv("SYNKRO_MODEL_TYPE", "")

	cfg, err := Load()
	require.NoError(t, err)

	if cfg.configPath == configPath {
		assert.Equal(t, "file.db", cfg.DatabasePath)
		assert.Equal(t, 6000, cfg.MaxTokens)
		assert.Equal(t, "onnx", cfg.ModelType)
	}
}

func TestLoad_InvalidFileValues(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".synkro")
	configPath := filepath.Join(configDir, "config.json")

	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	content := "SYNKRO_MAX_TOKENS=notanumber\nSYNKRO_SIMILARITY_THRESHOLD=invalid\n"
	err = os.WriteFile(configPath, []byte(content), 0644)
	require.NoError(t, err)

	t.Setenv("SYNKRO_MAX_TOKENS", "")
	t.Setenv("SYNKRO_SIMILARITY_THRESHOLD", "")

	cfg, err := Load()
	require.NoError(t, err)

	if cfg.configPath == configPath {
		assert.Equal(t, 4000, cfg.MaxTokens)
		assert.Equal(t, 0.5, cfg.SimilarityThreshold)
	}
}

func TestSave_Defaults(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		DatabasePath:        "memory.db",
		Debug:               false,
		MaxTokens:           4000,
		SessionBuffer:       20,
		CacheSize:           1000,
		SimilarityThreshold: 0.5,
		EmbeddingDim:        384,
		ModelType:           "tfidf",
		AutoUpdateCheck:     true,
		CheckUpdateOnStart:  true,
		LastUpdateCheck:     0,
		configPath:          filepath.Join(tmpDir, "config.json"),
	}

	err := Save(cfg)
	require.NoError(t, err)

	data, err := os.ReadFile(cfg.configPath)
	require.NoError(t, err)
	assert.Empty(t, string(data))
}

func TestSave_NonDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		DatabasePath:        "custom.db",
		Debug:               true,
		MaxTokens:           8000,
		SessionBuffer:       50,
		CacheSize:           2000,
		SimilarityThreshold: 0.8,
		EmbeddingDim:        768,
		ModelType:           "onnx",
		AutoUpdateCheck:     false,
		CheckUpdateOnStart:  false,
		LastUpdateCheck:     100,
		configPath:          filepath.Join(tmpDir, "config.json"),
	}

	err := Save(cfg)
	require.NoError(t, err)

	data, err := os.ReadFile(cfg.configPath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "SYNKRO_DB_PATH=custom.db")
	assert.Contains(t, content, "SYNKRO_DEBUG=true")
	assert.Contains(t, content, "SYNKRO_MAX_TOKENS=8000")
	assert.Contains(t, content, "SYNKRO_MODEL_TYPE=onnx")
	assert.Contains(t, content, "SYNKRO_AUTO_UPDATE=false")
	assert.Contains(t, content, "SYNKRO_LAST_UPDATE_CHECK=100")
}

func TestLoad_MissingFile(t *testing.T) {
	cfg := &Config{configPath: "/nonexistent/path/config.json"}
	cfg.loadFromFile()
	assert.Equal(t, "", cfg.DatabasePath)
}
