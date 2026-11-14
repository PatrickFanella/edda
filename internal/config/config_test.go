package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadUsesDefaultsWhenFileIsMissing(t *testing.T) {
	unsetenv(t, "EDDA_DB_URL", "EDDA_LLM_PROVIDER", "EDDA_LLM_OLLAMA_ENDPOINT", "EDDA_LLM_OLLAMA_MODEL")

	cfg, err := Load(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.DB.URL != "postgres://edda:edda@localhost:5432/edda?sslmode=disable" {
		t.Fatalf("unexpected default db url: %q", cfg.DB.URL)
	}
	if cfg.LLM.Provider != "ollama" {
		t.Fatalf("unexpected default provider: %q", cfg.LLM.Provider)
	}
	if cfg.LLM.Ollama.Endpoint != "http://localhost:11434" {
		t.Fatalf("unexpected default ollama endpoint: %q", cfg.LLM.Ollama.Endpoint)
	}
	if cfg.LLM.Ollama.ContextTokenBudget != 8000 {
		t.Fatalf("unexpected default ollama context token budget: %d", cfg.LLM.Ollama.ContextTokenBudget)
	}
	if cfg.LLM.Ollama.TimeoutSeconds != 600 {
		t.Fatalf("unexpected default ollama timeout seconds: %d", cfg.LLM.Ollama.TimeoutSeconds)
	}
	if cfg.LLM.Ollama.EmbeddingDimension != 768 {
		t.Fatalf("unexpected default ollama embedding dimension: %d", cfg.LLM.Ollama.EmbeddingDimension)
	}
}

func TestLoadMergesFileAndEnvironment(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")

	const fileConfig = `db:
  url: postgres://from-file/filedb?sslmode=disable
llm:
  provider: openai
  ollama:
    endpoint: http://from-file:11434
    model: file-model
`

	if err := os.WriteFile(configPath, []byte(fileConfig), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	t.Setenv("EDDA_LLM_PROVIDER", "ollama")
	t.Setenv("EDDA_LLM_OLLAMA_ENDPOINT", "http://from-env:11434")

	unsetenv(t, "EDDA_DB_URL", "EDDA_LLM_OLLAMA_MODEL")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.DB.URL != "postgres://from-file/filedb?sslmode=disable" {
		t.Fatalf("expected db url from file, got %q", cfg.DB.URL)
	}
	if cfg.LLM.Provider != "ollama" {
		t.Fatalf("expected env to override provider, got %q", cfg.LLM.Provider)
	}
	if cfg.LLM.Ollama.Endpoint != "http://from-env:11434" {
		t.Fatalf("expected env to override ollama endpoint, got %q", cfg.LLM.Ollama.Endpoint)
	}
	if cfg.LLM.Ollama.Model != "file-model" {
		t.Fatalf("expected model from file, got %q", cfg.LLM.Ollama.Model)
	}
	if cfg.LLM.Ollama.EmbeddingDimension != 768 {
		t.Fatalf("expected embedding dimension default, got %d", cfg.LLM.Ollama.EmbeddingDimension)
	}
}

func TestLoadClaudeConfigFromEnv(t *testing.T) {
	t.Setenv("EDDA_LLM_PROVIDER", "claude")
	t.Setenv("EDDA_LLM_CLAUDE_APIKEY", "sk-ant-test-key")
	t.Setenv("EDDA_LLM_CLAUDE_MODEL", "claude-opus-4-6")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.LLM.Provider != "claude" {
		t.Fatalf("expected claude provider, got %q", cfg.LLM.Provider)
	}
	if cfg.LLM.Claude.APIKey != "sk-ant-test-key" {
		t.Fatalf("expected claude api key, got %q", cfg.LLM.Claude.APIKey)
	}
	if cfg.LLM.Claude.Model != "claude-opus-4-6" {
		t.Fatalf("expected claude model override, got %q", cfg.LLM.Claude.Model)
	}
}

func TestLoadServerPortDefault(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Fatalf("expected default port 8080, got %d", cfg.Server.Port)
	}
}

func TestLoadClaudeModelDefault(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.LLM.Claude.Model != "claude-sonnet-5-5" {
		t.Fatalf("expected default claude model, got %q", cfg.LLM.Claude.Model)
	}
	if cfg.LLM.Claude.ContextTokenBudget != 8000 {
		t.Fatalf("expected default claude context token budget, got %d", cfg.LLM.Claude.ContextTokenBudget)
	}
}

func TestLoadContextTokenBudgetFromEnv(t *testing.T) {
	t.Setenv("EDDA_LLM_PROVIDER", "claude")
	t.Setenv("EDDA_LLM_CLAUDE_APIKEY", "sk-ant-test-key")
	t.Setenv("EDDA_LLM_CLAUDE_CONTEXTTOKENBUDGET", "6400")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.LLM.ContextTokenBudget() != 6400 {
		t.Fatalf("expected active provider context token budget from env, got %d", cfg.LLM.ContextTokenBudget())
	}
}

func TestLoadOllamaTimeoutFromEnv(t *testing.T) {
	t.Setenv("EDDA_LLM_PROVIDER", "ollama")
	t.Setenv("EDDA_LLM_OLLAMA_TIMEOUTSECONDS", "240")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.LLM.Ollama.TimeoutSeconds != 240 {
		t.Fatalf("expected ollama timeout seconds from env, got %d", cfg.LLM.Ollama.TimeoutSeconds)
	}
}

func TestLoadOllamaEmbeddingDimensionFromEnv(t *testing.T) {
	t.Setenv("EDDA_LLM_PROVIDER", "ollama")
	t.Setenv("EDDA_LLM_OLLAMA_EMBEDDINGDIMENSION", "1024")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.LLM.Ollama.EmbeddingDimension != 1024 {
		t.Fatalf("expected ollama embedding dimension from env, got %d", cfg.LLM.Ollama.EmbeddingDimension)
	}
}

func TestLoadRejectsInvalidProvider(t *testing.T) {
	t.Setenv("EDDA_LLM_PROVIDER", "invalid")

	_, err := Load("")
	if err == nil {
		t.Fatal("expected error for invalid provider, got nil")
	}
	if !strings.Contains(err.Error(), "provider") {
		t.Fatalf("expected provider-related error, got: %v", err)
	}
}

func TestLoadRejectsClaudeWithoutAPIKey(t *testing.T) {
	t.Setenv("EDDA_LLM_PROVIDER", "claude")
	// Explicitly clear all API key env vars to ensure the test is deterministic.
	unsetenv(t, "EDDA_LLM_CLAUDE_APIKEY", "ANTHROPIC_API_KEY")

	_, err := Load("")
	if err == nil {
		t.Fatal("expected error for claude without api key, got nil")
	}
	if !strings.Contains(err.Error(), "api key") {
		t.Fatalf("expected api key error, got: %v", err)
	}
}

func TestLoadClaudeAPIKeyFromAnthropicEnv(t *testing.T) {
	t.Setenv("EDDA_LLM_PROVIDER", "claude")
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-anthropic-key")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.LLM.Claude.APIKey != "sk-ant-anthropic-key" {
		t.Fatalf("expected api key from ANTHROPIC_API_KEY, got %q", cfg.LLM.Claude.APIKey)
	}
}

