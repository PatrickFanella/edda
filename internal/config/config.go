package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

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

// OllamaConfig holds Ollama-compatible LLM settings. APIKey is optional and is
// only required when the configured endpoint points at the Switchyard model
// broker or another authenticated Ollama-compatible broker. When empty,
// requests are sent to a vanilla ollama server.
type OllamaConfig struct {
	Endpoint           string `koanf:"endpoint"`
	EmbeddingEndpoint  string `koanf:"embeddingendpoint"`
	Model              string `koanf:"model"`
	EmbeddingModel     string `koanf:"embeddingmodel"`
	EmbeddingDimension int    `koanf:"embeddingdimension"`
	APIKey             string `koanf:"apikey"`
	ContextTokenBudget int    `koanf:"contexttokenbudget"`
	TimeoutSeconds     int    `koanf:"timeoutseconds"`
}

// RequestTimeout returns the configured Ollama request timeout, defaulting to
// three minutes when unset or invalid.
func (c OllamaConfig) RequestTimeout() time.Duration {
	if c.TimeoutSeconds <= 0 {
		return 3 * time.Minute
	}
	return time.Duration(c.TimeoutSeconds) * time.Second
}

// ClaudeConfig holds Claude-specific LLM settings.
type ClaudeConfig struct {
	APIKey             string `koanf:"apikey"`
	Model              string `koanf:"model"`
	ContextTokenBudget int    `koanf:"contexttokenbudget"`
}

// OpenRouterConfig holds OpenRouter-specific LLM settings. FallbackModels are
// tried in order by OpenRouter when Model errors; from the environment they are
// a comma-separated list. ReasoningEffort sets reasoning.effort on every request.
type OpenRouterConfig struct {
	APIKey             string   `koanf:"apikey"`
	Model              string   `koanf:"model"`
	FallbackModels     []string `koanf:"fallbackmodels"`
	ReasoningEffort    string   `koanf:"reasoningeffort"`
	ContextTokenBudget int      `koanf:"contexttokenbudget"`
}

var openRouterReasoningEfforts = map[string]bool{
	"": true, "none": true, "minimal": true, "low": true, "medium": true, "high": true, "xhigh": true, "max": true,
}

// LLMProfileConfig names a model/provider profile that can be routed to a
// specific game flow. Empty provider/model/budget fields inherit from the
// top-level active provider configuration.
type LLMProfileConfig struct {
	Provider           string `koanf:"provider"`
	Model              string `koanf:"model"`
	ContextTokenBudget int    `koanf:"contexttokenbudget"`
	ToolSupport        bool   `koanf:"toolsupport"`
}

// LLMRoutesConfig maps high-level game flows to named LLM profiles.
type LLMRoutesConfig struct {
	GMTurn         string `koanf:"gmturn"`
	PostTurnState  string `koanf:"postturnstate"`
	ChoiceFallback string `koanf:"choicefallback"`
	Summary        string `koanf:"summary"`
	Worldgen       string `koanf:"worldgen"`
}

// LLMConfig holds provider selection and per-provider settings.
type LLMConfig struct {
	Provider   string                      `koanf:"provider"`
	Ollama     OllamaConfig                `koanf:"ollama"`
	Claude     ClaudeConfig                `koanf:"claude"`
	OpenRouter OpenRouterConfig            `koanf:"openrouter"`
	Profiles   map[string]LLMProfileConfig `koanf:"profiles"`
	Routes     LLMRoutesConfig             `koanf:"routes"`
}

// ContextTokenBudget returns the configured context token budget for the active provider.
func (c LLMConfig) ContextTokenBudget() int {
	switch c.Provider {
	case "claude":
		return c.Claude.ContextTokenBudget
	case "openrouter":
		return c.OpenRouter.ContextTokenBudget
	default:
		return c.Ollama.ContextTokenBudget
	}
}

