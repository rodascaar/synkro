package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rodascaar/synkro/internal/memory"
)

type GraphNode struct {
	Memory *memory.Memory
	X      int
	Y      int
	Level  int
}

type GraphEdge struct {
	From     *GraphNode
	To       *GraphNode
	Type     string
	Strength float64
}

type GraphView struct {
	Nodes map[string]*GraphNode
	Edges []*GraphEdge
}

func NewGraphView() *GraphView {
	return &GraphView{
		Nodes: make(map[string]*GraphNode),
		Edges: make([]*GraphEdge, 0),
	}
}

func (gv *GraphView) AddMemory(mem *memory.Memory, level int) *GraphNode {
	node := &GraphNode{
		Memory: mem,
		X:      0,
		Y:      0,
		Level:  level,
	}
	gv.Nodes[mem.ID] = node
	return node
}

func (gv *GraphView) AddRelation(source, target *GraphNode, relType string, strength float64) {
	edge := &GraphEdge{
		From:     source,
		To:       target,
		Type:     relType,
		Strength: strength,
	}
	gv.Edges = append(gv.Edges, edge)
}

func (gv *GraphView) CalculateLayout(width, height int) {
	levels := make(map[int][]*GraphNode)
	maxLevel := 0

	for _, node := range gv.Nodes {
		if node.Level > maxLevel {
			maxLevel = node.Level
		}
		levels[node.Level] = append(levels[node.Level], node)
	}

	availableWidth := width - 10
	availableHeight := height - 10
	levelHeight := availableHeight / (maxLevel + 1)

	for level, nodes := range levels {
		y := 2 + level*levelHeight
		nodeCount := len(nodes)
		nodeWidth := availableWidth / (nodeCount + 1)

		for i, node := range nodes {
			node.X = 5 + (i+1)*nodeWidth
			node.Y = y
		}
	}
}

func (gv *GraphView) Render() string {
	if len(gv.Nodes) == 0 {
		return emptyStyle.Render("No memories to display in graph")
	}

	var builder strings.Builder
	builder.WriteString("MEMORY GRAPH\n\n")

	for _, edge := range gv.Edges {
		builder.WriteString(gv.renderEdge(edge))
	}

	builder.WriteString("\n")

	for _, node := range gv.Nodes {
		builder.WriteString(gv.renderNode(node))
	}

	return builder.String()
}

func (gv *GraphView) renderNode(node *GraphNode) string {
	style := normalStyle
	if node.Memory.Type == "decision" {
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffff00")).
			Bold(true)
	} else if node.Memory.Type == "task" {
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ff00")).
			Bold(true)
	}

	nodeStr := fmt.Sprintf("%s [%s]", node.Memory.Type[:1], truncateString(node.Memory.Title, 20))
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888")).
		Padding(0, 0, 1, 1).
		Render(lipgloss.PlaceHorizontal(node.X, lipgloss.Center, style.Render(nodeStr)))
}

func (gv *GraphView) renderEdge(edge *GraphEdge) string {
	arrow := "→"
	if edge.Type == "part_of" {
		arrow = "←"
	}

	var relColor string
	switch edge.Type {
	case "related_to":
		relColor = "#00e5cc"
	case "part_of":
		relColor = "#ff00ff"
	case "extends":
		relColor = "#00ff00"
	case "conflicts_with":
		relColor = "#ff0000"
	case "depends_on":
		relColor = "#ffff00"
	default:
		relColor = "#888"
	}

	relStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(relColor))
	edgeStr := fmt.Sprintf("%s %s (%.1f)", arrow, truncateString(edge.Type, 15), edge.Strength)

	return relStyle.Render(edgeStr)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func (gv *GraphView) GetNodeAt(x, y int) *GraphNode {
	for _, node := range gv.Nodes {
		if x >= node.X && x <= node.X+20 && y >= node.Y && y <= node.Y+3 {
			return node
		}
	}
	return nil
}
