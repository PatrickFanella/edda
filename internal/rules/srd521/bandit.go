package srd521

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

// BanditID is the only stat block executable by the bounded encounter fixture.
const BanditID = "srd-5.2.1:content/monsters/bandit"

// BanditProfile explicitly pins a record revision and one named weapon action.
// It carries no HP or private traits and does not grant general catalogue support.
type BanditProfile struct {
	Version  string `json:"version"`
	ActionID string `json:"action_id"`
}

// BanditRecord returns the validated pinned transcription and its content revision.
func BanditRecord() (Monster, string, error) {
	c, err := LoadMonsters()
	if err != nil {
		return Monster{}, "", err
	}
	for _, m := range c.Monsters {
		if m.ID == BanditID {
			if !m.availableWith(map[string]bool{"monster/stat-block": true, "action/melee-attack": true, "action/ranged-attack": true, "effect/slashing-damage": true, "effect/piercing-damage": true}) || len(m.Traits) != 0 {
				return Monster{}, "", fmt.Errorf("unsupported Bandit mechanics")
			}
			data, err := json.Marshal(m)
			if err != nil {
				return Monster{}, "", err
			}
			return m, fmt.Sprintf("bandit-v1:%x", sha256.Sum256(data)), nil
		}
	}
	return Monster{}, "", fmt.Errorf("missing Bandit record")
}

// Record validates the explicit revision and selected action before any roll.
func (p BanditProfile) Record() (Monster, MonsterAction, error) {
	m, version, err := BanditRecord()
	if err != nil {
		return Monster{}, MonsterAction{}, err
	}
	if p.Version == version {
		for _, a := range m.Actions {
			if a.ID == p.ActionID && (a.ID == "scimitar" || a.ID == "light-crossbow") {
				return m, a, nil
			}
		}
	}
	return Monster{}, MonsterAction{}, fmt.Errorf("unsupported Bandit revision or action")
}

// ResolveBanditAttack rolls one supported weapon attack after geometry validation.
func ResolveBanditAttack(p BanditProfile, armorClass int, dice Dice) (*AttackReceipt, error) {
	m, a, err := p.Record()
	if err != nil {
		return nil, err
	}
	if armorClass < 1 || dice == nil || a.Attack.Damage.Dice != 1 {
		return nil, fmt.Errorf("unsupported attack")
	}
	attack := a.Attack
	r, err := resolveWeaponAttack("srd-5.2.1", a.Name, derived(term("Bandit weapon attack", attack.Bonus, "monsters/bandit")), attack.Damage.Modifier, attack.Damage.Sides, attack.DamageType, armorClass, dice)
	if err == nil {
		r.SourceIDs = []string{m.ID, prefix + "d20-tests/attack-rolls", prefix + "damage/critical-hits"}
	}
	return r, err
}
