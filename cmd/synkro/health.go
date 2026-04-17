package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/rodascaar/synkro/internal/db"
	"github.com/rodascaar/synkro/internal/embeddings"
	"github.com/rodascaar/synkro/internal/graph"
	mcpserver "github.com/rodascaar/synkro/internal/mcp"
	"github.com/rodascaar/synkro/internal/memory"
	"github.com/rodascaar/synkro/internal/pruner"
	"github.com/rodascaar/synkro/internal/session"
	"github.com/spf13/cobra"
	"go/version"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check Synkro installation health",
	Long:  "Verify that Synkro is correctly installed and configured",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔍 Synkro Health Check")
		fmt.Println(strings.Repeat("=", 50))

		allOK := true

		// Check 1: Database
		fmt.Println("\n📊 Database:")
		d, err := db.New(cfg.DatabasePath)
		if err != nil {
			fmt.Printf("  ❌ Cannot open database: %v\n", err)
			allOK = false
		} else {
			fmt.Println("  ✅ Database accessible")
			_ = d.Close()
		}

		// Check 2: Go Version
		fmt.Println("\nGo Version:")
		fmt.Printf("  %s\n", runtime.Version())
		if !version.IsValid(runtime.Version()) {
			fmt.Println("  Warning: could not parse Go version")
			allOK = false
		} else if version.Compare(runtime.Version(), "go1.24.0") < 0 {
			fmt.Println("  Warning: Go version is below minimum 1.24.0")
			allOK = false
		} else {
			fmt.Println("  OK")
		}

		// Check 3: CGO
		fmt.Println("\n🔧 CGO Support:")
		if os.Getenv("CGO_ENABLED") == "1" {
			fmt.Println("  ✅ CGO enabled")
		} else {
			fmt.Println("  ⚠️  CGO may not be enabled")
			allOK = false
		}

		// Check 4: Database Tables
		fmt.Println("\n🗃️  Database Schema:")
		d, err = db.New(cfg.DatabasePath)
		if err == nil {
			// Check if tables exist
			rows, _ := d.DB().Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
			tables := []string{}
			for rows.Next() {
				var name string
				_ = rows.Scan(&name)
				tables = append(tables, name)
			}
			_ = rows.Close()

			expectedTables := []string{"memories", "memories_fts", "memory_embeddings", "embedding_cache", "memory_relations", "sessions", "session_memories"}
			allTablesExist := true
			for _, expected := range expectedTables {
				found := false
				for _, actual := range tables {
					if actual == expected {
						found = true
						break
					}
				}
				if !found {
					fmt.Printf("  ❌ Missing table: %s\n", expected)
					allTablesExist = false
				}
			}
			if allTablesExist {
				fmt.Printf("  ✅ All expected tables present (%d tables)\n", len(tables))
			} else {
				allOK = false
			}
			_ = d.Close()
		}

		// Check 5: MCP Server Components
		fmt.Println("\n🤖 MCP Server:")
		d, err = db.New(cfg.DatabasePath)
		if err == nil {
			repo := memory.NewRepository(d.DB())

			embedMgr, err := embeddings.NewEmbeddingManager(embeddings.Config{
				ModelType: embeddings.ModelTypeTFIDF,
				DB:        d.DB(),
			})
			if err != nil {
				fmt.Printf("  ❌ Cannot create embedding manager: %v\n", err)
				allOK = false
			} else {
				fmt.Println("  ✅ Embedding manager available")
				repo.SetEmbeddingGenerator(embedMgr)
			}

			graphRepo := graph.NewRepository(d.DB())
			g := graph.NewGraph(repo, graphRepo)
			if g != nil {
				fmt.Println("  ✅ Graph initialized")
			} else {
				fmt.Println("  ❌ Graph initialization failed")
				allOK = false
			}

			sessionRepo := session.NewRepository(d.DB())
			st := session.NewSessionTracker(sessionRepo)
			if st != nil {
				fmt.Println("  ✅ Session tracker initialized")
			} else {
				fmt.Println("  ❌ Session tracker initialization failed")
				allOK = false
			}

			cp := pruner.NewContextPruner()
			if cp != nil {
				fmt.Println("  ✅ Context pruner initialized")
			} else {
				fmt.Println("  ❌ Context pruner initialization failed")
				allOK = false
			}

			mcpServer := mcpserver.NewServer(repo, g, st, cp)
			mcpServer.SetVersion(Version)
			mcpServer.SetEmbeddingType(cfg.ModelType)
			if mcpServer != nil {
				fmt.Println("  ✅ MCP server can be started")
			} else {
				fmt.Println("  ❌ MCP server initialization failed")
				allOK = false
			}

			_ = d.Close()
		}

		// Check 6: Version
		fmt.Println("\n📦 Version:")
		fmt.Printf("  Synkro v%s\n", Version)
		fmt.Printf("  Commit: %s\n", Commit)
		fmt.Printf("  Built: %s\n", BuildTime)

		// Summary
		fmt.Println("\n" + strings.Repeat("=", 50))
		if allOK {
			fmt.Println("✅ All checks passed! Synkro is healthy.")
		} else {
			fmt.Println("⚠️  Some checks failed. See details above.")
		}
	},
}
