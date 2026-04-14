package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rodascaar/synkro/internal/graph"
	"github.com/rodascaar/synkro/internal/memory"
)

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00e5cc")).
			Bold(true).
			Padding(0, 1)

	sidebarStyle = lipgloss.NewStyle().
			Width(25).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444"))

	contentStyle = lipgloss.NewStyle().
			Width(55).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444"))

	detailStyle = lipgloss.NewStyle().
			Width(40).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888")).
			Padding(0, 1)

	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666")).
			Italic(true).
			Margin(1, 0)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00e5cc")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#fff"))

	relStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00e5cc"))
)

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Escape key.Binding
	Search key.Binding
	Add    key.Binding
	Quit   key.Binding
	Graph  key.Binding
	List   key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up, k.Down, k.Enter, k.Search, k.Add, k.Quit,
	}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.Enter, k.Escape},
		{k.Search, k.Graph, k.List},
		{k.Add, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter", " "),
		key.WithHelp("enter", "view"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc", "q"),
		key.WithHelp("esc/q", "back"),
	),
	Search: key.NewBinding(
		key.WithKeys("/", "s"),
		key.WithHelp("/s", "search"),
	),
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add memory"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	Graph: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "graph"),
	),
	List: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "list"),
	),
}

type model struct {
	repo       *memory.Repository
	graph      *graph.Graph
	graphView  *GraphView
	memories   []*memory.Memory
	selected   int
	sidebarSel int
	filterType string
	showGraph  bool
	searching  bool
	searchBox  textinput.Model
	help       help.Model
	width      int
	height     int
	ctx        context.Context
	err        error
	adding     bool
}

func InitialModel(repo *memory.Repository, g *graph.Graph) *model {
	m := &model{
		repo:       repo,
		graph:      g,
		graphView:  NewGraphView(),
		memories:   []*memory.Memory{},
		selected:   0,
		sidebarSel: 0,
		filterType: "",
		showGraph:  false,
		searching:  false,
	}
	return m
}

func (m *model) Init() tea.Cmd {
	m.searchBox = textinput.New()
	m.searchBox.Placeholder = "Search memories..."
	m.searchBox.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00e5cc"))
	m.searchBox.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#fff"))
	m.searchBox.CharLimit = 100
	m.searchBox.Width = 30

	m.help = help.New()
	m.ctx = context.Background()
	return m.loadMemories()
}

func (m *model) loadMemories() tea.Cmd {
	return func() tea.Msg {
		memories, err := m.repo.Search(m.ctx, "", memory.MemoryFilter{Limit: 100})
		if err != nil {
			return errMsg{err}
		}
		return memoriesLoadedMsg{memories}
	}
}

func (m *model) filterMemories() []*memory.Memory {
	if m.filterType == "" {
		return m.memories
	}

	filtered := make([]*memory.Memory, 0)
	for _, mem := range m.memories {
		if mem.Type == m.filterType {
			filtered = append(filtered, mem)
		}
	}
	return filtered
}

func (m *model) searchMemories(query string) []*memory.Memory {
	if query == "" {
		return m.filterMemories()
	}

	base := m.filterMemories()
	filtered := make([]*memory.Memory, 0)
	queryLower := strings.ToLower(query)

	for _, mem := range base {
		titleMatch := strings.Contains(strings.ToLower(mem.Title), queryLower)
		contentMatch := strings.Contains(strings.ToLower(mem.Content), queryLower)
		if titleMatch || contentMatch {
			filtered = append(filtered, mem)
		}
	}
	return filtered
}

type memoriesLoadedMsg struct {
	memories []*memory.Memory
}

type errMsg struct {
	err error
}

