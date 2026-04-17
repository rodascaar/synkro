package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rodascaar/synkro/internal/memory"
)

type addModel struct {
	repo         *memory.Repository
	parent       *model
	fields       [4]textinput.Model
	currentField int
	errMsg       string
	ctx          context.Context
}

func NewAddModel(parent *model, repo *memory.Repository) *addModel {
	typeInput := textinput.New()
	typeInput.Placeholder = "Type (note/decision/task/context)"
	typeInput.SetValue("note")
	typeInput.CharLimit = 20

	titleInput := textinput.New()
	titleInput.Placeholder = "Title (required)"
	titleInput.CharLimit = 100

	contentInput := textinput.New()
	contentInput.Placeholder = "Content (detailed information)"
	contentInput.CharLimit = 2000

	tagsInput := textinput.New()
	tagsInput.Placeholder = "Tags (comma separated, optional)"
	tagsInput.CharLimit = 100

	return &addModel{
		repo:         repo,
		parent:       parent,
		fields:       [4]textinput.Model{typeInput, titleInput, contentInput, tagsInput},
		currentField: 1,
		ctx:          context.Background(),
	}
}

func (m *addModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *addModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.currentField < 3 {
				m.currentField++
				return m, textinput.Blink
			} else {
				tags := []string{}
				if m.fields[3].Value() != "" {
					tags = strings.Split(m.fields[3].Value(), ",")
					for i := range tags {
						tags[i] = strings.TrimSpace(tags[i])
					}
				}

				mem := &memory.Memory{
					Type:    m.fields[0].Value(),
					Title:   m.fields[1].Value(),
					Content: m.fields[2].Value(),
					Source:  sourcePtr("tui"),
					Status:  "active",
					Tags:    tags,
				}

				if mem.Title == "" {
					m.errMsg = "Error: Title is required"
					return m, nil
				}

				if err := m.repo.Create(m.ctx, mem); err != nil {
					m.errMsg = fmt.Sprintf("Error: %v", err)
					return m, nil
				}

				return m.parent, m.parent.loadMemories()
			}
		case tea.KeyEscape:
			return m.parent, nil
		case tea.KeyTab:
			if m.currentField < 3 {
				m.currentField++
				return m, textinput.Blink
			}
		case tea.KeyShiftTab:
			if m.currentField > 0 {
				m.currentField--
				return m, textinput.Blink
			}
		}
	}

	var cmd tea.Cmd
	m.fields[m.currentField], cmd = m.fields[m.currentField].Update(msg)
	return m, cmd
}

func (m *addModel) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Render("ADD NEW MEMORY"),
		"",
		m.renderField("Type", m.fields[0], m.currentField == 0),
		m.renderField("Title", m.fields[1], m.currentField == 1),
		m.renderField("Content", m.fields[2], m.currentField == 2),
		m.renderField("Tags", m.fields[3], m.currentField == 3),
		"",
		func() string {
			if m.errMsg != "" {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("#ff5555")).
					Render(m.errMsg)
			}
			return emptyStyle.Render("Enter: Next/Save • Esc: Cancel • Tab: Next field")
		}(),
	)
}

func (m *addModel) renderField(label string, input textinput.Model, focused bool) string {
	style := normalStyle
	if focused {
		style = selectedStyle
	}
	return lipgloss.JoinHorizontal(lipgloss.Top,
		style.Render(fmt.Sprintf("%-12s", label+":")),
		" ",
		input.View(),
	)
}

func sourcePtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
