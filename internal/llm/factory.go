package llm

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/config"
)

// LLMProvider is an alias for the provider interface used by the game engine.
type LLMProvider = Provider

const (
	ollamaTagsPath                  = "api/tags"
	ollamaHealthCheckTimeout        = 3 * time.Second
	ollamaHealthCheckErrorBodyLimit = 512
)

// NewLLMProvider constructs the configured LLM provider implementation.
func NewLLMProvider(cfg config.Config) (LLMProvider, error) {
	switch cfg.LLM.Provider {
	case "ollama":
		if err := validateOllamaEndpoint(cfg.LLM.Ollama.Endpoint, cfg.LLM.Ollama.APIKey); err != nil {
			return nil, err
		}
		return NewOllamaClientWithTimeout(cfg.LLM.Ollama.Endpoint, cfg.LLM.Ollama.Model, cfg.LLM.Ollama.RequestTimeout()).
			WithAPIKey(cfg.LLM.Ollama.APIKey), nil
	case "claude":
		if strings.TrimSpace(cfg.LLM.Claude.APIKey) == "" {
			return nil, errors.New("claude provider unavailable: missing api key (set llm.claude.apikey, EDDA_LLM_CLAUDE_APIKEY, or ANTHROPIC_API_KEY)")
		}
		return NewClaudeClient("", cfg.LLM.Claude.APIKey, cfg.LLM.Claude.Model), nil
	default:
		return nil, fmt.Errorf("unsupported llm provider %q (supported: ollama, claude)", cfg.LLM.Provider)
	}
}

func validateOllamaEndpoint(baseURL, apiKey string) error {
	if baseURL == "" {
		baseURL = defaultOllamaBaseURL
	}

	parsed, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return fmt.Errorf("ollama provider unavailable: invalid endpoint %q: %w", baseURL, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("ollama provider unavailable: invalid endpoint scheme %q for %q", parsed.Scheme, baseURL)
	}

	parsed.RawQuery = ""
	parsed.Fragment = ""
	tagsURL, err := url.JoinPath(parsed.String(), ollamaTagsPath)
	if err != nil {
		return fmt.Errorf("ollama provider unavailable: failed to build health-check URL from %q: %w", baseURL, err)
	}
	client := &http.Client{Timeout: ollamaHealthCheckTimeout}

	req, err := http.NewRequest(http.MethodGet, tagsURL, nil)
	if err != nil {
		return fmt.Errorf("ollama provider unavailable: failed to build health-check request: %w", err)
	}
	// /api/tags is a passthrough endpoint on the llama-line broker but still
	// requires the Bearer token. Vanilla ollama ignores the header.
	if key := strings.TrimSpace(apiKey); key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ollama provider unavailable: cannot reach %s: %w", tagsURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, ollamaHealthCheckErrorBodyLimit))
		return fmt.Errorf("ollama provider unavailable: %s returned HTTP %d: %s", tagsURL, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}
