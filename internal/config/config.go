package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// DBConfig holds database connection settings.
type DBConfig struct {
	URL string `koanf:"url"`
}

// OllamaConfig holds Ollama-specific LLM settings.
type OllamaConfig struct {
	Endpoint string `koanf:"endpoint"`
	Model    string `koanf:"model"`
}

// ClaudeConfig holds Claude-specific LLM settings.
type ClaudeConfig struct {
	APIKey string `koanf:"apikey"`
	Model  string `koanf:"model"`
}

// LLMConfig holds provider selection and per-provider settings.
type LLMConfig struct {
	Provider string       `koanf:"provider"`
	Ollama   OllamaConfig `koanf:"ollama"`
	Claude   ClaudeConfig `koanf:"claude"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port int `koanf:"port"`
}

// Config is the top-level configuration, composed of per-concern slices.
type Config struct {
	DB     DBConfig     `koanf:"db"`
	LLM    LLMConfig    `koanf:"llm"`
	Server ServerConfig `koanf:"server"`
}

// Validate checks that the configuration is internally consistent.
func (c *Config) Validate() error {
	switch c.LLM.Provider {
	case "ollama", "claude":
	default:
		return fmt.Errorf("unknown llm provider: %q", c.LLM.Provider)
	}
	if c.LLM.Provider == "claude" && c.LLM.Claude.APIKey == "" {
		return errors.New("claude provider requires api key (set llm.claude.apikey or GM_LLM_CLAUDE_APIKEY)")
	}
	return nil
}

func Load(path string) (Config, error) {
	k := koanf.New(".")

	defaults := map[string]any{
		"db.url":              "postgres://game_master:game_master@localhost:5432/game_master?sslmode=disable",
		"llm.provider":        "ollama",
		"llm.ollama.endpoint": "http://localhost:11434",
		"llm.ollama.model":    "llama3.2",
		"llm.claude.model":    "claude-sonnet-4-6",
		"server.port":         8080,
	}

	if err := k.Load(confmap.Provider(defaults, "."), nil); err != nil {
		return Config{}, err
	}

	if path != "" {
		if _, err := os.Stat(path); err == nil {
			if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
				return Config{}, err
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return Config{}, err
		}
	}

	if err := k.Load(env.Provider("GM_", ".", func(key string) string {
		trimmed := strings.TrimPrefix(key, "GM_")
		return strings.ToLower(strings.ReplaceAll(trimmed, "_", "."))
	}), nil); err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return Config{}, err
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
