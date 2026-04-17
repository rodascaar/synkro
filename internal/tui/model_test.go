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

func strPtr(s string) *string { return &s }

func setupTestTUI(t *testing.T) (*memory.Repository, func()) {
	d, err := db.New(":memory:")
	require.NoError(t, err)

	repo := memory.NewRepository(d.DB())

	cleanup := func() { _ = d.Close() }
	return repo, cleanup
}

func createMemories(t *testing.T, repo *memory.Repository) {
	t.Helper()
	types := []string{"note", "decision", "task", "note"}
	for i, typ := range types {
		mem := &memory.Memory{
			Type:    typ,
			Title:   "Test Title " + string(rune('0'+i)),
			Content: "Test Content " + string(rune('0'+i)),
			Source:  strPtr("test"),
			Status:  "active",
		}
		require.NoError(t, repo.Create(context.Background(), mem))
	}

	archived := &memory.Memory{
		Type:    "note",
		Title:   "Archived Note",
		Content: "Old content",
		Source:  strPtr("test"),
		Status:  "archived",
	}
	require.NoError(t, repo.Create(context.Background(), archived))
}

func TestModel_Init(t *testing.T) {
	repo, cleanup := setupTestTUI(t)
	defer cleanup()

	model := tui.InitialModel(repo, nil)

	cmd := model.Init()
	assert.NotNil(t, cmd)
}

func TestModel_Update_Navigation(t *testing.T) {
	repo, cleanup := setupTestTUI(t)
	defer cleanup()
	createMemories(t, repo)

	model := tui.InitialModel(repo, nil)
	model.Init()

	_, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.NotNil(t, newModel)

	newModel, _ = newModel.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.NotNil(t, newModel)

	newModel, _ = newModel.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.NotNil(t, newModel)
}

func TestModel_Update_Filter(t *testing.T) {
	repo, cleanup := setupTestTUI(t)
	defer cleanup()
	createMemories(t, repo)

	model := tui.InitialModel(repo, nil)
	model.Init()
	_, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	_ = newModel.View()

	newModel, _ = newModel.Update(tea.KeyMsg{Type: tea.KeyTab})
	_ = newModel.View()

	newModel, _ = newModel.Update(tea.KeyMsg{Type: tea.KeyTab})
	_ = newModel.View()

	newModel, _ = newModel.Update(tea.KeyMsg{Type: tea.KeyTab})
	_ = newModel.View()

	newModel, _ = newModel.Update(tea.KeyMsg{Type: tea.KeyTab})
	_ = newModel.View()

	newModel, _ = newModel.Update(tea.KeyMsg{Type: tea.KeyTab})
	_ = newModel.View()
}

func TestModel_Update_Search(t *testing.T) {
	repo, cleanup := setupTestTUI(t)
	defer cleanup()
	createMemories(t, repo)

	model := tui.InitialModel(repo, nil)
	model.Init()
	_, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	assert.NotNil(t, cmd)

	newModel, _ = newModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t', 'e', 's', 't'}})
	assert.NotNil(t, newModel)

	newModel, _ = newModel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.NotNil(t, newModel)
}

func TestModel_Update_Graph(t *testing.T) {
	repo, cleanup := setupTestTUI(t)
	defer cleanup()
	createMemories(t, repo)

	model := tui.InitialModel(repo, nil)
	model.Init()
	_, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	_ = newModel.View()

	newModel, _ = newModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	assert.NotNil(t, newModel)
}

func TestModel_View(t *testing.T) {
	repo, cleanup := setupTestTUI(t)
	defer cleanup()

	model := tui.InitialModel(repo, nil)
	model.Init()
	_, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	view := model.View()
	assert.Contains(t, view, "SYNKRO")
	assert.Contains(t, view, "FILTERS")
	assert.Contains(t, view, "SHORTCUTS")
}

func TestModel_View_Empty(t *testing.T) {
	d, err := db.New(":memory:")
	require.NoError(t, err)
	defer func() { _ = d.Close() }()

	repo := memory.NewRepository(d.DB())
	model := tui.InitialModel(repo, nil)
	model.Init()
	_, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	view := model.View()
	assert.Contains(t, view, "Press 'a' to add your first memory")
}

func TestAddModel_Update(t *testing.T) {
	repo, cleanup := setupTestTUI(t)
	defer cleanup()

	parent := tui.InitialModel(repo, nil)
	parent.Init()

	addModel := tui.NewAddModel(parent, repo)

	cmd := addModel.Init()
	assert.NotNil(t, cmd)

	newModel, cmd := addModel.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.NotNil(t, newModel)
	assert.NotNil(t, cmd)

	newModel, _ = newModel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.NotNil(t, newModel)
}

func TestModel_View_WithTags(t *testing.T) {
	repo, cleanup := setupTestTUI(t)
	defer cleanup()

	mem := &memory.Memory{
		Type:    "note",
		Title:   "Tagged Memory",
		Content: "Content with tags",
		Source:  strPtr("test"),
		Status:  "active",
		Tags:    []string{"go", "testing", "tui"},
	}
	require.NoError(t, repo.Create(context.Background(), mem))

	model := tui.InitialModel(repo, nil)
	cmd := model.Init()
	assert.NotNil(t, cmd)

	loadedModel, _ := model.Update(cmd())
	_, _ = loadedModel.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	view := loadedModel.View()
	assert.Contains(t, view, "Tagged Memory")
}

func TestModel_View_RenderContent(t *testing.T) {
	repo, cleanup := setupTestTUI(t)
	defer cleanup()

	mem := &memory.Memory{
		Type:    "decision",
		Title:   "Architecture Decision",
		Content: "We decided to use SQLite for local storage because it's simple and requires no server.",
		Source:  strPtr("team-meeting"),
		Status:  "active",
	}
	require.NoError(t, repo.Create(context.Background(), mem))

	model := tui.InitialModel(repo, nil)
	cmd := model.Init()
	assert.NotNil(t, cmd)

	loadedModel, _ := model.Update(cmd())
	_, _ = loadedModel.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	view := loadedModel.View()
	assert.Contains(t, view, "Architecture Decision")
	assert.Contains(t, view, "DETAILS")
	assert.Contains(t, view, "decision")
}