func (e errMsg) Error() string { return e.err.Error() }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searching {
			switch msg.Type {
			case tea.KeyEnter, tea.KeyEscape:
				m.searching = false
				m.selected = 0
				return m, nil
			default:
				var searchCmd tea.Cmd
				m.searchBox, searchCmd = m.searchBox.Update(msg)
				m.selected = 0
				return m, searchCmd
			}
		}

		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Up):
			if m.selected > 0 && !m.showGraph {
				m.selected--
			}
			return m, nil

		case key.Matches(msg, keys.Down):
			if !m.showGraph {
				displayed := m.getDisplayedMemories()
				if m.selected < len(displayed)-1 {
					m.selected++
				}
			}
			return m, nil

		case key.Matches(msg, keys.Enter):
			if len(m.getDisplayedMemories()) > 0 && !m.showGraph {
				m.showGraph = !m.showGraph
				m.selected = 0
			}
			return m, nil

		case key.Matches(msg, keys.Escape):
			if m.showGraph {
				m.showGraph = false
			} else {
				return m, tea.Quit
			}
			return m, nil

		case key.Matches(msg, keys.Search):
			m.searching = true
			m.showGraph = false
			m.searchBox.Focus()
			m.searchBox.SetValue("")
			m.selected = 0
			return m, textinput.Blink

		case key.Matches(msg, keys.Add):
			m.adding = true
			return NewAddModel(m, m.repo), nil

		case key.Matches(msg, keys.Graph):
			if len(m.getDisplayedMemories()) > 0 {
				m.showGraph = true
			}
			return m, nil

		case key.Matches(msg, keys.List):
			if m.showGraph {
				m.showGraph = false
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.searchBox.Width = m.width/3 - 4
		return m, nil

	case memoriesLoadedMsg:
		m.memories = msg.memories
		m.selected = 0
		return m, nil

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, cmd
}

func (m *model) getDisplayedMemories() []*memory.Memory {
	query := m.searchBox.Value()
	return m.searchMemories(query)
}

func (m *model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		m.renderHeader(),
		lipgloss.JoinHorizontal(lipgloss.Top,
			sidebarStyle.Render(m.renderSidebar()),
			contentStyle.Render(m.renderContent()),
			detailStyle.Render(m.renderDetail()),
		),
		m.renderFooter(),
	)
}

func (m *model) renderHeader() string {
	return headerStyle.Render("SYNKRO")
}

