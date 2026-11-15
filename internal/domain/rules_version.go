package domain

import (
	"errors"
	"fmt"
)

// RulesVersion identifies mechanics independently of presentation/detail mode.
type RulesVersion string

const (
	RulesVersionLegacy RulesVersion = "legacy"
	RulesVersionSRD521 RulesVersion = "srd-5.2.1"
)

// ErrLegacyRulesMutation means the caller only implements the legacy engine.
var ErrLegacyRulesMutation = errors.New("legacy gameplay cannot mutate a rules-versioned campaign; character conversion is not supported")

// NormalizeRulesVersion defaults omitted metadata to legacy without conversion.
func NormalizeRulesVersion(value RulesVersion) (RulesVersion, error) {
	switch value {
	case "", RulesVersionLegacy:
		return RulesVersionLegacy, nil
	case RulesVersionSRD521:
		return value, nil
	default:
		return "", fmt.Errorf("unsupported rules_version %q", value)
	}
}

// RequireLegacyRules refuses non-legacy campaigns, including unknown versions.
func RequireLegacyRules(value RulesVersion) error {
	if value == "" || value == RulesVersionLegacy {
		return nil
	}
	return fmt.Errorf("%w (rules_version=%s)", ErrLegacyRulesMutation, value)
}
