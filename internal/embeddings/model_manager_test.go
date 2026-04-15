package embeddings

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelManager_NewWithNilConfig(t *testing.T) {
	mm := NewModelManager(nil)
	assert.NotNil(t, mm)
	models := mm.ListModels()
	assert.Len(t, models, 3)
}

func TestModelManager_NewWithConfig(t *testing.T) {
	tmpDir := t.TempDir()
	mm := NewModelManager(&ManagerConfig{
		DownloadDir:  tmpDir,
		CacheDir:     filepath.Join(tmpDir, "cache"),
		MaxModels:    5,
		AutoDownload: false,
	})
	assert.NotNil(t, mm)
}

func TestModelManager_ListModels(t *testing.T) {
	mm := NewModelManager(&ManagerConfig{DownloadDir: t.TempDir()})
	models := mm.ListModels()
	assert.Len(t, models, 3)

	names := make(map[string]bool)
	for _, m := range models {
		names[m.Name] = true
	}
	assert.True(t, names["all-MiniLM-L6-v2"])
	assert.True(t, names["paraphrase-multilingual-MiniLM-L12-v2"])
	assert.True(t, names["stsb-roberta-base-v2"])
}

func TestModelManager_GetModel(t *testing.T) {
	mm := NewModelManager(&ManagerConfig{DownloadDir: t.TempDir()})

	model, err := mm.GetModel("all-MiniLM-L6-v2")
	require.NoError(t, err)
	assert.Equal(t, "all-MiniLM-L6-v2", model.Name)
	assert.Equal(t, 384, model.Dimension)
	assert.Equal(t, "en", model.Language)
	assert.NotEmpty(t, model.Description)
}

func TestModelManager_GetModel_NotFound(t *testing.T) {
	mm := NewModelManager(&ManagerConfig{DownloadDir: t.TempDir()})

	_, err := mm.GetModel("nonexistent-model")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestModelManager_GetPreferredModel(t *testing.T) {
	tmpDir := t.TempDir()
	mm := NewModelManager(&ManagerConfig{
		DownloadDir:    tmpDir,
		PreferredModel: "all-MiniLM-L6-v2",
	})

	model, err := mm.GetPreferredModel()
	require.NoError(t, err)
	assert.Equal(t, "all-MiniLM-L6-v2", model.Name)
}

func TestModelManager_GetPreferredModel_NotConfigured(t *testing.T) {
	tmpDir := t.TempDir()
	mm := NewModelManager(&ManagerConfig{
		DownloadDir:    tmpDir,
		PreferredModel: "",
	})

	_, err := mm.GetPreferredModel()
	assert.Error(t, err)
}

func TestModelManager_DeleteModel_NotDownloaded(t *testing.T) {
	tmpDir := t.TempDir()
	mm := NewModelManager(&ManagerConfig{
		DownloadDir:    tmpDir,
		PreferredModel: "all-MiniLM-L6-v2",
	})

	err := mm.DeleteModel("all-MiniLM-L6-v2")
	assert.NoError(t, err)
}

func TestModelManager_DeleteModel_Downloaded(t *testing.T) {
	tmpDir := t.TempDir()
	modelDir := filepath.Join(tmpDir, "all-MiniLM-L6-v2")
	require.NoError(t, os.MkdirAll(modelDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(modelDir, "all-MiniLM-L6-v2.onnx"), []byte("fake model"), 0644))

	mm := NewModelManager(&ManagerConfig{
		DownloadDir:    tmpDir,
		PreferredModel: "all-MiniLM-L6-v2",
	})

	model, err := mm.GetModel("all-MiniLM-L6-v2")
	require.NoError(t, err)
	assert.True(t, model.Downloaded)

	err = mm.DeleteModel("all-MiniLM-L6-v2")
	require.NoError(t, err)

	model, err = mm.GetModel("all-MiniLM-L6-v2")
	require.NoError(t, err)
	assert.False(t, model.Downloaded)
}

func TestModelManager_ValidateModel_NotDownloaded(t *testing.T) {
	tmpDir := t.TempDir()
	mm := NewModelManager(&ManagerConfig{DownloadDir: tmpDir})

	err := mm.ValidateModel("all-MiniLM-L6-v2")
	assert.Error(t, err)
}

func TestModelManager_ValidateModel_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	mm := NewModelManager(&ManagerConfig{DownloadDir: tmpDir})

	err := mm.ValidateModel("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestModelInfo_Fields(t *testing.T) {
	mm := NewModelManager(&ManagerConfig{DownloadDir: t.TempDir()})

	model, err := mm.GetModel("stsb-roberta-base-v2")
	require.NoError(t, err)
	assert.Equal(t, 768, model.Dimension)
	assert.Equal(t, "en", model.Language)
	assert.NotEmpty(t, model.License)
	assert.NotEmpty(t, model.Params)
	assert.Greater(t, model.MaxSeqLen, 0)
	assert.NotEmpty(t, model.Benchmarks)
}

func TestFileExists(t *testing.T) {
	assert.True(t, fileExists(os.TempDir()))
	assert.False(t, fileExists(filepath.Join(os.TempDir(), "nonexistent_file_12345")))
}

func TestFindONNXRuntimePath(t *testing.T) {
	path := findONNXRuntimePath()
	if path != "" {
		assert.FileExists(t, path)
	}
}