func TestLoadClaudeAPIKeyFromEDDALLMEnv(t *testing.T) {
	t.Setenv("EDDA_LLM_PROVIDER", "claude")
	t.Setenv("EDDA_LLM_CLAUDE_APIKEY", "sk-ant-edda-claude-key")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.LLM.Claude.APIKey != "sk-ant-edda-claude-key" {
		t.Fatalf("expected api key from EDDA_LLM_CLAUDE_APIKEY, got %q", cfg.LLM.Claude.APIKey)
	}
}
func TestEDDALLMClaudeAPIKeyOverridesAnthropicAPIKey(t *testing.T) {
	t.Setenv("EDDA_LLM_PROVIDER", "claude")
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-lower-priority")
	t.Setenv("EDDA_LLM_CLAUDE_APIKEY", "sk-ant-higher-priority")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.LLM.Claude.APIKey != "sk-ant-higher-priority" {
		t.Fatalf("expected EDDA_LLM_CLAUDE_APIKEY to override ANTHROPIC_API_KEY, got %q", cfg.LLM.Claude.APIKey)
	}
}

func TestEDDALLMClaudeAPIKeyHighestPriority(t *testing.T) {
	t.Setenv("EDDA_LLM_PROVIDER", "claude")
	t.Setenv("EDDA_LLM_CLAUDE_APIKEY", "sk-ant-highest-priority")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.LLM.Claude.APIKey != "sk-ant-highest-priority" {
		t.Fatalf("expected EDDA_LLM_CLAUDE_APIKEY to have highest priority, got %q", cfg.LLM.Claude.APIKey)
	}
}

