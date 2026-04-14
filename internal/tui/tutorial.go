package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tutorialStep struct {
	title   string
	content string
}

var tutorialSteps = []tutorialStep{
	{
		title:   "👋 Welcome to Synkro!",
		content: "Synkro is your intelligent memory system for LLMs.\n\nIt stores, organizes, and retrieves your important information so AI assistants can use it effectively.",
	},
	{
		title:   "📦 What is a Memory?",
		content: "Memories are like notes for AI assistants.\n\nTypes:\n  • Note: General information\n  • Decision: Technical/architectural decisions\n  • Task: Actionable items\n  • Context: Background information",
	},
	{
		title:   "🔍 How Search Works",
		content: "Synkro uses two powerful methods:\n\n1. Full-Text Search (FTS5): Exact matches\n2. Semantic Search (Embeddings): Similar meanings\n\nBoth work together for best results!",
	},
	{
		title:   "🤖 Using with LLMs",
		content: "Synkro integrates with AI assistants via MCP (Model Context Protocol).\n\nConfigure your IDE to use synkro as an MCP server.",
	},
	{
		title:   "🎯 Ready to Start!",
		content: "You're ready to use Synkro!\n\nPress any key to continue to main interface.\n\nQuick tips:\n  • Press 'a' to add your first memory\n  • Press '/' to search\n  • Press 'g' to view graph",
	},
}

type tutorialModel struct {
	step     int
	quitting bool
}

func InitialTutorialModel() tea.Model {
	return tutorialModel{
		step:     0,
		quitting: false,
	}
}

func (m tutorialModel) Init() tea.Cmd {
	return nil
}

func (m tutorialModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyEnter, tea.KeySpace:
			if m.step < len(tutorialSteps)-1 {
				m.step++
			} else {
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m tutorialModel) View() string {
	step := tutorialSteps[m.step]

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00e5cc")).
		Bold(true).
		Padding(0, 1)

	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Padding(1, 2)

	progressStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Faint(true)

	progress := strings.Builder{}
	for i := 0; i < len(tutorialSteps); i++ {
		if i == m.step {
			progress.WriteString("● ")
		} else if i < m.step {
			progress.WriteString("✓ ")
		} else {
			progress.WriteString("○ ")
		}
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Padding(1, 0)

	help := helpStyle.Render("Press Enter to continue • Esc to skip tutorial")

	sections := []string{
		titleStyle.Render(step.title),
		"",
		contentStyle.Render(step.content),
		"",
		progressStyle.Render("Progress: " + progress.String()),
		"",
		help,
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
