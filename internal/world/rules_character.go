package world

import (
	"fmt"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	"strings"
)

// RulesChoices selects a fixed catalogue build; it cannot grant custom mechanics.
type RulesChoices struct {
	Origin     string `json:"origin"`
	Class      string `json:"class"`
	Background string `json:"background"`
}

// ValidateRulesCharacter is shared by guided creation, interview and persistence.
// Existing guided clients can identify the fixed build through concept/background.
func ValidateRulesCharacter(p *CharacterProfile) (*srd521.SupportedBuild, error) {
	if p == nil || strings.TrimSpace(p.Name) == "" {
		return nil, fmt.Errorf("character name is required")
	}
	build, err := srd521.SupportedBuildForConcept(p.Concept, p.Background)
	if err != nil {
		return nil, err
	}
	if c := p.RulesChoices; c != nil {
		selected, err := srd521.SupportedBuildForConcept(c.Origin+" "+c.Class, c.Background)
		if err != nil {
			return nil, err
		}
		if selected.Class != build.Class {
			return nil, fmt.Errorf("structured rules choices must match the displayed concept and background")
		}
	}
	return build, nil
}