func (m *model) renderSidebar() string {
	filters := []struct {
		name string
		typ  string
	}{
		{"All", ""},
		{"Decisions", "decision"},
		{"Tasks", "task"},
		{"Notes", "note"},
		{"Archive", "archived"},
	}

	var content strings.Builder
	content.WriteString("FILTERS\n\n")

	for i, filter := range filters {
		prefix := "  "
		isSelected := (i == m.sidebarSel)
		style := normalStyle
		if isSelected {
			style = selectedStyle
			prefix = "● "
		}

		content.WriteString(style.Render(prefix + filter.name))
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString("SHORTCUTS\n\n")
	content.WriteString("/  : Search\n")
	if m.showGraph {
		content.WriteString("l  : List\n")
	} else {
		content.WriteString("g  : Graph\n")
	}
	content.WriteString("q  : Quit\n")

	return content.String()
}

func (m *model) renderContent() string {
	if m.showGraph {
		return m.renderGraphView()
	}

	if m.searching {
		return "SEARCH\n\n" + m.searchBox.View()
	}

	displayed := m.getDisplayedMemories()

	if len(displayed) == 0 {
		return emptyStyle.Render(
			"📥 Press 'a' to add your first memory\n" +
				"🔍 Press '/' to search",
		)
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintf("MEMORIES (%d)", len(displayed)))
	query := m.searchBox.Value()
	if query != "" {
		content.WriteString(fmt.Sprintf(" [search: %s]", query))
	}
	if m.filterType != "" {
		content.WriteString(fmt.Sprintf(" [filter: %s]", m.filterType))
	}
	content.WriteString("\n\n")

	start := 0
	if m.selected > m.height/3 {
		start = m.selected - m.height/3
	}

	end := start + m.height/3
	if end > len(displayed) {
		end = len(displayed)
	}

	for i := start; i < end; i++ {
		mem := displayed[i]
		style := normalStyle
		if i == m.selected {
			style = selectedStyle
		}

		bullet := "○ "
		if i == m.selected {
			bullet = "● "
		}

		content.WriteString(style.Render(bullet))
		content.WriteString(style.Render(fmt.Sprintf("[%s] %s", mem.Type, mem.Title)))
		if i < end-1 {
			content.WriteString("\n")
		}
	}

	return content.String()
}

func (m *model) renderGraphView() string {
	displayed := m.getDisplayedMemories()

	if len(displayed) == 0 {
		return emptyStyle.Render("No memories to display in graph")
	}

	m.graphView = NewGraphView()

	for _, mem := range displayed {
		level := 0
		if mem.Type == "decision" {
			level = 0
		} else if mem.Type == "task" {
			level = 1
		} else {
			level = 2
		}

		node := m.graphView.AddMemory(mem, level)

		relations, err := m.graph.GetRelations(m.ctx, mem.ID)
		if err == nil {
			for _, rel := range relations {
				targetMem, err := m.repo.Get(m.ctx, rel.TargetID)
				if err != nil {
					continue
				}
				targetNode := m.graphView.AddMemory(targetMem, level+1)
				m.graphView.AddRelation(node, targetNode, rel.Type, rel.Strength)
			}
		}
	}

	availableWidth := m.width - 30
	availableHeight := m.height - 5
	m.graphView.CalculateLayout(availableWidth, availableHeight)

	return m.graphView.Render()
}

func (m *model) renderDetail() string {
	displayed := m.getDisplayedMemories()

	if len(displayed) == 0 {
		return emptyStyle.Render(
			"No memories to display",
		)
	}

	mem := displayed[m.selected]

	var content strings.Builder
	content.WriteString("DETAILS\n\n")
	content.WriteString(fmt.Sprintf("Type: %s\n", mem.Type))
	content.WriteString(fmt.Sprintf("ID: %s\n", mem.ID))
	content.WriteString(fmt.Sprintf("Source: %s\n", mem.Source))
	content.WriteString(fmt.Sprintf("Status: %s\n", mem.Status))
	if len(mem.Tags) > 0 {
		content.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(mem.Tags, ", ")))
	}
	content.WriteString("\n")
	content.WriteString("CONTENT:\n")
	content.WriteString(lipgloss.NewStyle().Width(36).Render(mem.Content))

	if m.showGraph {
		content.WriteString("\n\n")
		content.WriteString("RELATIONS\n\n")
		relations, err := m.graph.GetRelations(m.ctx, mem.ID)
		if err == nil && len(relations) > 0 {
			for _, rel := range relations {
				targetMem, err := m.repo.Get(m.ctx, rel.TargetID)
				if err != nil {
					continue
				}

				var relName string
				switch rel.Type {
				case "related_to":
					relName = "Related to"
				case "part_of":
					relName = "Part of"
				case "extends":
					relName = "Extends"
				case "conflicts_with":
					relName = "Conflicts with"
				case "depends_on":
					relName = "Depends on"
				default:
					relName = rel.Type
				}

				arrow := "->"
				if rel.Type == "part_of" {
					arrow = "<-"
				}

				content.WriteString(relStyle.Render(fmt.Sprintf("%s %s\n", arrow, relName)))
				content.WriteString(fmt.Sprintf("  %s\n", targetMem.Title))
				content.WriteString(fmt.Sprintf("  Strength: %.1f\n", rel.Strength))
				content.WriteString("\n")
			}
		} else {
			content.WriteString("No relations found\n")
			content.WriteString("\n")
			content.WriteString("Relations connect memories\n")
			content.WriteString("that are related, similar,\n")
			content.WriteString("or depend on each other.")
		}
	}

	return content.String()
}

func (m *model) renderFooter() string {
	query := m.searchBox.Value()
	var statusInfo strings.Builder
	statusInfo.WriteString("DB: memory.db")
	if query != "" {
		statusInfo.WriteString(" | Searching")
	}
	if m.filterType != "" {
		statusInfo.WriteString(" | Filtered")
	}

	return lipgloss.JoinHorizontal(lipgloss.Bottom,
		statusStyle.Render(statusInfo.String()),
		lipgloss.NewStyle().Padding(0, 1).Render(m.help.View(keys)),
	)
}
