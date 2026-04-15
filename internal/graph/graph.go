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
	relation  []*memory.MemoryRelation
}

func NewGraph(repo *memory.Repository, graphRepo *Repository) *Graph {
	g := &Graph{
		repo:      repo,
		graphRepo: graphRepo,
		relation:  make([]*memory.MemoryRelation, 0),
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
	g.relation = relations
	g.mu.Unlock()
}

func (g *Graph) AddRelation(ctx context.Context, relation *memory.MemoryRelation) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.relation = append(g.relation, relation)

	if g.graphRepo != nil {
		return g.graphRepo.Add(ctx, relation)
	}

	return nil
}

func (g *Graph) GetRelations(ctx context.Context, memoryID string) ([]*memory.MemoryRelation, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var relations []*memory.MemoryRelation

	if g.graphRepo != nil {
		dbRelations, err := g.graphRepo.Get(ctx, memoryID)
		if err == nil {
			return dbRelations, nil
		}
	}

	for _, rel := range g.relation {
		if rel.SourceID == memoryID || rel.TargetID == memoryID {
			relations = append(relations, rel)
		}
	}

	return relations, nil
}

func (g *Graph) FindPath(ctx context.Context, fromID, toID string) ([]string, error) {
	if fromID == toID {
		return []string{fromID}, nil
	}

	parent := make(map[string]string)
	visited := make(map[string]bool)
	queue := []string{fromID}
	visited[fromID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		relations, _ := g.GetRelations(ctx, current)
		for _, rel := range relations {
			next := rel.TargetID
			if visited[next] {
				continue
			}
			visited[next] = true
			parent[next] = current

			if next == toID {
				path := []string{next}
				for node := current; node != fromID; node = parent[node] {
					path = append([]string{node}, path...)
				}
				path = append([]string{fromID}, path...)
				return path, nil
			}

			queue = append(queue, next)
		}
	}

	return nil, fmt.Errorf("no path found from %s to %s", fromID, toID)
}

func (g *Graph) GetStats(ctx context.Context) (map[string]int, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	stats := make(map[string]int)

	for _, rel := range g.relation {
		stats[string(rel.Type)]++
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

	filtered := make([]*memory.MemoryRelation, 0, len(g.relation))
	for _, rel := range g.relation {
		if !(rel.SourceID == sourceID && rel.TargetID == targetID) {
			filtered = append(filtered, rel)
		}
	}
	g.relation = filtered

	return nil
}
