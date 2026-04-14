package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rodascaar/synkro/internal/config"
	"github.com/rodascaar/synkro/internal/db"
	"github.com/rodascaar/synkro/internal/embeddings"
	"github.com/rodascaar/synkro/internal/graph"
	mcpserver "github.com/rodascaar/synkro/internal/mcp"
	"github.com/rodascaar/synkro/internal/memory"
	"github.com/rodascaar/synkro/internal/pruner"
	"github.com/rodascaar/synkro/internal/session"
	"github.com/rodascaar/synkro/internal/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "synkro",
	Short: "Synkro memory management",
	Long:  "Synkro - Motor de Contexto Inteligente para LLMs",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Verificar actualizaciones al inicio (asíncrono)
		if cfg.CheckUpdateOnStart && cmd.Name() != "update" {
			go checkForUpdatesAsync()
		}
	},
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new memory",
	Run: func(cmd *cobra.Command, args []string) {
		memType, _ := cmd.Flags().GetString("type")
		title, _ := cmd.Flags().GetString("title")
		content, _ := cmd.Flags().GetString("content")
		source, _ := cmd.Flags().GetString("source")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		if title == "" {
			fmt.Fprintln(os.Stderr, "Error: title is required")
			os.Exit(1)
		}

		d, err := db.New(cfg.DatabasePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
			os.Exit(1)
		}
		defer d.Close()

		repo := memory.NewRepository(d.DB())

		embedMgr, err := embeddings.NewEmbeddingManager(embeddings.Config{
			ModelType: embeddings.ModelTypeTFIDF,
			DB:        d.DB(),
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating embedding manager: %v\n", err)
			os.Exit(1)
		}
		repo.SetEmbeddingGenerator(embedMgr)

		mem := &memory.Memory{
			Type:    memType,
			Title:   title,
			Content: content,
			Source:  source,
			Status:  "active",
			Tags:    tags,
		}

		if err := repo.Create(context.Background(), mem); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating memory: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Memory created: %s\n", mem.ID)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List memories",
	Run: func(cmd *cobra.Command, args []string) {
		memType, _ := cmd.Flags().GetString("type")
		status, _ := cmd.Flags().GetString("status")
		limit, _ := cmd.Flags().GetInt("limit")

		d, err := db.New(cfg.DatabasePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
			os.Exit(1)
		}
		defer d.Close()

		repo := memory.NewRepository(d.DB())

		filter := memory.MemoryFilter{
			Type:   memType,
			Status: status,
			Limit:  limit,
		}

		memories, err := repo.Search(context.Background(), "", filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error searching memories: %v\n", err)
			os.Exit(1)
		}

		for _, mem := range memories {
			fmt.Printf("%s - %s\n", mem.ID, mem.Title)
		}
	},
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search memories",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "Error: query is required")
			os.Exit(1)
		}
		query := args[0]

		d, err := db.New(cfg.DatabasePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
			os.Exit(1)
		}
		defer d.Close()

		repo := memory.NewRepository(d.DB())

		filter := memory.MemoryFilter{
			Limit: 20,
		}

		memories, err := repo.Search(context.Background(), query, filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error searching memories: %v\n", err)
			os.Exit(1)
		}

		for _, mem := range memories {
			fmt.Printf("%s - %s: %s\n", mem.ID, mem.Title, mem.Content)
		}
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize database",
	Run: func(cmd *cobra.Command, args []string) {
		withModels, _ := cmd.Flags().GetBool("with-models")

		d, err := db.New(cfg.DatabasePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer d.Close()
		fmt.Printf("Database initialized at %s\n", cfg.DatabasePath)

		// Preguntar si es nuevo usuario
		fmt.Println("")
		fmt.Print("¿Eres nuevo en Synkro? ¿Quieres ver un tutorial? (y/n): ")
		var response string
		fmt.Scanln(&response)

		if strings.ToLower(strings.TrimSpace(response)) == "y" || strings.ToLower(strings.TrimSpace(response)) == "yes" {
			fmt.Println("")
			fmt.Println("Iniciando tutorial...")

			p := tea.NewProgram(tui.InitialTutorialModel(), tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Tutorial error: %v\n", err)
			}
		}

		if withModels {
			fmt.Println("Embedding models enabled (MiniLM)")
			fmt.Println("Note: For full embeddings functionality, download all-minilm-l6-v2.ggml from:")
			fmt.Println("  https://huggingface.co/ggerganov/all-minilm-l6-v2-ggml")
			fmt.Println("Place it in: internal/embeddings/models/")
		}
	},
}

var cfg *config.Config

func init() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(examplesCmd)
	rootCmd.AddCommand(healthCmd)

	addCmd.Flags().StringP("type", "t", "note", "Memory type")
	addCmd.Flags().StringP("title", "", "", "Memory title")
	addCmd.Flags().StringP("content", "c", "", "Memory content")
	addCmd.Flags().StringP("source", "s", "user", "Memory source")
	addCmd.Flags().StringSliceP("tags", "", []string{}, "Memory tags")

	listCmd.Flags().StringP("type", "t", "", "Filter by type")
	listCmd.Flags().StringP("status", "s", "", "Filter by status")
	listCmd.Flags().IntP("limit", "l", 50, "Limit results")

	initCmd.Flags().BoolP("with-models", "m", false, "Enable embedding models")
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch Synkro TUI",
	Run: func(cmd *cobra.Command, args []string) {
		d, err := db.New(cfg.DatabasePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
			os.Exit(1)
		}
		defer d.Close()

		repo := memory.NewRepository(d.DB())
		graphRepo := graph.NewRepository(d.DB())
		g := graph.NewGraph(repo, graphRepo)

		m := tui.InitialModel(repo, g)

		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
			os.Exit(1)
		}
	},
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run Synkro MCP server",
	Run: func(cmd *cobra.Command, args []string) {
		d, err := db.New(cfg.DatabasePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
			os.Exit(1)
		}
		defer d.Close()

		repo := memory.NewRepository(d.DB())

		embedMgr, err := embeddings.NewEmbeddingManager(embeddings.Config{
			ModelType: embeddings.ModelTypeTFIDF,
			DB:        d.DB(),
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating embedding manager: %v\n", err)
			os.Exit(1)
		}
		repo.SetEmbeddingGenerator(embedMgr)

		graphRepo := graph.NewRepository(d.DB())
		g := graph.NewGraph(repo, graphRepo)

		sessionRepo := session.NewRepository(d.DB())
		st := session.NewSessionTracker(sessionRepo)

		cp := pruner.NewContextPruner()

		mcpServer := mcpserver.NewServer(repo, g, st, cp)
		if err := mcpServer.Run(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
			os.Exit(1)
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func checkForUpdatesAsync() {
	// Solo verificar una vez cada 24 horas (86400 segundos)
	now := int(time.Now().Unix())
	if now-cfg.LastUpdateCheck < 86400 {
		return
	}

	latest, err := checkLatestRelease()
	if err != nil {
		return
	}

	currentVersion := fmt.Sprintf("v%s", Version)
	if latest.TagName != currentVersion {
		fmt.Printf("\n🔄 Update available: %s (current: %s)\n", latest.TagName, Version)
		fmt.Printf("📦 Download: %s\n\n", latest.HTMLURL)
		fmt.Println("Run 'synkro update' to install.")
	}

	// Actualizar último check
	cfg.LastUpdateCheck = now
	config.Save(cfg)
}
