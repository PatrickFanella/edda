package srd521

import "fmt"

// ConsumableUse is the basic Potion of Healing slice.
const ConsumableUse = "consumable-use"

// HealingReceipt records dice, capped HP and the exact inventory cost.
type HealingReceipt struct {
	RulesVersion   string   `json:"rules_version"`
	Kind           string   `json:"kind"`
	ItemID         string   `json:"item_id"`
	Item           string   `json:"item"`
	TargetID       string   `json:"target_id"`
	Target         string   `json:"target"`
	Dice           []int    `json:"dice"`
	Bonus          int      `json:"bonus"`
	RolledHealing  int      `json:"rolled_healing"`
	Healing        int      `json:"healing"`
	HPBefore       int      `json:"hp_before"`
	HPAfter        int      `json:"hp_after"`
	QuantityBefore int      `json:"quantity_before"`
	QuantityAfter  int      `json:"quantity_after"`
	SourceIDs      []string `json:"source_ids"`
}

// ResolveHealingPotion implements only basic 2d4+2 healing, capped at Max HP.
// Quantity is revalidated and decremented by the transactional engine caller.
func ResolveHealingPotion(hp, maxHP int, dice Dice) (*HealingReceipt, error) {
	if hp <= 0 || maxHP < hp || dice == nil {
		return nil, fmt.Errorf("conscious target and dice required")
	}
	r := &HealingReceipt{RulesVersion: "srd-5.2.1", Kind: ConsumableUse, Item: "Potion of Healing", Bonus: 2, HPBefore: hp, SourceIDs: []string{prefix + "items/potion-of-healing", prefix + "damage/healing"}}
	r.RolledHealing = 2
	for i := 0; i < 2; i++ {
		v := dice.Intn(4)
		if v < 0 || v >= 4 {
			return nil, fmt.Errorf("invalid d4")
		}
		r.Dice = append(r.Dice, v+1)
		r.RolledHealing += v + 1
	}
	r.HPAfter = min(maxHP, hp+r.RolledHealing)
	r.Healing = r.HPAfter - hp
	return r, nil
}
