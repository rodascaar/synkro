package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	DatabasePath      string  `json:"database_path"`
	ModelType         string  `json:"model_type"`
	ModelDir          string  `json:"model_dir"`
	PreferredModel    string  `json:"preferred_model"`
	ConflictThreshold float64 `json:"conflict_threshold,omitempty"`
	configPath        string  `json:"-"`
}

func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	configPath := filepath.Join(home, ".synkro", "config.json")

	cfg := &Config{
		DatabasePath:   getEnv("SYNKRO_DB_PATH", "memory.db"),
		ModelType:      getEnv("SYNKRO_MODEL_TYPE", "tfidf"),
		ModelDir:       getEnv("SYNKRO_MODEL_DIR", "models"),
		PreferredModel: getEnv("SYNKRO_PREFERRED_MODEL", "all-MiniLM-L6-v2"),
		configPath:     configPath,
	}

	cfg.loadFromFile()
	cfg.applyEnvOverrides()

	return cfg, nil
}

// fieldSetter vincula una variable de entorno a su setter en el Config.
type fieldSetter struct {
	envKey string
	set    func(raw string)
}

func (c *Config) setters() []fieldSetter {
	return []fieldSetter{
		{"SYNKRO_DB_PATH", func(v string) {
			if v != "" {
				c.DatabasePath = v
			}
		}},
		{"SYNKRO_MODEL_TYPE", func(v string) {
			if v != "" {
				c.ModelType = v
			}
		}},
		{"SYNKRO_MODEL_DIR", func(v string) {
			if v != "" {
				c.ModelDir = v
			}
		}},
		{"SYNKRO_PREFERRED_MODEL", func(v string) {
			if v != "" {
				c.PreferredModel = v
			}
		}},
		{"SYNKRO_CONFLICT_THRESHOLD", func(v string) {
			if v != "" {
				if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 && f <= 1 {
					c.ConflictThreshold = f
				}
			}
		}},
	}
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
	settersByKey := make(map[string]func(string), 6)
	for _, f := range c.setters() {
		settersByKey[f.envKey] = f.set
	}

	for _, line := range strings.Split(content, "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		if set, ok := settersByKey[strings.TrimSpace(parts[0])]; ok {
			set(strings.TrimSpace(parts[1]))
		}
	}

	_ = Save(c)
}

func (c *Config) applyEnvOverrides() {
	for _, f := range c.setters() {
		if v := os.Getenv(f.envKey); v != "" {
			f.set(v)
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
