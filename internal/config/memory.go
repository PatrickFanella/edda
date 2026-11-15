package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// ResolveMemoryEmbedding selects and validates the effective memory settings.
// Only an omitted memoryembedding block can inherit legacy Ollama settings.
// Disabled memory needs no endpoint, model, credentials or database validation.
func (c LLMConfig) ResolveMemoryEmbedding() (MemoryEmbeddingConfig, error) {
	var m MemoryEmbeddingConfig
	if c.MemoryEmbedding != nil {
		m = *c.MemoryEmbedding
	} else if c.Provider == "ollama" {
		m = MemoryEmbeddingConfig{
			Enabled: true, Provider: "ollama",
			Endpoint: c.Ollama.EmbeddingEndpoint, Model: c.Ollama.EmbeddingModel,
			Dimension: c.Ollama.EmbeddingDimension, APIKey: c.Ollama.APIKey,
		}
		if m.Endpoint == "" {
			m.Endpoint = c.Ollama.Endpoint
		}
	}
	if !m.Enabled {
		return m, nil
	}
	if m.Provider == "" {
		m.Provider = "ollama"
	}
	if m.Provider != "ollama" {
		return m, fmt.Errorf("llm.memoryembedding.provider %q is unsupported (use ollama for the Ollama-compatible /api/embed API)", m.Provider)
	}
	// Reject credentials and URL suffixes that would be misplaced when /api/embed
	// is appended. Keep errors free of endpoint values, which may contain secrets.
	u, err := url.Parse(m.Endpoint)
	if err != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || strings.TrimSpace(m.Endpoint) != m.Endpoint {
		return m, errors.New("llm.memoryembedding.endpoint must be an absolute HTTP(S) base URL without credentials, query or fragment")
	}
	if strings.HasSuffix(strings.TrimRight(u.Path, "/"), "/api/embed") {
		return m, errors.New("llm.memoryembedding.endpoint must be a base URL, without /api/embed")
	}
	if strings.TrimSpace(m.Model) == "" {
		return m, errors.New("llm.memoryembedding.model is required when semantic memory is enabled")
	}
	if m.Dimension <= 0 {
		return m, errors.New("llm.memoryembedding.dimension must be positive")
	}
	return m, nil
}
