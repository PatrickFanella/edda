package srd521

import "fmt"

const EquipmentEquip = "equipment-equip"
const EquipmentRemove = "equipment-remove"

// IsEquipment identifies the two explicit, nonmagical fixture mutations.
func IsEquipment(kind string) bool { return kind == EquipmentEquip || kind == EquipmentRemove }

// HeavyArmorTraining comes only from supported class choices, never item stats.
func HeavyArmorTraining(b *Build) bool {
	return b != nil && (b.Class == "fighter" || (b.Class == "cleric" && b.Casting != nil && b.Casting.Choices != nil && b.Casting.Choices.DivineOrder == "Protector"))
}

// EquipmentReceipt records the authoritative build change and elapsed campaign time.
// Source-derived values are adapted from SRD 5.2.1 p. 92; see Attribution.
type EquipmentReceipt struct {
	RulesVersion   string         `json:"rules_version"`
	Kind           string         `json:"kind"`
	ItemID         string         `json:"item_id"`
	Equipment      string         `json:"equipment"`
	EquippedBefore bool           `json:"equipped_before"`
	EquippedAfter  bool           `json:"equipped_after"`
	ACBefore       int            `json:"ac_before"`
	ACAfter        int            `json:"ac_after"`
	TimeBefore     int            `json:"time_before"`
	TimeAfter      int            `json:"time_after"`
	Expired        *Concentration `json:"expired,omitempty"`
	SourceIDs      []string       `json:"source_ids"`
}

// ChangeEquipment mutates only a detached supported build. Inventory/hand
// validation and atomic persistence belong to the shared game/engine boundary.
func ChangeEquipment(b *Build, kind, item string, equipped bool, now int) (*EquipmentReceipt, error) {
	if b == nil || !IsEquipment(kind) || now < 0 || equipped == (kind == EquipmentEquip) || (item != "chain-mail" && item != "shield") {
		return nil, fmt.Errorf("unavailable equipment change")
	}
	if b.Class != "fighter" && b.Class != "cleric" {
		return nil, fmt.Errorf("unsupported armor training")
	}
	if item == "chain-mail" && (!HeavyArmorTraining(b) || b.Abilities["strength"] < 13) {
		return nil, fmt.Errorf("trained chain mail fixture requires Strength 13; reduced-speed armor is unavailable")
	}
	if (item == "chain-mail" && (b.Armor == "chain-mail") != equipped) || (item == "shield" && b.Shield != equipped) {
		return nil, fmt.Errorf("equipment and build disagree")
	}
	seconds := 6 // One Utilize action, accounted as one six-second fixture interval outside combat.
	if item == "chain-mail" {
		seconds = 300
		if kind == EquipmentEquip {
			seconds = 600
		}
	}
	r := &EquipmentReceipt{RulesVersion: "srd-5.2.1", Kind: kind, Equipment: item, EquippedBefore: equipped, EquippedAfter: !equipped, TimeBefore: now, TimeAfter: now + seconds, SourceIDs: []string{prefix + "armor/" + item}}
	if item == "chain-mail" {
		b.Armor = "none"
		if !equipped {
			b.Armor = "chain-mail"
		}
	} else {
		b.Shield = !equipped
	}
	if b.Casting != nil && b.Casting.Concentration != nil && b.Casting.Concentration.ExpiresAt <= r.TimeAfter {
		r.Expired = b.Casting.Concentration
		b.Casting.Concentration = nil
		r.SourceIDs = append(r.SourceIDs, prefix+"rules-glossary/concentration")
	}
	return r, nil
}