func unsetenv(t *testing.T, keys ...string) {
	t.Helper()
	for _, key := range keys {
		key := key
		if orig, ok := os.LookupEnv(key); ok {
			t.Cleanup(func() { _ = os.Setenv(key, orig) })
		} else {
			t.Cleanup(func() { _ = os.Unsetenv(key) })
		}
		_ = os.Unsetenv(key)
	}
}

func TestClaudeAPIKeyPrecedenceFileVsEnv(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")

	const fileConfig = `llm:
  provider: claude
  claude:
    apikey: sk-ant-from-file
`
	if err := os.WriteFile(configPath, []byte(fileConfig), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Verify file key is used when no env vars are set.
	t.Run("file_key_used_when_no_env", func(t *testing.T) {
		unsetenv(t, "EDDA_LLM_CLAUDE_APIKEY", "ANTHROPIC_API_KEY")
		cfg, err := Load(configPath)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.LLM.Claude.APIKey != "sk-ant-from-file" {
			t.Fatalf("expected file api key, got %q", cfg.LLM.Claude.APIKey)
		}
	})

	// Verify ANTHROPIC_API_KEY overrides file key.
	t.Run("anthropic_env_overrides_file", func(t *testing.T) {
		unsetenv(t, "EDDA_LLM_CLAUDE_APIKEY")
		t.Setenv("ANTHROPIC_API_KEY", "sk-ant-anthropic-override")
		cfg, err := Load(configPath)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.LLM.Claude.APIKey != "sk-ant-anthropic-override" {
			t.Fatalf("expected ANTHROPIC_API_KEY to override file key, got %q", cfg.LLM.Claude.APIKey)
		}
	})

	// Verify EDDA_LLM_CLAUDE_APIKEY has highest priority over all others.
	t.Run("edda_llm_claude_apikey_highest_priority", func(t *testing.T) {
		t.Setenv("EDDA_LLM_CLAUDE_APIKEY", "sk-ant-edda-llm-highest")
		t.Setenv("ANTHROPIC_API_KEY", "sk-ant-anthropic-override")
		cfg, err := Load(configPath)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.LLM.Claude.APIKey != "sk-ant-edda-llm-highest" {
			t.Fatalf("expected EDDA_LLM_CLAUDE_APIKEY to have highest priority, got %q", cfg.LLM.Claude.APIKey)
		}
	})
}

func TestValidateAcceptsValidProviders(t *testing.T) {
	for _, provider := range []string{"ollama", "claude", "openrouter"} {
		t.Run(provider, func(t *testing.T) {
			cfg := Config{LLM: LLMConfig{Provider: provider}}
			if provider == "ollama" {
				cfg.LLM.Ollama.EmbeddingDimension = 768
			}
			if provider == "claude" {
				cfg.LLM.Claude.APIKey = "sk-ant-test"
			}
			if provider == "openrouter" {
				cfg.LLM.OpenRouter.APIKey = "sk-or-test"
			}
			if err := cfg.Validate(); err != nil {
				t.Fatalf("valid provider %q rejected: %v", provider, err)
			}
		})
	}
}

func TestLoadLLMProfilesAndRoutes(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")

	const fileConfig = `llm:
  provider: openrouter
  openrouter:
    apikey: sk-or-test
    model: default-model
  profiles:
    qwen-fast:
      provider: openrouter
      model: qwen/qwen3-235b-a22b-2507
      contexttokenbudget: 12000
      toolsupport: true
  routes:
    gmturn: qwen-fast
    postturnstate: qwen-fast
    choicefallback: qwen-fast
`
	if err := os.WriteFile(configPath, []byte(fileConfig), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	profile := cfg.LLM.Profiles["qwen-fast"]
	if profile.Model != "qwen/qwen3-235b-a22b-2507" {
		t.Fatalf("unexpected profile model: %q", profile.Model)
	}
	if profile.ContextTokenBudget != 12000 {
		t.Fatalf("unexpected profile context token budget: %d", profile.ContextTokenBudget)
	}
	if !profile.ToolSupport {
		t.Fatal("expected profile tool support")
	}
	if cfg.LLM.Routes.PostTurnState != "qwen-fast" || cfg.LLM.Routes.ChoiceFallback != "qwen-fast" {
		t.Fatalf("unexpected routes: %+v", cfg.LLM.Routes)
	}
}

func TestLoadRejectsUnknownLLMRouteProfile(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")

	const fileConfig = `llm:
  provider: openrouter
  openrouter:
    apikey: sk-or-test
  routes:
    postturnstate: missing-profile
`
	if err := os.WriteFile(configPath, []byte(fileConfig), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected unknown route profile error, got nil")
	}
	if !strings.Contains(err.Error(), "unknown profile") {
		t.Fatalf("expected unknown profile error, got: %v", err)
	}
}

func TestLoadOpenRouterFallbackModelsAndEffortFromEnv(t *testing.T) {
	t.Setenv("EDDA_LLM_PROVIDER", "openrouter")
	t.Setenv("EDDA_LLM_OPENROUTER_APIKEY", "sk-or-test")
	t.Setenv("EDDA_LLM_OPENROUTER_MODEL", "anthropic/claude-sonnet-5.5")
	t.Setenv("EDDA_LLM_OPENROUTER_FALLBACKMODELS", "openai/gpt-6.1-terra,openai/gpt-5.6-terra")
	t.Setenv("EDDA_LLM_OPENROUTER_REASONINGEFFORT", "high")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	got := cfg.LLM.OpenRouter.FallbackModels
	if len(got) != 2 || got[0] != "openai/gpt-6.1-terra" || got[1] != "openai/gpt-5.6-terra" {
		t.Fatalf("FallbackModels = %q", got)
	}
	if cfg.LLM.OpenRouter.ReasoningEffort != "high" {
		t.Fatalf("ReasoningEffort = %q, want high", cfg.LLM.OpenRouter.ReasoningEffort)
	}
}

func TestValidateRejectsUnknownOpenRouterReasoningEffort(t *testing.T) {
	cfg := Config{LLM: LLMConfig{Provider: "openrouter", OpenRouter: OpenRouterConfig{APIKey: "sk-or-test", ReasoningEffort: "hihg"}}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() accepted an unknown reasoning effort")
	}
}

func TestLoadChatGPTFallbackAndClosedRegistrationFromEnv(t *testing.T) {
	t.Setenv("EDDA_LLM_PROVIDER", "openrouter")
	t.Setenv("EDDA_LLM_OPENROUTER_APIKEY", "sk-or-test")
	t.Setenv("EDDA_LLM_FALLBACKPROVIDER", "chatgpt")
	t.Setenv("EDDA_LLM_CHATGPT_CREDENTIALSFILE", "/data/chatgpt-auth.json")
	t.Setenv("EDDA_LLM_CHATGPT_REASONINGEFFORT", "high")
	t.Setenv("EDDA_SERVER_ALLOWREGISTRATION", "false")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.LLM.FallbackProvider != "chatgpt" || cfg.LLM.ChatGPT.CredentialsFile != "/data/chatgpt-auth.json" {
		t.Fatalf("LLM = %+v", cfg.LLM)
	}
	if cfg.LLM.ChatGPT.Model != "gpt-6.1-terra" || cfg.LLM.ChatGPT.ReasoningEffort != "high" {
		t.Fatalf("ChatGPT = %+v", cfg.LLM.ChatGPT)
	}
	if cfg.Server.AllowRegistration {
		t.Fatal("AllowRegistration = true, want false")
	}
}

func TestLoadAllowsRegistrationByDefault(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.Server.AllowRegistration {
		t.Fatal("AllowRegistration = false, want true by default")
	}
}

func TestValidateChatGPTNeedsCredentialsFile(t *testing.T) {
	cfg := Config{LLM: LLMConfig{Provider: "openrouter", FallbackProvider: "chatgpt", OpenRouter: OpenRouterConfig{APIKey: "sk-or-test"}}}
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "credentials file") {
		t.Fatalf("Validate() = %v, want credentials-file error", err)
	}
}
