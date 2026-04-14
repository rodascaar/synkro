package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	DatabasePath        string
	Debug               bool
	MaxTokens           int
	SessionBuffer       int
	CacheSize           int
	SimilarityThreshold float64
	EmbeddingDim        int
	ModelType           string
	AutoUpdateCheck     bool
	CheckUpdateOnStart  bool
	LastUpdateCheck     int
	configPath          string
}

func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	configDir := filepath.Join(home, ".synkro")
	configPath := filepath.Join(configDir, "config.json")

	cfg := &Config{
		DatabasePath:        getEnv("SYNKRO_DB_PATH", "memory.db"),
		Debug:               getBoolEnv("SYNKRO_DEBUG", false),
		MaxTokens:           getIntEnv("SYNKRO_MAX_TOKENS", 4000),
		SessionBuffer:       getIntEnv("SYNKRO_SESSION_BUFFER", 20),
		CacheSize:           getIntEnv("SYNKRO_CACHE_SIZE", 1000),
		SimilarityThreshold: getFloatEnv("SYNKRO_SIMILARITY_THRESHOLD", 0.5),
		EmbeddingDim:        getIntEnv("SYNKRO_EMBEDDING_DIM", 384),
		ModelType:           getEnv("SYNKRO_MODEL_TYPE", "tfidf"),
		AutoUpdateCheck:     getBoolEnv("SYNKRO_AUTO_UPDATE", true),
		CheckUpdateOnStart:  getBoolEnv("SYNKRO_CHECK_UPDATE_ON_START", true),
		LastUpdateCheck:     getIntEnv("SYNKRO_LAST_UPDATE_CHECK", 0),
		configPath:          configPath,
	}

	cfg.loadFromFile()

	return cfg, nil
}

func (c *Config) loadFromFile() {
	data, err := os.ReadFile(c.configPath)
	if err != nil {
		return
	}

	content := string(data)
	if content == "" {
		return
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "SYNKRO_DB_PATH":
			if value != "" {
				c.DatabasePath = value
			}
		case "SYNKRO_DEBUG":
			c.Debug = getBoolEnv("SYNKRO_DEBUG", c.Debug)
			if value != "" {
				if parsed, err := strconv.ParseBool(value); err == nil {
					c.Debug = parsed
				}
			}
		case "SYNKRO_MAX_TOKENS":
			if value != "" {
				if parsed, err := strconv.Atoi(value); err == nil {
					c.MaxTokens = parsed
				}
			}
		case "SYNKRO_SESSION_BUFFER":
			if value != "" {
				if parsed, err := strconv.Atoi(value); err == nil {
					c.SessionBuffer = parsed
				}
			}
		case "SYNKRO_CACHE_SIZE":
			if value != "" {
				if parsed, err := strconv.Atoi(value); err == nil {
					c.CacheSize = parsed
				}
			}
		case "SYNKRO_SIMILARITY_THRESHOLD":
			if value != "" {
				if parsed, err := strconv.ParseFloat(value, 64); err == nil {
					c.SimilarityThreshold = parsed
				}
			}
		case "SYNKRO_EMBEDDING_DIM":
			if value != "" {
				if parsed, err := strconv.Atoi(value); err == nil {
					c.EmbeddingDim = parsed
				}
			}
		case "SYNKRO_MODEL_TYPE":
			if value != "" {
				c.ModelType = value
			}
		case "SYNKRO_AUTO_UPDATE":
			if value != "" {
				if parsed, err := strconv.ParseBool(value); err == nil {
					c.AutoUpdateCheck = parsed
				}
			}
		case "SYNKRO_CHECK_UPDATE_ON_START":
			if value != "" {
				if parsed, err := strconv.ParseBool(value); err == nil {
					c.CheckUpdateOnStart = parsed
				}
			}
		case "SYNKRO_LAST_UPDATE_CHECK":
			if value != "" {
				if parsed, err := strconv.Atoi(value); err == nil {
					c.LastUpdateCheck = parsed
				}
			}
		}
	}
}

func Save(cfg *Config) error {
	configDir := filepath.Dir(cfg.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	var lines []string
	if cfg.DatabasePath != "memory.db" {
		lines = append(lines, "SYNKRO_DB_PATH="+cfg.DatabasePath)
	}
	if cfg.Debug {
		lines = append(lines, "SYNKRO_DEBUG=true")
	}
	if cfg.MaxTokens != 4000 {
		lines = append(lines, "SYNKRO_MAX_TOKENS="+strconv.Itoa(cfg.MaxTokens))
	}
	if cfg.SessionBuffer != 20 {
		lines = append(lines, "SYNKRO_SESSION_BUFFER="+strconv.Itoa(cfg.SessionBuffer))
	}
	if cfg.CacheSize != 1000 {
		lines = append(lines, "SYNKRO_CACHE_SIZE="+strconv.Itoa(cfg.CacheSize))
	}
	if cfg.SimilarityThreshold != 0.5 {
		lines = append(lines, "SYNKRO_SIMILARITY_THRESHOLD="+strconv.FormatFloat(cfg.SimilarityThreshold, 'f', -1, 64))
	}
	if cfg.EmbeddingDim != 384 {
		lines = append(lines, "SYNKRO_EMBEDDING_DIM="+strconv.Itoa(cfg.EmbeddingDim))
	}
	if cfg.ModelType != "tfidf" {
		lines = append(lines, "SYNKRO_MODEL_TYPE="+cfg.ModelType)
	}
	if !cfg.AutoUpdateCheck {
		lines = append(lines, "SYNKRO_AUTO_UPDATE=false")
	}
	if !cfg.CheckUpdateOnStart {
		lines = append(lines, "SYNKRO_CHECK_UPDATE_ON_START=false")
	}
	if cfg.LastUpdateCheck != 0 {
		lines = append(lines, "SYNKRO_LAST_UPDATE_CHECK="+strconv.Itoa(cfg.LastUpdateCheck))
	}

	content := strings.Join(lines, "\n")
	if content != "" {
		content += "\n"
	}

	return os.WriteFile(cfg.configPath, []byte(content), 0644)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getFloatEnv(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}
