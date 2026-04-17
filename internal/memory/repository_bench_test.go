package memory

import (
	"context"
	"fmt"
	"testing"

	"github.com/rodascaar/synkro/internal/db"
)

func BenchmarkRepository_Search(b *testing.B) {
	d, err := db.New(":memory:")
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = d.Close() }()

	repo := NewRepository(d.DB())

	for i := 0; i < 1000; i++ {
		mem := &Memory{
			Type:    "note",
			Title:   fmt.Sprintf("Title %d", i),
			Content: fmt.Sprintf("Content %d with some more text", i),
			Source:  sourcePtr("test"),
			Status:  "active",
		}
		_ = repo.Create(context.Background(), mem)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.Search(context.Background(), "test query", MemoryFilter{})
	}
}

func BenchmarkRepository_HybridSearch(b *testing.B) {
	d, err := db.New(":memory:")
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = d.Close() }()

	repo := NewRepository(d.DB())

	for i := 0; i < 100; i++ {
		mem := &Memory{
			Type:    "note",
			Title:   fmt.Sprintf("Title %d", i),
			Content: fmt.Sprintf("Content %d", i),
			Source:  sourcePtr("test"),
			Status:  "active",
		}
		_ = repo.Create(context.Background(), mem)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.HybridSearch(context.Background(), "test query", 10, HybridSearchFilter{})
	}
}
