package game

import (
	"context"
	"fmt"

	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
)

// ResolveCharacterCheck is the common web/TUI boundary for the supported SRD
// sheet fixture. Callers provide trusted campaign metadata and adjudicated test
// context, never modifiers supplied by a model or client. It does not mutate
// state or opt a campaign into the legacy turn engine.
func ResolveCharacterCheck(ctx context.Context, version domain.RulesVersion, pc domain.PlayerCharacter, check srd521.Check, difficulty srd521.Difficulty, dice srd521.Dice) (*srd521.CheckResolution, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if version != domain.RulesVersionSRD521 {
		return nil, fmt.Errorf("SRD check requires srd-5.2.1 campaign metadata")
	}
	sheet, err := srd521.Project(pc.Stats, pc.Level, pc.HP, pc.MaxHP, pc.Experience)
	if err != nil {
		return nil, fmt.Errorf("project check character: %w", err)
	}
	return srd521.ResolveCheck(sheet, check, difficulty, dice)
}
