package game

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"github.com/google/uuid"
)

func prepareEquipment(ctx context.Context, database statedb.DBTX, campaign domain.Campaign, pc domain.PlayerCharacter, intent ActionIntent, lock bool) (*StoredAction, error) {
	p, err := PrepareCharacterAction(ctx, campaign, pc, ActionIntent{ActorID: intent.ActorID, Kind: srd521.AbilityCheck, Ability: "strength"})
	if err != nil {
		return nil, err
	}
	p.Intent = intent
	r := &StoredAction{Preview: p}
	if !p.Available {
		return r, nil
	}
	p.Available = false
	p.Modifier = nil
	p.Reason = "Only owned ordinary chain mail or shield fixtures with supported training and free hands are available."
	if intent.Ability != "" || intent.Skill != "" || intent.SlotLevel != 0 {
		return r, nil
	}
	id, err := uuid.Parse(intent.EntryID)
	if err != nil {
		return r, nil
	}
	locking := ""
	if lock {
		locking = " FOR UPDATE"
	}
	// Fingerprint and lock the selected item plus ALL equipped items: a second
	// suit, unknown weapon or changed ownership must invalidate the preview.
	rows, err := database.Query(ctx, `SELECT id,item_type,rarity,properties,equipped,quantity,updated_at FROM items WHERE campaign_id=$1 AND player_character_id=$2 AND (id=$3 OR equipped) ORDER BY id`+locking, campaign.ID, pc.ID, id)
	if err != nil {
		return nil, err
	}
	type item struct {
		ID           uuid.UUID
		Type, Rarity string
		Properties   []byte
		Equipped     bool
		Quantity     int
		Updated      time.Time
	}
	items := []item{}
	for rows.Next() {
		var i item
		if err = rows.Scan(&i.ID, &i.Type, &i.Rarity, &i.Properties, &i.Equipped, &i.Quantity, &i.Updated); err != nil {
			rows.Close()
			return nil, err
		}
		items = append(items, i)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return nil, err
	}
	var selected *item
	hands := 0
	armors := map[string]int{}
	for n := range items {
		i := &items[n]
		var props map[string]any
		if json.Unmarshal(i.Properties, &props) != nil {
			return r, nil
		}
		key, _ := props["srd521_id"].(string)
		if i.ID == id {
			selected = i
			if i.Type != "armor" || i.Rarity != "common" || i.Quantity != 1 || len(props) != 1 || (key != "chain-mail" && key != "shield") {
				return r, nil
			}
		}
		if !i.Equipped {
			continue
		}
		if i.Quantity != 1 || len(props) != 1 || i.Rarity != "common" {
			return r, nil
		}
		switch {
		case i.Type == "armor" && key == "chain-mail":
			armors[key]++
		case i.Type == "armor" && key == "shield":
			armors[key]++
			hands++
		case i.Type == "weapon" && key == "longsword":
			hands++
		default:
			return r, nil
		}
	}
	if selected == nil || hands > 2 || armors["chain-mail"] > 1 || armors["shield"] > 1 {
		return r, nil
	}
	var e struct {
		Build srd521.Build `json:"srd521"`
	}
	if err = json.Unmarshal(pc.Stats, &e); err != nil {
		return nil, err
	}
	if (e.Build.Armor == "chain-mail") != (armors["chain-mail"] == 1) || e.Build.Shield != (armors["shield"] == 1) {
		p.Reason = "Inventory and build disagree; equipment changes require a reconciled fixture."
		return r, nil
	}
	var props map[string]string
	if err = json.Unmarshal(selected.Properties, &props); err != nil {
		return r, nil
	}
	key := props["srd521_id"]
	if intent.Kind == srd521.EquipmentEquip && ((key == "shield" && (hands >= 2 || e.Build.Shield)) || (key == "chain-mail" && e.Build.Armor != "none")) {
		return r, nil
	}
	// A supported caster must retain a free hand for its recorded components.
	// Broader focus/weapon and War Caster arrangements remain unavailable.
	if e.Build.Casting != nil && (hands > 1 || (intent.Kind == srd521.EquipmentEquip && key == "shield" && hands >= 1)) {
		p.Reason = "The supported caster fixture requires a free hand for spell components."
		return r, nil
	}
	now, err := ReadSpellClock(ctx, database, campaign.ID)
	if err != nil {
		return nil, err
	}
	before, err := srd521.Project(pc.Stats, pc.Level, pc.HP, pc.MaxHP, pc.Experience)
	if err != nil {
		return r, nil
	}
	receipt, err := srd521.ChangeEquipment(&e.Build, intent.Kind, key, selected.Equipped, now)
	if err != nil {
		p.Reason = err.Error()
		return r, nil
	}
	stats, err := json.Marshal(map[string]any{"srd521": e.Build})
	if err != nil {
		return nil, err
	}
	after, err := srd521.Project(stats, pc.Level, pc.HP, pc.MaxHP, pc.Experience)
	if err != nil {
		return r, nil
	}
	receipt.ItemID = id.String()
	receipt.ACBefore = before.Defenses["armor-class"].Value
	receipt.ACAfter = after.Defenses["armor-class"].Value
	encoded, err := json.Marshal([]any{p.ExpectedRevision, items, now, intent})
	if err != nil {
		return nil, err
	}
	p.ExpectedRevision = fmt.Sprintf("equipment-v1:%x", sha256.Sum256(encoded))
	r.Equipment = receipt
	r.Build = &e.Build
	p.Available = true
	p.Reason = ""
	p.CostStatus = "fixture-only"
	p.Costs = []ActionCost{{ResourceID: "campaign-seconds", Amount: receipt.TimeAfter - now}}
	p.Options = []string{fmt.Sprintf("%s %s; AC %d to %d; %d seconds of uninterrupted downtime. Shield: one Utilize action, accounted as six seconds outside combat.", intent.Kind, key, receipt.ACBefore, receipt.ACAfter, receipt.TimeAfter-now)}
	p.SourceIDs = receipt.SourceIDs
	return r, nil
}
