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

// ClaudeConfig holds Claude-specific LLM settings. ReasoningEffort sets
// output_config.effort on every request.
type ClaudeConfig struct {
	APIKey             string `koanf:"apikey"`
	Model              string `koanf:"model"`
	ReasoningEffort    string `koanf:"reasoningeffort"`
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

// ChatGPTConfig holds settings for the ChatGPT-plan provider, which uses the
// Responses API on the signed-in user's ChatGPT plan. CredentialsFile is the
// file written by cmd/chatgpt-login; the server must be able to rewrite it,
// because each token refresh rotates the refresh token.
type ChatGPTConfig struct {
	CredentialsFile    string `koanf:"credentialsfile"`
	Model              string `koanf:"model"`
	ReasoningEffort    string `koanf:"reasoningeffort"`
	ContextTokenBudget int    `koanf:"contexttokenbudget"`
}

// ClaudeCodeConfig points the claudecode provider at cmd/claude-bridge, which
// runs the Claude Code CLI on the owner's Claude subscription. The bridge
// sets the model and effort.
type ClaudeCodeConfig struct {
	Endpoint           string `koanf:"endpoint"`
	Token              string `koanf:"token"`
	ContextTokenBudget int    `koanf:"contexttokenbudget"`
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
	// Skills gives this profile the load_skill and read_skill_file tools for
	// the skills in llm.skillsdir. Each skill the model reads adds a round
	// trip, so it is meant for the GM-turn profile.
	Skills bool `koanf:"skills"`
}

// LLMRoutesConfig maps high-level game flows to named LLM profiles.
type LLMRoutesConfig struct {
	GMTurn         string `koanf:"gmturn"`
	PostTurnState  string `koanf:"postturnstate"`
	ChoiceFallback string `koanf:"choicefallback"`
	Summary        string `koanf:"summary"`
	Worldgen       string `koanf:"worldgen"`
}

// LLMConfig holds provider selection and per-provider settings. When
// FallbackProvider is set, a request that fails on Provider is sent again to
// FallbackProvider.
type LLMConfig struct {
	Provider         string           `koanf:"provider"`
	FallbackProvider string           `koanf:"fallbackprovider"`
	Ollama           OllamaConfig     `koanf:"ollama"`
	Claude           ClaudeConfig     `koanf:"claude"`
	OpenRouter       OpenRouterConfig `koanf:"openrouter"`
	ChatGPT          ChatGPTConfig    `koanf:"chatgpt"`
	ClaudeCode       ClaudeCodeConfig `koanf:"claudecode"`
	// SkillsDir holds the skills that profiles with skills can load.
	SkillsDir string                      `koanf:"skillsdir"`
	Profiles  map[string]LLMProfileConfig `koanf:"profiles"`
	Routes    LLMRoutesConfig             `koanf:"routes"`
}

// ContextTokenBudget returns the configured context token budget for the active provider.
func (c LLMConfig) ContextTokenBudget() int {
	switch c.Provider {
	case "claude":
		return c.Claude.ContextTokenBudget
	case "openrouter":
		return c.OpenRouter.ContextTokenBudget
	case "chatgpt":
		return c.ChatGPT.ContextTokenBudget
	case "claudecode":
		return c.ClaudeCode.ContextTokenBudget
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

// ServerConfig holds HTTP server settings. AllowRegistration controls whether
// POST /api/v1/auth/register creates accounts; turn it off on a public server
// whose model costs fall on one person.
type ServerConfig struct {
	Port              int    `koanf:"port"`
	JWTSecret         string `koanf:"jwtsecret"`
	AllowRegistration bool   `koanf:"allowregistration"`
}

// Config is the top-level configuration, composed of per-concern slices.
type Config struct {
	DB     DBConfig     `koanf:"db"`
	LLM    LLMConfig    `koanf:"llm"`
	Server ServerConfig `koanf:"server"`
}

// Validate checks that the configuration is internally consistent.
func (c *Config) Validate() error {
	if !knownProviders[c.LLM.Provider] {
		return fmt.Errorf("unknown llm provider: %q", c.LLM.Provider)
	}
	if c.LLM.FallbackProvider != "" && !knownProviders[c.LLM.FallbackProvider] {
		return fmt.Errorf("unknown llm fallback provider: %q", c.LLM.FallbackProvider)
	}
	if c.usesProvider("chatgpt") && strings.TrimSpace(c.LLM.ChatGPT.CredentialsFile) == "" {
		return errors.New("chatgpt provider requires a credentials file (set llm.chatgpt.credentialsfile or EDDA_LLM_CHATGPT_CREDENTIALSFILE)")
	}
	if !openRouterReasoningEfforts[c.LLM.Claude.ReasoningEffort] {
		return fmt.Errorf("unknown claude reasoning effort: %q (use low, medium, high, xhigh or max)", c.LLM.Claude.ReasoningEffort)
	}
	if c.usesProvider("claudecode") && (strings.TrimSpace(c.LLM.ClaudeCode.Endpoint) == "" || strings.TrimSpace(c.LLM.ClaudeCode.Token) == "") {
		return errors.New("claudecode provider requires llm.claudecode.endpoint and llm.claudecode.token (EDDA_LLM_CLAUDECODE_ENDPOINT, EDDA_LLM_CLAUDECODE_TOKEN)")
	}
	if !openRouterReasoningEfforts[c.LLM.ChatGPT.ReasoningEffort] {
		return fmt.Errorf("unknown chatgpt reasoning effort: %q (use none, minimal, low, medium, high, xhigh or max)", c.LLM.ChatGPT.ReasoningEffort)
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
		if !knownProviders[provider] {
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

var knownProviders = map[string]bool{"ollama": true, "claude": true, "openrouter": true, "chatgpt": true, "claudecode": true}

// usesProvider reports whether name is the active, fallback or a profile provider.
func (c *Config) usesProvider(name string) bool {
	if c.LLM.Provider == name || c.LLM.FallbackProvider == name {
		return true
	}
	for _, p := range c.LLM.Profiles {
		if p.Provider == name {
			return true
		}
	}
	return false
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
		"llm.claude.model":                  "claude-sonnet-5-5",
		"llm.claude.contexttokenbudget":     8000,
		"llm.openrouter.model":              "openai/gpt-4o-mini",
		"llm.openrouter.contexttokenbudget": 8000,
		"llm.chatgpt.model":                 "gpt-6.1-terra",
		"llm.chatgpt.contexttokenbudget":    8000,
		"llm.claudecode.contexttokenbudget": 8000,
		"llm.skillsdir":                     "skills",
		"server.port":                       8080,
		"server.jwtsecret":                  "",
		"server.allowregistration":          true,
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
