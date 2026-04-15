package tui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rodascaar/synkro/internal/tui"
	"github.com/stretchr/testify/assert"
)

func TestTutorialModel_Init(t *testing.T) {
	model := tui.InitialTutorialModel()
	cmd := model.Init()
	assert.Nil(t, cmd)
}

func TestTutorialModel_View_FirstStep(t *testing.T) {
	model := tui.InitialTutorialModel()
	view := model.View()
	assert.Contains(t, view, "Welcome to Synkro!")
	assert.Contains(t, view, "intelligent memory system")
	assert.Contains(t, view, "Press Enter to continue")
}

func TestTutorialModel_View_LastStep(t *testing.T) {
	model := tui.InitialTutorialModel()
	for i := 0; i < 4; i++ {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	}
	view := model.View()
	assert.Contains(t, view, "Ready to Start!")
	assert.Contains(t, view, "Press 'a' to add your first memory")
}

func TestTutorialModel_Update_Enter(t *testing.T) {
	model := tui.InitialTutorialModel()
	view1 := model.View()
	assert.Contains(t, view1, "Welcome to Synkro!")

	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view2 := model.View()
	assert.NotContains(t, view2, "Welcome to Synkro!")
	assert.Contains(t, view2, "What is a Memory?")
}

func TestTutorialModel_Update_Space(t *testing.T) {
	model := tui.InitialTutorialModel()
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeySpace})
	view := model.View()
	assert.Contains(t, view, "What is a Memory?")
}

func TestTutorialModel_Update_Esc(t *testing.T) {
	model := tui.InitialTutorialModel()
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.NotNil(t, cmd)
}

func TestTutorialModel_Update_CtrlC(t *testing.T) {
	model := tui.InitialTutorialModel()
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	assert.NotNil(t, cmd)
}

func TestTutorialModel_Update_EnterOnLastStep(t *testing.T) {
	model := tui.InitialTutorialModel()
	for i := 0; i < 5; i++ {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	}
	view := model.View()
	assert.Contains(t, view, "Ready to Start!")
}

func TestTutorialModel_ProgressIndicators(t *testing.T) {
	model := tui.InitialTutorialModel()
	view := model.View()
	assert.Contains(t, view, "●")
	assert.Contains(t, view, "○")

	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view = model.View()
	assert.Contains(t, view, "✓")
}

func TestTutorialModel_AllStepsHaveContent(t *testing.T) {
	model := tui.InitialTutorialModel()

	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotNil(t, cmd, "5th Enter should quit (last step)")
}
