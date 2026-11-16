package combat

import (
	"fmt"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
)

// BanditAttack validates the selected named action against symmetric known facts.
// Long range, unseen targets and close ranged attacks need unsupported rulings.
func BanditAttack(profile *srd521.BanditProfile, space *Space) error {
	if profile == nil || space == nil {
		return fmt.Errorf("explicit Bandit and spatial facts required")
	}
	_, action, err := profile.Record()
	if err != nil {
		return err
	}
	if err := space.Validate(); err != nil {
		return err
	}
	if action.Attack.Kind == "melee" {
		return space.Melee()
	}
	if !*space.LineOfSight || space.Cover == "total" || *space.DistanceFeet <= 5 || *space.DistanceFeet > action.Attack.Range {
		return fmt.Errorf("Bandit ranged attack requires visible target beyond 5 feet within normal range")
	}
	return nil
}
