package graph

import (
	"context"
	"fmt"
	"sync"

	"github.com/rodascaar/synkro/internal/memory"
)

type Graph struct {
	mu        sync.RWMutex
	repo      *memory.Repository
	graphRepo *Repository
	outgoing  map[string][]*memory.MemoryRelation
	incoming  map[string][]*memory.MemoryRelation
}

func NewGraph(repo *memory.Repository, graphRepo *Repository) *Graph {
	g := &Graph{
		repo:      repo,
		graphRepo: graphRepo,
		outgoing:  make(map[string][]*memory.MemoryRelation),
		incoming:  make(map[string][]*memory.MemoryRelation),
	}

	if graphRepo != nil {
		g.loadFromDB(context.Background())
	}

	return g
}

func (g *Graph) loadFromDB(ctx context.Context) {
	if g.graphRepo == nil {
		return
	}

	relations, err := g.graphRepo.LoadAll(ctx)
	if err != nil {
		return
	}

	g.mu.Lock()
	for _, rel := range relations {
		g.outgoing[rel.SourceID] = append(g.outgoing[rel.SourceID], rel)
		g.incoming[rel.TargetID] = append(g.incoming[rel.TargetID], rel)
	}
	g.mu.Unlock()
}

func (g *Graph) AddRelation(ctx context.Context, relation *memory.MemoryRelation) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.outgoing[relation.SourceID] = append(g.outgoing[relation.SourceID], relation)
	g.incoming[relation.TargetID] = append(g.incoming[relation.TargetID], relation)

	if g.graphRepo != nil {
		return g.graphRepo.Add(ctx, relation)
	}

	return nil
}

func (g *Graph) GetRelations(ctx context.Context, memoryID string) ([]*memory.MemoryRelation, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.graphRepo != nil {
		dbRelations, err := g.graphRepo.Get(ctx, memoryID)
		if err == nil {
			return dbRelations, nil
		}
	}

	var relations []*memory.MemoryRelation
	relations = append(relations, g.outgoing[memoryID]...)
	relations = append(relations, g.incoming[memoryID]...)

	return relations, nil
}

type neighbor struct {
	nodeID   string
	reversed bool
}

func (g *Graph) neighbors(current string) []neighbor {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var neighbors []neighbor

	for _, rel := range g.outgoing[current] {
		neighbors = append(neighbors, neighbor{
			nodeID:   rel.TargetID,
			reversed: false,
		})
	}

	for _, rel := range g.incoming[current] {
		neighbors = append(neighbors, neighbor{
			nodeID:   rel.SourceID,
			reversed: true,
		})
	}

	return neighbors
}

func (g *Graph) FindPath(ctx context.Context, fromID, toID string) ([]string, error) {
	if fromID == toID {
		return []string{fromID}, nil
	}

	parent := make(map[string]string)
	parentReversed := make(map[string]bool)
	visited := make(map[string]bool)
	queue := []string{fromID}
	visited[fromID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		neighbors := g.neighbors(current)
		for _, n := range neighbors {
			if visited[n.nodeID] {
				continue
			}
			visited[n.nodeID] = true
			parent[n.nodeID] = current
			parentReversed[n.nodeID] = n.reversed

			if n.nodeID == toID {
				path := []string{n.nodeID}
				for node := current; node != fromID; node = parent[node] {
					path = append([]string{node}, path...)
				}
				path = append([]string{fromID}, path...)
				return path, nil
			}

			queue = append(queue, n.nodeID)
		}
	}

	return nil, fmt.Errorf("no path found from %s to %s", fromID, toID)
}

func (g *Graph) GetStats(ctx context.Context) (map[string]int, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	stats := make(map[string]int)

	for _, rels := range g.outgoing {
		stats[string(memory.RelationExtends)] += len(rels)
	}
	for _, rels := range g.incoming {
		for _, rel := range rels {
			stats[string(rel.Type)]++
		}
	}

	return stats, nil
}

func (g *Graph) DeleteRelation(ctx context.Context, sourceID, targetID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.graphRepo != nil {
		if err := g.graphRepo.Delete(ctx, sourceID, targetID); err != nil {
			return err
		}
	}

	filterOutgoing := make([]*memory.MemoryRelation, 0, len(g.outgoing[sourceID]))
	for _, rel := range g.outgoing[sourceID] {
		if rel.TargetID != targetID {
			filterOutgoing = append(filterOutgoing, rel)
		}
	}
	if len(filterOutgoing) > 0 {
		g.outgoing[sourceID] = filterOutgoing
	} else {
		delete(g.outgoing, sourceID)
	}

	filterIncoming := make([]*memory.MemoryRelation, 0, len(g.incoming[targetID]))
	for _, rel := range g.incoming[targetID] {
		if rel.SourceID != sourceID {
			filterIncoming = append(filterIncoming, rel)
		}
	}
	if len(filterIncoming) > 0 {
		g.incoming[targetID] = filterIncoming
	} else {
		delete(g.incoming, targetID)
	}

	return nil
}
