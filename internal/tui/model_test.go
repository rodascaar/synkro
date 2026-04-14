package tui_test

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rodascaar/synkro/internal/db"
	"github.com/rodascaar/synkro/internal/memory"
	"github.com/rodascaar/synkro/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestTUI(t *testing.T) (*db.Database, *memory.Repository, func()) {
	d, err := db.New(":memory:")
	require.NoError(t, err)

	repo := memory.NewRepository(d.DB())

	for i := 0; i < 3; i++ {
		mem := &memory.Memory{
			Type:    "note",
			Title:   "Test Title " + string(rune('0'+i)),
			Content: "Test Content " + string(rune('0'+i)),
			Source:  "test",
			Status:  "active",
		}
		require.NoError(t, repo.Create(context.Background(), mem))
	}

	cleanup := func() { d.Close() }
	return d, repo, cleanup
}

func TestModel_Init(t *testing.T) {
	_, repo, cleanup := setupTestTUI(t)
	defer cleanup()

	model := tui.InitialModel(repo, nil)

	cmd := model.Init()
	assert.NotNil(t, cmd)
}

func TestModel_Update_Keys(t *testing.T) {
	_, repo, cleanup := setupTestTUI(t)
	defer cleanup()

	model := tui.InitialModel(repo, nil)
	model.Init()

	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.NotNil(t, newModel)
	assert.Nil(t, cmd)

	newModel, cmd = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.NotNil(t, newModel)
	assert.Nil(t, cmd)

	newModel, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	assert.NotNil(t, newModel)
	assert.Nil(t, cmd)
}

func TestAddModel_Update(t *testing.T) {
	_, repo, cleanup := setupTestTUI(t)
	defer cleanup()

	parent := tui.InitialModel(repo, nil)
	parent.Init()

	addModel := tui.NewAddModel(parent, repo)

	cmd := addModel.Init()
	assert.NotNil(t, cmd)

	newModel, cmd := addModel.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.NotNil(t, newModel)
	assert.NotNil(t, cmd)
}
