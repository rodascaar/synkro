package config

import (
	"os"
	"strconv"
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
}

func Load() (*Config, error) {
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
	}
	return cfg, nil
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

func Save(cfg *Config) error {
	// Guardar configuración en variables de entorno para la próxima sesión
	os.Setenv("SYNKRO_LAST_UPDATE_CHECK", strconv.Itoa(cfg.LastUpdateCheck))
	return nil
}
