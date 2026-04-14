package tui

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/rodascaar/synkro/internal/graph"
	"github.com/rodascaar/synkro/internal/memory"
)

type TUI struct {
	repo     *memory.Repository
	graph    *graph.Graph
	memories []*memory.Memory
	selected int
	viewMode string
	running  bool
}

func NewTUI(repo *memory.Repository, g *graph.Graph) *TUI {
	return &TUI{
		repo:     repo,
		graph:    g,
		memories: []*memory.Memory{},
		selected: 0,
		viewMode: "list",
		running:  true,
	}
}

func (t *TUI) Run() error {
	if err := t.loadMemories(); err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)

	for t.running {
		t.render()

		fmt.Print("Command (l/g/?/q/↑/↓): ")

		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		input = strings.TrimSpace(input)

		if err := t.handleInput(input); err != nil {
			if err.Error() == "quit" {
				return nil
			}
		}
	}

	t.running = false
	fmt.Println("\nGoodbye!")
	return nil
}

func (t *TUI) loadMemories() error {
	ctx := context.Background()
	memories, err := t.repo.Search(ctx, "", memory.MemoryFilter{Limit: 50})
	if err != nil {
		return err
	}
	t.memories = memories
	return nil
}

func (t *TUI) render() {
	switch t.viewMode {
	case "list":
		t.renderList()
	case "graph":
		t.renderGraph()
	default:
		t.renderList()
	}
}

func (t *TUI) renderList() {
	fmt.Println("\n=== Synkro TUI - Memory List ===")

	if len(t.memories) == 0 {
		fmt.Println("  No memories found.")
		t.renderHelp()
		return
	}

	fmt.Printf("Total: %d memories\n\n", len(t.memories))

	for i, mem := range t.memories {
		if i >= 20 {
			break
		}

		prefix := "  "
		if i == t.selected {
			prefix = "> "
		}

		fmt.Printf("%s%s [%s] %s\n",
			prefix,
			mem.Type,
			mem.ID,
			mem.Title)

		if i == t.selected {
			fmt.Println("--- Details ---")

			relations, err := t.graph.GetRelations(context.Background(), mem.ID)
			if err == nil && len(relations) > 0 {
				fmt.Println("Relations:")
				for _, rel := range relations {
					targetMem, err := t.repo.Get(context.Background(), rel.TargetID)
					if err != nil {
						continue
					}
					arrow := "->"
					if rel.Type == "part_of" {
						arrow = "<-"
					}
					fmt.Printf("  %s %s %s (%.1f)\n", arrow, rel.Type, targetMem.Title, rel.Strength)
				}
			}
		}
	}

	t.renderHelp()
}

func (t *TUI) renderGraph() {
	if t.selected >= len(t.memories) {
		return
	}

	mem := t.memories[t.selected]

	fmt.Println("\n=== Synkro TUI - Relation Graph ===")
	fmt.Println("\nMemory:")
	fmt.Printf("  %s\n", mem.Title)
	fmt.Printf("  %s\n", mem.ID)

	fmt.Println("\n=== Relations ===")

	relations, err := t.graph.GetRelations(context.Background(), mem.ID)
	if err != nil || len(relations) == 0 {
		fmt.Println("  No relations for this memory.")
		return
	}

	fmt.Printf("Total: %d relations\n\n", len(relations))

	targetRelations := make(map[string][]*memory.MemoryRelation)
	for _, rel := range relations {
		targetRelations[rel.TargetID] = append(targetRelations[rel.TargetID], rel)
	}

	for targetID, rels := range targetRelations {
		targetMem, err := t.repo.Get(context.Background(), targetID)
		if err != nil {
			continue
		}

		fmt.Printf("%s\n", targetMem.Title)
		fmt.Printf("  %s\n", targetID)

		for _, rel := range rels {
			relType := strings.ToUpper(rel.Type)
			switch relType {
			case "DEPENDS_ON":
				fmt.Printf("    Depends on:")
			case "EXTENDS":
				fmt.Printf("    Extends:")
			case "CONFLICTS_WITH":
				fmt.Printf("    Conflicts with:")
			case "EXAMPLE_OF":
				fmt.Printf("    Is example of:")
			case "PART_OF":
				fmt.Printf("    Part of:")
			case "RELATED_TO":
				fmt.Printf("    Related to:")
			}

			fmt.Printf("      Strength: %.1f\n", rel.Strength)
			fmt.Printf("      %s\n", mem.Title)
			fmt.Printf("      (%s)\n\n", rel.Type)
		}
	}

	t.renderHelp()
}

func (t *TUI) renderHelp() {
	fmt.Println("\n--- Commands ---")
	fmt.Println("  l     - List memories")
	fmt.Println("  g     - View relation graph")
	fmt.Println("  ↑     - Move up")
	fmt.Println("  ↓     - Move down")
	fmt.Println("  Enter - View details/graph")
	fmt.Println("  ?     - Show help")
	fmt.Println("  q     - Quit")
}

func (t *TUI) handleInput(input string) error {
	if len(input) == 0 {
		return nil
	}

	switch input {
	case "q", "quit":
		t.running = false
		return fmt.Errorf("quit")
	case "l", "list":
		t.viewMode = "list"
		t.selected = 0
	case "g", "graph":
		if len(t.memories) > 0 {
			t.viewMode = "graph"
		}
	case "h", "help", "?":
	case "up", "arrowup", "k", "↑":
		if t.selected > 0 {
			t.selected--
		}
	case "down", "arrowdown", "j", "↓":
		if t.selected < len(t.memories)-1 {
			t.selected++
		}
	case "enter":
		if len(t.memories) > 0 {
			t.viewMode = "graph"
		}
	case "esc":
		t.viewMode = "list"
	default:
		fmt.Printf("  Unknown command: %s\n", input)
	}

	return nil
}
