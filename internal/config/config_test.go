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
	t.Setenv("SYNKRO_MODEL_TYPE", "")

	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".synkro")
	configPath := filepath.Join(configDir, "config.json")

	backup := ""
	if data, err := os.ReadFile(configPath); err == nil {
		backup = string(data)
	}
	_ = os.Remove(configPath)

	t.Cleanup(func() {
		if backup != "" {
			_ = os.WriteFile(configPath, []byte(backup), 0644)
		}
	})
	t.Setenv("SYNKRO_CONFIG_PATH", "")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "memory.db", cfg.DatabasePath)
	assert.Equal(t, "tfidf", cfg.ModelType)
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("SYNKRO_DB_PATH", "custom.db")
	t.Setenv("SYNKRO_MODEL_TYPE", "onnx")
	t.Setenv("SYNKRO_MODEL_DIR", "/tmp/models")
	t.Setenv("SYNKRO_PREFERRED_MODEL", "stsb-roberta-base-v2")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "custom.db", cfg.DatabasePath)
	assert.Equal(t, "onnx", cfg.ModelType)
	assert.Equal(t, "/tmp/models", cfg.ModelDir)
	assert.Equal(t, "stsb-roberta-base-v2", cfg.PreferredModel)
}

func TestLoad_FromJSONFile(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".synkro")
	configPath := filepath.Join(configDir, "config.json")

	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	fileCfg := Config{
		DatabasePath: "file.db",
		ModelType:    "onnx",
		configPath:   configPath,
	}
	data, _ := json.MarshalIndent(fileCfg, "", "  ")
	err = os.WriteFile(configPath, append(data, '\n'), 0644)
	require.NoError(t, err)

	t.Setenv("SYNKRO_DB_PATH", "")
	t.Setenv("SYNKRO_MODEL_TYPE", "")

	cfg, err := Load()
	require.NoError(t, err)

	if cfg.configPath == configPath {
		assert.Equal(t, "file.db", cfg.DatabasePath)
		assert.Equal(t, "onnx", cfg.ModelType)
	}
}

func TestLoad_FromKeyValueFile_MigratesToJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".synkro")
	configPath := filepath.Join(configDir, "config.json")

	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	content := "SYNKRO_DB_PATH=file.db\nSYNKRO_MODEL_TYPE=onnx\n"
	err = os.WriteFile(configPath, []byte(content), 0644)
	require.NoError(t, err)

	t.Setenv("SYNKRO_DB_PATH", "")
	t.Setenv("SYNKRO_MODEL_TYPE", "")

	cfg, err := Load()
	require.NoError(t, err)

	if cfg.configPath == configPath {
		assert.Equal(t, "file.db", cfg.DatabasePath)
		assert.Equal(t, "onnx", cfg.ModelType)

		migratedData, err := os.ReadFile(configPath)
		require.NoError(t, err)
		assert.True(t, json.Valid(migratedData), "config should be migrated to JSON")
	}
}

func TestSave_WritesJSON(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		DatabasePath:   "custom.db",
		ModelType:      "onnx",
		configPath:     filepath.Join(tmpDir, "config.json"),
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
	assert.Equal(t, "onnx", loaded.ModelType)
}

func TestSave_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	original := &Config{
		DatabasePath: "test.db",
		ModelType:    "onnx",
		configPath:   filepath.Join(tmpDir, "config.json"),
	}

	err := Save(original)
	require.NoError(t, err)

	reloaded := &Config{configPath: original.configPath}
	reloaded.loadFromFile()

	assert.Equal(t, original.DatabasePath, reloaded.DatabasePath)
	assert.Equal(t, original.ModelType, reloaded.ModelType)
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