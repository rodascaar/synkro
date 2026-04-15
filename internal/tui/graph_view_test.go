package tui_test

import (
	"testing"

	"github.com/rodascaar/synkro/internal/memory"
	"github.com/rodascaar/synkro/internal/tui"
	"github.com/stretchr/testify/assert"
)

func TestGraphView_New(t *testing.T) {
	gv := tui.NewGraphView()
	assert.NotNil(t, gv)
	assert.NotNil(t, gv.Nodes)
	assert.NotNil(t, gv.Edges)
}

func TestGraphView_AddMemory(t *testing.T) {
	gv := tui.NewGraphView()
	mem := &memory.Memory{ID: "mem-1", Type: "note", Title: "Test Note", Status: "active"}
	node := gv.AddMemory(mem, 0)
	assert.NotNil(t, node)
	assert.Equal(t, mem, node.Memory)
	assert.Equal(t, 0, node.Level)
	assert.Equal(t, node, gv.Nodes["mem-1"])
}

func TestGraphView_AddRelation(t *testing.T) {
	gv := tui.NewGraphView()
	mem1 := &memory.Memory{ID: "mem-1", Type: "note", Title: "A", Status: "active"}
	mem2 := &memory.Memory{ID: "mem-2", Type: "note", Title: "B", Status: "active"}
	n1 := gv.AddMemory(mem1, 0)
	n2 := gv.AddMemory(mem2, 1)
	gv.AddRelation(n1, n2, "related_to", 0.8)
	assert.Len(t, gv.Edges, 1)
	assert.Equal(t, "related_to", gv.Edges[0].Type)
	assert.Equal(t, 0.8, gv.Edges[0].Strength)
}

func TestGraphView_CalculateLayout(t *testing.T) {
	gv := tui.NewGraphView()
	for i := 0; i < 3; i++ {
		mem := &memory.Memory{ID: "mem-" + string(rune('0'+i)), Type: "note", Title: "Node", Status: "active"}
		gv.AddMemory(mem, 0)
	}
	gv.CalculateLayout(80, 24)
	for _, node := range gv.Nodes {
		assert.Greater(t, node.X, 0)
		assert.Greater(t, node.Y, 0)
	}
}

func TestGraphView_CalculateLayout_MultipleLevels(t *testing.T) {
	gv := tui.NewGraphView()
	gv.AddMemory(&memory.Memory{ID: "a", Type: "note", Title: "A", Status: "active"}, 0)
	gv.AddMemory(&memory.Memory{ID: "b", Type: "note", Title: "B", Status: "active"}, 1)
	gv.AddMemory(&memory.Memory{ID: "c", Type: "note", Title: "C", Status: "active"}, 1)
	gv.CalculateLayout(100, 50)
	nodeA := gv.Nodes["a"]
	nodeB := gv.Nodes["b"]
	nodeC := gv.Nodes["c"]
	assert.Less(t, nodeA.Y, nodeB.Y)
	assert.Equal(t, nodeB.Level, nodeC.Level)
}

func TestGraphView_Render_Empty(t *testing.T) {
	gv := tui.NewGraphView()
	rendered := gv.Render()
	assert.Contains(t, rendered, "No memories to display")
}

func TestGraphView_Render_WithNodes(t *testing.T) {
	gv := tui.NewGraphView()
	gv.AddMemory(&memory.Memory{ID: "a", Type: "note", Title: "Alpha", Status: "active"}, 0)
	gv.AddMemory(&memory.Memory{ID: "b", Type: "decision", Title: "Beta", Status: "active"}, 1)
	n1 := gv.Nodes["a"]
	n2 := gv.Nodes["b"]
	gv.AddRelation(n1, n2, "related_to", 0.9)
	rendered := gv.Render()
	assert.Contains(t, rendered, "MEMORY GRAPH")
	assert.Contains(t, rendered, "Alpha")
	assert.Contains(t, rendered, "Beta")
}

func TestGraphView_GetNodeAt(t *testing.T) {
	gv := tui.NewGraphView()
	mem := &memory.Memory{ID: "a", Type: "note", Title: "Alpha", Status: "active"}
	gv.AddMemory(mem, 0)
	gv.CalculateLayout(80, 24)
	node := gv.Nodes["a"]
	found := gv.GetNodeAt(node.X, node.Y)
	assert.NotNil(t, found)
	assert.Equal(t, "a", found.Memory.ID)
}

func TestGraphView_GetNodeAt_NotFound(t *testing.T) {
	gv := tui.NewGraphView()
	mem := &memory.Memory{ID: "a", Type: "note", Title: "Alpha", Status: "active"}
	gv.AddMemory(mem, 0)
	gv.CalculateLayout(80, 24)
	found := gv.GetNodeAt(999, 999)
	assert.Nil(t, found)
}

func TestGraphView_TruncateStringViaRender(t *testing.T) {
	gv := tui.NewGraphView()
	gv.AddMemory(&memory.Memory{ID: "a", Type: "note", Title: "This is a very long title that should be truncated", Status: "active"}, 0)
	gv.CalculateLayout(80, 24)
	rendered := gv.Render()
	assert.Contains(t, rendered, "...")
}
