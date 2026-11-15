package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveMemoryEmbedding(t *testing.T) {
	legacy := OllamaConfig{
		Endpoint: "http://localhost:11434", EmbeddingEndpoint: "https://embed.example.test/broker",
		EmbeddingModel: "legacy-model", EmbeddingDimension: 768, APIKey: "legacy-key",
	}
	explicit := MemoryEmbeddingConfig{Enabled: true, Endpoint: "https://embed.example.test", Model: "independent-model", Dimension: 768, APIKey: "embed-key"}
	for _, provider := range []string{"ollama", "claude", "openrouter", "chatgpt", "claudecode"} {
		t.Run(provider, func(t *testing.T) {
			cfg := LLMConfig{Provider: provider, Ollama: legacy}
			got, err := cfg.ResolveMemoryEmbedding()
			if err != nil {
				t.Fatal(err)
			}
			if provider == "ollama" {
				if !got.Enabled || got.Endpoint != legacy.EmbeddingEndpoint || got.Model != legacy.EmbeddingModel || got.Dimension != legacy.EmbeddingDimension || got.APIKey != legacy.APIKey {
					t.Fatal("legacy Ollama settings were not selected")
				}
				cfg.Ollama.EmbeddingEndpoint = ""
				got, err = cfg.ResolveMemoryEmbedding()
				if err != nil || got.Endpoint != legacy.Endpoint {
					t.Fatalf("legacy endpoint fallback: %v", err)
				}
			} else if got.Enabled {
				t.Fatal("non-Ollama GM unexpectedly inherited legacy embedding settings")
			}
			cfg.MemoryEmbedding = &explicit
			got, err = cfg.ResolveMemoryEmbedding()
			if err != nil || !got.Enabled || got.Provider != "ollama" || got.Endpoint != explicit.Endpoint || got.Model != explicit.Model || got.Dimension != explicit.Dimension || got.APIKey != explicit.APIKey {
				t.Fatalf("explicit settings did not take precedence: %v", err)
			}
			cfg.MemoryEmbedding = &MemoryEmbeddingConfig{Enabled: false}
			got, err = cfg.ResolveMemoryEmbedding()
			if err != nil || got.Enabled {
				t.Fatalf("explicit disable ignored: %v", err)
			}
		})
	}
}

func TestValidateMemoryEmbedding(t *testing.T) {
	for _, tt := range []struct {
		name, want string
		change     func(*MemoryEmbeddingConfig)
	}{
		{"valid", "", func(*MemoryEmbeddingConfig) {}},
		{"provider", "unsupported", func(m *MemoryEmbeddingConfig) { m.Provider = "openai" }},
		{"missing endpoint", "endpoint", func(m *MemoryEmbeddingConfig) { m.Endpoint = "" }},
		{"relative endpoint", "endpoint", func(m *MemoryEmbeddingConfig) { m.Endpoint = "/embed" }},
		{"missing host", "endpoint", func(m *MemoryEmbeddingConfig) { m.Endpoint = "http:///embed" }},
		{"unsupported scheme", "endpoint", func(m *MemoryEmbeddingConfig) { m.Endpoint = "ftp://example.test" }},
		{"malformed URL", "endpoint", func(m *MemoryEmbeddingConfig) { m.Endpoint = "http://[" }},
		{"credentials", "endpoint", func(m *MemoryEmbeddingConfig) { m.Endpoint = "http://user:private@example.test" }},
		{"query", "endpoint", func(m *MemoryEmbeddingConfig) { m.Endpoint = "http://example.test?token=private" }},
		{"fragment", "endpoint", func(m *MemoryEmbeddingConfig) { m.Endpoint = "http://example.test#fragment" }},
		{"full API path", "base URL", func(m *MemoryEmbeddingConfig) { m.Endpoint = "http://example.test/api/embed/" }},
		{"empty model", "model", func(m *MemoryEmbeddingConfig) { m.Model = "" }},
		{"blank model", "model", func(m *MemoryEmbeddingConfig) { m.Model = "  " }},
		{"zero dimension", "dimension", func(m *MemoryEmbeddingConfig) { m.Dimension = 0 }},
		{"negative dimension", "dimension", func(m *MemoryEmbeddingConfig) { m.Dimension = -1 }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := MemoryEmbeddingConfig{Enabled: true, Endpoint: "https://embed.example.test", Model: "test-embed", Dimension: 768}
			tt.change(&m)
			cfg := Config{LLM: LLMConfig{Provider: "claude", Claude: ClaudeConfig{APIKey: "test-key"}, MemoryEmbedding: &m}}
			err := cfg.Validate()
			if tt.want == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Validate() = %v, want %q", err, tt.want)
			}
			if err != nil && strings.Contains(err.Error(), "private") {
				t.Fatal("validation error exposed URL credentials")
			}
		})
	}
}

func TestLoadMemoryEmbedding(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(`llm:
  provider: ollama
  memoryembedding:
    enabled: false
`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.LLM.MemoryEmbedding == nil || cfg.LLM.MemoryEmbedding.Enabled {
		t.Fatal("YAML explicit false must remain distinct from an omitted block")
	}
	t.Setenv("EDDA_LLM_MEMORYEMBEDDING_ENABLED", "true")
	t.Setenv("EDDA_LLM_MEMORYEMBEDDING_PROVIDER", "ollama")
	t.Setenv("EDDA_LLM_MEMORYEMBEDDING_ENDPOINT", "https://embed.example.test")
	t.Setenv("EDDA_LLM_MEMORYEMBEDDING_MODEL", "test-embed")
	t.Setenv("EDDA_LLM_MEMORYEMBEDDING_DIMENSION", "768")
	t.Setenv("EDDA_LLM_MEMORYEMBEDDING_APIKEY", "test-key")
	cfg, err = Load(path)
	if err != nil {
		t.Fatal(err)
	}
	m, err := cfg.LLM.ResolveMemoryEmbedding()
	if err != nil || !m.Enabled || m.Provider != "ollama" || m.Endpoint != "https://embed.example.test" || m.Model != "test-embed" || m.Dimension != 768 || m.APIKey != "test-key" {
		t.Fatalf("environment did not override YAML settings: %v", err)
	}
	t.Setenv("EDDA_LLM_MEMORYEMBEDDING_ENABLED", "false")
	t.Setenv("EDDA_LLM_MEMORYEMBEDDING_ENDPOINT", "")
	t.Setenv("EDDA_LLM_MEMORYEMBEDDING_MODEL", "")
	t.Setenv("EDDA_LLM_MEMORYEMBEDDING_DIMENSION", "0")
	cfg, err = Load("")
	if err != nil {
		t.Fatal(err)
	}
	m, err = cfg.LLM.ResolveMemoryEmbedding()
	if err != nil || m.Enabled {
		t.Fatalf("environment explicit false must override legacy Ollama defaults: %v", err)
	}
}

func TestLoadLegacyMemoryDefaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.LLM.MemoryEmbedding != nil {
		t.Fatal("absent memoryembedding configuration must stay absent after loading defaults")
	}
	m, err := cfg.LLM.ResolveMemoryEmbedding()
	if err != nil || !m.Enabled || m.Model != "nomic-embed-text" || m.Dimension != 768 || m.Endpoint != cfg.LLM.Ollama.Endpoint {
		t.Fatalf("legacy memory defaults were not selected: %v", err)
	}
}
