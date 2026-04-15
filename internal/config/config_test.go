package config

import (
	"encoding/json"
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

func TestLoad_FromJSONFile(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".synkro")
	configPath := filepath.Join(configDir, "config.json")

	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	fileCfg := Config{
		DatabasePath: "file.db",
		MaxTokens:    6000,
		ModelType:    "onnx",
		configPath:   configPath,
	}
	data, _ := json.MarshalIndent(fileCfg, "", "  ")
	err = os.WriteFile(configPath, append(data, '\n'), 0644)
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

func TestLoad_FromKeyValueFile_MigratesToJSON(t *testing.T) {
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

		migratedData, err := os.ReadFile(configPath)
		require.NoError(t, err)
		assert.True(t, json.Valid(migratedData), "config should be migrated to JSON")
	}
}

func TestLoad_InvalidFileValues(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".synkro")
	configPath := filepath.Join(configDir, "config.json")

	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	invalidJSON := `{ "max_tokens": "notanumber" }`
	err = os.WriteFile(configPath, []byte(invalidJSON), 0644)
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

func TestSave_WritesJSON(t *testing.T) {
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
	assert.True(t, json.Valid(data), "saved config must be valid JSON")

	var loaded Config
	err = json.Unmarshal(data, &loaded)
	require.NoError(t, err)
	assert.Equal(t, "custom.db", loaded.DatabasePath)
	assert.True(t, loaded.Debug)
	assert.Equal(t, 8000, loaded.MaxTokens)
	assert.Equal(t, "onnx", loaded.ModelType)
	assert.False(t, loaded.AutoUpdateCheck)
	assert.Equal(t, 100, loaded.LastUpdateCheck)
}

func TestSave_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	original := &Config{
		DatabasePath:        "test.db",
		Debug:               true,
		MaxTokens:           2000,
		SessionBuffer:       10,
		CacheSize:           500,
		SimilarityThreshold: 0.7,
		EmbeddingDim:        512,
		ModelType:           "onnx",
		AutoUpdateCheck:     false,
		CheckUpdateOnStart:  false,
		LastUpdateCheck:     42,
		configPath:          filepath.Join(tmpDir, "config.json"),
	}

	err := Save(original)
	require.NoError(t, err)

	reloaded := &Config{configPath: original.configPath}
	reloaded.loadFromFile()

	assert.Equal(t, original.DatabasePath, reloaded.DatabasePath)
	assert.Equal(t, original.Debug, reloaded.Debug)
	assert.Equal(t, original.MaxTokens, reloaded.MaxTokens)
	assert.Equal(t, original.SessionBuffer, reloaded.SessionBuffer)
	assert.Equal(t, original.CacheSize, reloaded.CacheSize)
	assert.Equal(t, original.SimilarityThreshold, reloaded.SimilarityThreshold)
	assert.Equal(t, original.EmbeddingDim, reloaded.EmbeddingDim)
	assert.Equal(t, original.ModelType, reloaded.ModelType)
	assert.Equal(t, original.AutoUpdateCheck, reloaded.AutoUpdateCheck)
	assert.Equal(t, original.CheckUpdateOnStart, reloaded.CheckUpdateOnStart)
	assert.Equal(t, original.LastUpdateCheck, reloaded.LastUpdateCheck)
}

func TestLoad_MissingFile(t *testing.T) {
	cfg := &Config{configPath: "/nonexistent/path/config.json"}
	cfg.loadFromFile()
	assert.Equal(t, "", cfg.DatabasePath)
}

func TestConfigPathFieldNotSerialized(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		DatabasePath: "test.db",
		configPath:   filepath.Join(tmpDir, "config.json"),
	}

	err := Save(cfg)
	require.NoError(t, err)

	data, err := os.ReadFile(cfg.configPath)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "configPath")
	assert.NotContains(t, string(data), "config_path")
}
