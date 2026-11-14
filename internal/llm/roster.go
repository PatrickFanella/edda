package llm

import (
	"fmt"
	"strings"

	"git.subcult.tv/subculture-collective/edda/internal/config"
)

// Flow names a high-level LLM use case that can be routed to a profile.
type Flow string

const (
	FlowDefault        Flow = "default"
	FlowGMTurn         Flow = "gmturn"
	FlowPostTurnState  Flow = "postturnstate"
	FlowChoiceFallback Flow = "choicefallback"
	FlowSummary        Flow = "summary"
	FlowWorldgen       Flow = "worldgen"
)

// RoutedProvider is a provider plus metadata about the route/profile that
// produced it. The Provider interface remains the only thing lower-level code
// needs in order to make calls.
type RoutedProvider struct {
	ProfileName        string
	Flow               Flow
	ProviderName       string
	Model              string
	ContextTokenBudget int
	ToolSupport        bool
	Provider           Provider
}

// Roster owns the configured LLM providers for the process. Missing routes
// fall back to the default provider, preserving the old single-model behavior.
type Roster struct {
	defaultProvider RoutedProvider
	routes          map[Flow]RoutedProvider
}

// NewRoster constructs all configured LLM provider profiles and their routes.
func NewRoster(cfg config.Config) (*Roster, error) {
	defaultProvider, err := NewLLMProvider(cfg)
	if err != nil {
		return nil, err
	}
	defaultRouted := RoutedProvider{
		ProfileName:        "default",
		Flow:               FlowDefault,
		ProviderName:       cfg.LLM.Provider,
		Model:              activeModel(cfg.LLM),
		ContextTokenBudget: cfg.LLM.ContextTokenBudget(),
		ToolSupport:        true,
		Provider:           defaultProvider,
	}

	profileProviders := map[string]RoutedProvider{}
	for name, profile := range cfg.LLM.Profiles {
		profileCfg := applyProfile(cfg, profile)
		provider, err := NewLLMProvider(profileCfg)
		if err != nil {
			return nil, fmt.Errorf("create llm profile %q: %w", name, err)
		}
		profileProviders[name] = RoutedProvider{
			ProfileName:        name,
			ProviderName:       profileCfg.LLM.Provider,
			Model:              activeModel(profileCfg.LLM),
			ContextTokenBudget: profileCfg.LLM.ContextTokenBudget(),
			ToolSupport:        profile.ToolSupport,
			Provider:           provider,
		}
	}

	roster := &Roster{defaultProvider: defaultRouted, routes: map[Flow]RoutedProvider{}}
	for flow, profileName := range configuredRoutes(cfg.LLM.Routes) {
		routed, ok := profileProviders[profileName]
		if !ok {
			return nil, fmt.Errorf("llm route %q references unknown profile %q", flow, profileName)
		}
		routed.Flow = flow
		roster.routes[flow] = routed
	}

	return roster, nil
}

// Provider returns the provider for a flow, falling back to default.
func (r *Roster) Provider(flow Flow) Provider {
	return r.Info(flow).Provider
}

// Info returns provider metadata for a flow, falling back to default.
func (r *Roster) Info(flow Flow) RoutedProvider {
	if r == nil {
		return RoutedProvider{}
	}
	if routed, ok := r.routes[flow]; ok && routed.Provider != nil {
		return routed
	}
	routed := r.defaultProvider
	routed.Flow = flow
	return routed
}

func configuredRoutes(routes config.LLMRoutesConfig) map[Flow]string {
	out := map[Flow]string{}
	if routes.GMTurn != "" {
		out[FlowGMTurn] = routes.GMTurn
	}
	if routes.PostTurnState != "" {
		out[FlowPostTurnState] = routes.PostTurnState
	}
	if routes.ChoiceFallback != "" {
		out[FlowChoiceFallback] = routes.ChoiceFallback
	}
	if routes.Summary != "" {
		out[FlowSummary] = routes.Summary
	}
	if routes.Worldgen != "" {
		out[FlowWorldgen] = routes.Worldgen
	}
	return out
}

func applyProfile(cfg config.Config, profile config.LLMProfileConfig) config.Config {
	provider := strings.TrimSpace(profile.Provider)
	if provider == "" {
		provider = cfg.LLM.Provider
	}
	cfg.LLM.Provider = provider

	if model := strings.TrimSpace(profile.Model); model != "" {
		switch provider {
		case "claude":
			cfg.LLM.Claude.Model = model
		case "openrouter":
			cfg.LLM.OpenRouter.Model = model
		case "chatgpt":
			cfg.LLM.ChatGPT.Model = model
		default:
			cfg.LLM.Ollama.Model = model
		}
	}
	if profile.Skills && provider == "claudecode" {
		cfg.LLM.ClaudeCode.Skills = true
	}
	if budget := profile.ContextTokenBudget; budget > 0 {
		switch provider {
		case "claude":
			cfg.LLM.Claude.ContextTokenBudget = budget
		case "openrouter":
			cfg.LLM.OpenRouter.ContextTokenBudget = budget
		case "chatgpt":
			cfg.LLM.ChatGPT.ContextTokenBudget = budget
		case "claudecode":
			cfg.LLM.ClaudeCode.ContextTokenBudget = budget
		default:
			cfg.LLM.Ollama.ContextTokenBudget = budget
		}
	}
	return cfg
}

func activeModel(cfg config.LLMConfig) string {
	switch cfg.Provider {
	case "claude":
		return cfg.Claude.Model
	case "openrouter":
		return cfg.OpenRouter.Model
	case "chatgpt":
		return cfg.ChatGPT.Model
	case "claudecode":
		return "claude-code"
	default:
		return cfg.Ollama.Model
	}
}