// RoutedProfileNames returns the non-empty profile names referenced by routes.
func (c LLMConfig) RoutedProfileNames() map[string]string {
	routes := map[string]string{}
	if c.Routes.GMTurn != "" {
		routes["gmturn"] = c.Routes.GMTurn
	}
	if c.Routes.PostTurnState != "" {
		routes["postturnstate"] = c.Routes.PostTurnState
	}
	if c.Routes.ChoiceFallback != "" {
		routes["choicefallback"] = c.Routes.ChoiceFallback
	}
	if c.Routes.Summary != "" {
		routes["summary"] = c.Routes.Summary
	}
	if c.Routes.Worldgen != "" {
		routes["worldgen"] = c.Routes.Worldgen
	}
	return routes
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port      int    `koanf:"port"`
	JWTSecret string `koanf:"jwtsecret"`
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
	case "ollama", "claude", "openrouter":
	default:
		return fmt.Errorf("unknown llm provider: %q", c.LLM.Provider)
	}
	if c.LLM.Provider == "claude" && c.LLM.Claude.APIKey == "" {
		return errors.New("claude provider requires api key (set llm.claude.apikey, EDDA_LLM_CLAUDE_APIKEY, or ANTHROPIC_API_KEY)")
	}
	if c.LLM.Provider == "openrouter" && c.LLM.OpenRouter.APIKey == "" {
		return errors.New("openrouter provider requires api key (set llm.openrouter.apikey or EDDA_LLM_OPENROUTER_APIKEY)")
	}
	if !openRouterReasoningEfforts[c.LLM.OpenRouter.ReasoningEffort] {
		return fmt.Errorf("unknown openrouter reasoning effort: %q (use none, minimal, low, medium, high, xhigh or max)", c.LLM.OpenRouter.ReasoningEffort)
	}
	if c.LLM.Provider == "ollama" && c.LLM.Ollama.EmbeddingDimension <= 0 {
		return errors.New("ollama embedding dimension must be positive")
	}
	for name, profile := range c.LLM.Profiles {
		provider := profile.Provider
		if provider == "" {
			provider = c.LLM.Provider
		}
		switch provider {
		case "ollama", "claude", "openrouter":
		default:
			return fmt.Errorf("llm profile %q has unknown provider: %q", name, provider)
		}
	}
	for route, profileName := range c.LLM.RoutedProfileNames() {
		if _, ok := c.LLM.Profiles[profileName]; !ok {
			return fmt.Errorf("llm route %q references unknown profile %q", route, profileName)
		}
	}
	return nil
}

func Load(path string) (Config, error) {
	k := koanf.New(".")

	defaults := map[string]any{
		"db.url":                            "postgres://edda:edda@localhost:5432/edda?sslmode=disable",
		"llm.provider":                      "ollama",
		"llm.ollama.endpoint":               "http://localhost:11434",
		"llm.ollama.model":                  "qwen3:14b",
		"llm.ollama.embeddingendpoint":      "",
		"llm.ollama.embeddingmodel":         "nomic-embed-text",
		"llm.ollama.embeddingdimension":     768,
		"llm.ollama.apikey":                 "",
		"llm.ollama.contexttokenbudget":     8000,
		"llm.ollama.timeoutseconds":         600,
		"llm.claude.model":                  "claude-sonnet-4-6",
		"llm.claude.contexttokenbudget":     8000,
		"llm.openrouter.model":              "openai/gpt-4o-mini",
		"llm.openrouter.contexttokenbudget": 8000,
		"server.port":                       8080,
		"server.jwtsecret":                  "",
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

	// Support ANTHROPIC_API_KEY as the lowest-priority env var for the Claude API key.
	// Note: this is loaded after the config file, so it takes precedence over file-based config.
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		if err := k.Load(confmap.Provider(map[string]any{"llm.claude.apikey": apiKey}, "."), nil); err != nil {
			return Config{}, err
		}
	}

	// Support OPENROUTER_API_KEY as the lowest-priority env var for the OpenRouter API key.
	if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
		if err := k.Load(confmap.Provider(map[string]any{"llm.openrouter.apikey": apiKey}, "."), nil); err != nil {
			return Config{}, err
		}
	}

	// EDDA_-prefixed env vars have the highest priority (includes EDDA_LLM_CLAUDE_APIKEY).
	if err := k.Load(env.Provider("EDDA_", ".", func(key string) string {
		trimmed := strings.TrimPrefix(key, "EDDA_")
		return strings.ToLower(strings.ReplaceAll(trimmed, "_", "."))
	}), nil); err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return Config{}, err
	}
	cfg.LLM.OpenRouter.FallbackModels = splitList(cfg.LLM.OpenRouter.FallbackModels)

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// splitList splits comma-separated entries (an env var arrives as one string)
// and drops blanks.
func splitList(values []string) []string {
	var out []string
	for _, v := range values {
		for _, part := range strings.Split(v, ",") {
			if part = strings.TrimSpace(part); part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}
