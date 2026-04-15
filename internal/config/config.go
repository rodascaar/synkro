package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	DatabasePath        string  `json:"database_path"`
	Debug               bool    `json:"debug"`
	MaxTokens           int     `json:"max_tokens"`
	SessionBuffer       int     `json:"session_buffer"`
	CacheSize           int     `json:"cache_size"`
	SimilarityThreshold float64 `json:"similarity_threshold"`
	EmbeddingDim        int     `json:"embedding_dim"`
	ModelType           string  `json:"model_type"`
	ModelDir            string  `json:"model_dir"`
	PreferredModel      string  `json:"preferred_model"`
	AutoUpdateCheck     bool    `json:"auto_update"`
	CheckUpdateOnStart  bool    `json:"check_update_on_start"`
	LastUpdateCheck     int     `json:"last_update_check"`
	configPath          string  `json:"-"`
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
		ModelDir:            getEnv("SYNKRO_MODEL_DIR", "models"),
		PreferredModel:      getEnv("SYNKRO_PREFERRED_MODEL", "all-MiniLM-L6-v2"),
		AutoUpdateCheck:     getBoolEnv("SYNKRO_AUTO_UPDATE", true),
		CheckUpdateOnStart:  getBoolEnv("SYNKRO_CHECK_UPDATE_ON_START", true),
		LastUpdateCheck:     getIntEnv("SYNKRO_LAST_UPDATE_CHECK", 0),
		configPath:          configPath,
	}

	cfg.loadFromFile()

	cfg.applyEnvOverrides()

	return cfg, nil
}

func (c *Config) loadFromFile() {
	data, err := os.ReadFile(c.configPath)
	if err != nil {
		return
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		return
	}

	if strings.Contains(content, "=") && !strings.HasPrefix(content, "{") {
		c.migrateFromKeyValue(content)
		return
	}

	if err := json.Unmarshal(data, c); err != nil {
		return
	}
}

func (c *Config) migrateFromKeyValue(content string) {
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
			if parsed, err := strconv.ParseBool(value); err == nil {
				c.Debug = parsed
			}
		case "SYNKRO_MAX_TOKENS":
			if parsed, err := strconv.Atoi(value); err == nil {
				c.MaxTokens = parsed
			}
		case "SYNKRO_SESSION_BUFFER":
			if parsed, err := strconv.Atoi(value); err == nil {
				c.SessionBuffer = parsed
			}
		case "SYNKRO_CACHE_SIZE":
			if parsed, err := strconv.Atoi(value); err == nil {
				c.CacheSize = parsed
			}
		case "SYNKRO_SIMILARITY_THRESHOLD":
			if parsed, err := strconv.ParseFloat(value, 64); err == nil {
				c.SimilarityThreshold = parsed
			}
		case "SYNKRO_EMBEDDING_DIM":
			if parsed, err := strconv.Atoi(value); err == nil {
				c.EmbeddingDim = parsed
			}
		case "SYNKRO_MODEL_TYPE":
			if value != "" {
				c.ModelType = value
			}
		case "SYNKRO_AUTO_UPDATE":
			if parsed, err := strconv.ParseBool(value); err == nil {
				c.AutoUpdateCheck = parsed
			}
		case "SYNKRO_CHECK_UPDATE_ON_START":
			if parsed, err := strconv.ParseBool(value); err == nil {
				c.CheckUpdateOnStart = parsed
			}
		case "SYNKRO_LAST_UPDATE_CHECK":
			if parsed, err := strconv.Atoi(value); err == nil {
				c.LastUpdateCheck = parsed
			}
		}
	}

	_ = Save(c)
}

func (c *Config) applyEnvOverrides() {
	if v := os.Getenv("SYNKRO_DB_PATH"); v != "" {
		c.DatabasePath = v
	}
	if v := os.Getenv("SYNKRO_DEBUG"); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			c.Debug = parsed
		}
	}
	if v := os.Getenv("SYNKRO_MAX_TOKENS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			c.MaxTokens = parsed
		}
	}
	if v := os.Getenv("SYNKRO_SESSION_BUFFER"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			c.SessionBuffer = parsed
		}
	}
	if v := os.Getenv("SYNKRO_CACHE_SIZE"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			c.CacheSize = parsed
		}
	}
	if v := os.Getenv("SYNKRO_SIMILARITY_THRESHOLD"); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			c.SimilarityThreshold = parsed
		}
	}
	if v := os.Getenv("SYNKRO_EMBEDDING_DIM"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			c.EmbeddingDim = parsed
		}
	}
	if v := os.Getenv("SYNKRO_MODEL_TYPE"); v != "" {
		c.ModelType = v
	}
	if v := os.Getenv("SYNKRO_MODEL_DIR"); v != "" {
		c.ModelDir = v
	}
	if v := os.Getenv("SYNKRO_PREFERRED_MODEL"); v != "" {
		c.PreferredModel = v
	}
	if v := os.Getenv("SYNKRO_AUTO_UPDATE"); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			c.AutoUpdateCheck = parsed
		}
	}
	if v := os.Getenv("SYNKRO_CHECK_UPDATE_ON_START"); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			c.CheckUpdateOnStart = parsed
		}
	}
	if v := os.Getenv("SYNKRO_LAST_UPDATE_CHECK"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			c.LastUpdateCheck = parsed
		}
	}
}

func Save(cfg *Config) error {
	configDir := filepath.Dir(cfg.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cfg.configPath, append(data, '\n'), 0644)
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
