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
		if cfg.CheckUpdateOnStart && cmd.Name() != "update" && cmd.Name() != "self-update" && cmd.Name() != "mcp" {
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
		_, _ = fmt.Scanln(&response)

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
			fmt.Println("Note: For full embeddings functionality, download models using:")
			fmt.Println("  synkro model download all-MiniLM-L6-v2")
			fmt.Println("  synkro model download paraphrase-multilingual-MiniLM-L12-v2")
		}
	},
}

var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "Manage embedding models",
}

var modelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available models",
	Run: func(cmd *cobra.Command, args []string) {
		modelMgr := embeddings.NewModelManager(&embeddings.ManagerConfig{
			DownloadDir:  "models",
			CacheDir:     "cache",
			MaxModels:    5,
			AutoDownload: false,
		})

		models := modelMgr.ListModels()
		fmt.Println("Available Models:")
		fmt.Println()

		for _, model := range models {
			status := "📥"
			if model.Downloaded {
				status = "✓"
			}

			fmt.Printf("%s %s\n", status, model.Name)
			fmt.Printf("  Dimension: %d\n", model.Dimension)
			fmt.Printf("  Language: %s\n", model.Language)
			fmt.Printf("  License: %s\n", model.License)

			if model.Params != "" {
				fmt.Printf("  Parameters: %s\n", model.Params)
			}

			if model.MaxSeqLen > 0 {
				fmt.Printf("  Max Sequence Length: %d\n", model.MaxSeqLen)
			}

			if len(model.Benchmarks) > 0 {
				fmt.Printf("  Benchmarks:\n")
				for name, score := range model.Benchmarks {
					fmt.Printf("    %s: %.2f\n", name, score)
				}
			}

			if model.Description != "" {
				fmt.Printf("  Description: %s\n", model.Description)
			}

			if model.Downloaded {
				sizeMB := float64(model.FileSize) / (1024 * 1024)
				fmt.Printf("  Downloaded: %.2f MB\n", sizeMB)
			}

			fmt.Println()
		}
	},
}

var modelDownloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download a model",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		modelName := args[0]

		modelMgr := embeddings.NewModelManager(&embeddings.ManagerConfig{
			DownloadDir:  "models",
			CacheDir:     "cache",
			MaxModels:    5,
			AutoDownload: false,
		})

		fmt.Printf("Downloading model: %s\n", modelName)

		progressChan := make(chan float64)
		go func() {
			if err := modelMgr.DownloadModel(context.Background(), modelName, func(progress float64) {
				progressChan <- progress
			}); err != nil {
				fmt.Fprintf(os.Stderr, "Error downloading model: %v\n", err)
				os.Exit(1)
			}
			close(progressChan)
		}()

		for progress := range progressChan {
			percent := int(progress * 100)
			fmt.Printf("\rProgress: %d%% [%-50s]", percent, strings.Repeat("=", percent/2)+strings.Repeat(" ", 50-percent/2))
		}
		fmt.Println("\n✓ Model downloaded successfully")
	},
}

var modelDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a model",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		modelName := args[0]

		modelMgr := embeddings.NewModelManager(&embeddings.ManagerConfig{
			DownloadDir:  "models",
			CacheDir:     "cache",
			MaxModels:    5,
			AutoDownload: false,
		})

		if err := modelMgr.DeleteModel(modelName); err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting model: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Model %s deleted successfully\n", modelName)
	},
}

var modelInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show model information",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		modelName := args[0]

		modelMgr := embeddings.NewModelManager(&embeddings.ManagerConfig{
			DownloadDir:  "models",
			CacheDir:     "cache",
			MaxModels:    5,
			AutoDownload: false,
		})

		model, err := modelMgr.GetModel(modelName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Model: %s\n", model.Name)
		fmt.Printf("  Dimension: %d\n", model.Dimension)
		fmt.Printf("  Language: %s\n", model.Language)
		fmt.Printf("  License: %s\n", model.License)
		fmt.Printf("  Downloaded: %v\n", model.Downloaded)

		if model.Params != "" {
			fmt.Printf("  Parameters: %s\n", model.Params)
		}

		if model.MaxSeqLen > 0 {
			fmt.Printf("  Max Sequence Length: %d\n", model.MaxSeqLen)
		}

		if len(model.Benchmarks) > 0 {
			fmt.Printf("  Benchmarks:\n")
			for name, score := range model.Benchmarks {
				fmt.Printf("    %s: %.2f\n", name, score)
			}
		}

		if model.Description != "" {
			fmt.Printf("  Description: %s\n", model.Description)
		}

		if model.Downloaded {
			sizeMB := float64(model.FileSize) / (1024 * 1024)
			fmt.Printf("  Size: %.2f MB\n", sizeMB)
			fmt.Printf("  Path: %s\n", model.DownloadPath)
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
	rootCmd.AddCommand(checkUpdateCmd)
	rootCmd.AddCommand(selfUpdateCmd)
	rootCmd.AddCommand(examplesCmd)
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(modelCmd)
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
	_ = config.Save(cfg)
}
