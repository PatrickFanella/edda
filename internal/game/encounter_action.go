package game

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"git.subcult.tv/subculture-collective/edda/internal/combat"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"github.com/google/uuid"
)

// ReadEncounter restores the latest durable fixture, including completed state.
// The indexed campaign history query has no recent-turn cutoff.
func ReadEncounter(ctx context.Context, database statedb.DBTX, campaignID uuid.UUID) (*combat.Encounter, error) {
	var data []byte
	err := database.QueryRow(ctx, `SELECT COALESCE((SELECT call->'Result'->'encounter' FROM session_logs l CROSS JOIN LATERAL jsonb_array_elements(COALESCE(l.tool_calls,'[]'::jsonb)) WITH ORDINALITY AS t(call,position) WHERE l.campaign_id=$1 AND call->'Result' ? 'encounter' ORDER BY l.turn_number DESC,position DESC LIMIT 1),'null'::jsonb)`, dbutil.ToPgtype(campaignID)).Scan(&data)
	if err != nil {
		return nil, err
	}
	var encounter *combat.Encounter
	if err := json.Unmarshal(data, &encounter); err != nil {
		return nil, err
	}
	if encounter != nil && (len(encounter.Actors) != 2 || encounter.Turn < 0 || encounter.Turn >= 2 || encounter.Round < 1) {
		return nil, fmt.Errorf("invalid saved encounter")
	}
	if encounter != nil && encounter.Reaction == nil {
		encounter.Reaction = &combat.ReactionState{Available: true}
	}
	return encounter, nil
}

// PrepareStoredAction adds a durable encounter fingerprint and enforces turn
// ownership on the same path used by confirmation. Preparation never rolls.
func PrepareStoredAction(ctx context.Context, database statedb.DBTX, campaign domain.Campaign, pc domain.PlayerCharacter, intent ActionIntent, lock bool) (*StoredAction, error) {
	base := intent
	switch intent.Kind {
	case combat.EncounterStart, combat.EncounterReact:
		base.Kind = srd521.MartialAttack
		if intent.Kind == combat.EncounterStart && intent.EntryID == "" {
			if sheet, err := srd521.Project(pc.Stats, pc.Level, pc.HP, pc.MaxHP, pc.Experience); err == nil && sheet.Spellcasting != nil {
				base = ActionIntent{ActorID: intent.ActorID, Kind: srd521.AbilityCheck, Ability: "wisdom"}
			}
		}
	case combat.EncounterEndTurn, combat.EncounterMove, combat.EncounterDash, combat.EncounterApproach, combat.EncounterRetreat, combat.EncounterDecline:
		base = ActionIntent{ActorID: intent.ActorID, Kind: srd521.AbilityCheck, Ability: "strength"}
	}
	result, err := prepareStoredActionBase(ctx, database, campaign, pc, base, lock)
	if err != nil {
		return nil, err
	}
	if intent.Kind == combat.EncounterStart && base.Kind == srd521.AbilityCheck {
		if intent.Ability != "" || intent.Skill != "" || intent.SlotLevel != 0 {
			result.Preview.Available = false
		} else {
			result.Targets, err = readMartialTargets(ctx, database, campaign, pc, lock)
			if err != nil {
				return nil, err
			}
		}
	}
	p := result.Preview
	result.Clock, err = ReadSpellClock(ctx, database, campaign.ID)
	if err != nil {
		return nil, err
	}
	encodedClock, err := json.Marshal([]any{p.ExpectedRevision, result.Clock})
	if err != nil {
		return nil, err
	}
	if srd521.IsLifecycle(intent.Kind) || intent.Kind == srd521.SpellCast || intent.Kind == combat.EncounterEndTurn || intent.Kind == combat.EncounterStart || intent.Kind == combat.EncounterReact || intent.Kind == combat.EncounterDecline || intent.Kind == combat.EncounterMove || intent.Kind == combat.EncounterDash || intent.Kind == combat.EncounterApproach || intent.Kind == combat.EncounterRetreat {
		p.ExpectedRevision = fmt.Sprintf("clock-v1:%x", sha256.Sum256(encodedClock))
	}
	if sheet, err := srd521.Project(pc.Stats, pc.Level, pc.HP, pc.MaxHP, pc.Experience); err == nil && sheet.Spellcasting != nil {
		var envelope struct {
			Build srd521.Build `json:"srd521"`
		}
		if err := json.Unmarshal(pc.Stats, &envelope); err != nil {
			return nil, err
		}
		c := envelope.Build.Casting
		if intent.Kind == srd521.LongRest && c.LastLongRest != nil && result.Clock-*c.LastLongRest < 16*3600 {
			p.Available = false
			p.Reason = "Wait 16 hours after the previous Long Rest before starting another."
		}
	}

	defer func() {
		if p.Encounter != nil {
			p.Encounter = p.Encounter.Public()
		}
	}()
	if base != intent {
		p.Intent = intent
	}
	encounter, err := ReadEncounter(ctx, database, campaign.ID)
	if err != nil {
		return nil, err
	}
	result.Encounter = encounter
	p.Encounter = encounter
	if encounter != nil {
		encoded, err := json.Marshal([]any{p.ExpectedRevision, encounter})
		if err != nil {
			return nil, err
		}
		p.ExpectedRevision = fmt.Sprintf("encounter-v1:%x", sha256.Sum256(encoded))
	}
	unavailable := func(reason string) (*StoredAction, error) { p.Available = false; p.Reason = reason; return result, nil }
	active := encounter != nil && encounter.Status == "active"
	// The existing two-actor encounter has no allied participant slice.
	if active && intent.Kind == srd521.SpellCast && intent.EntryID == "cure-wounds" {
		result.HealingTargets = nil
		p.Targets = []ActionTarget{{ID: pc.ID, Label: pc.Name + " (self; touch)"}}
		p.Options = []string{"Cure Wounds; self only in the two-actor encounter; verbal and somatic components; 2d8 + Wisdom capped at maximum HP; one level 1 slot; Magic action."}
	}
	if active && intent.Kind == srd521.ConsumableUse {
		result.HealingTargets = nil
		p.Targets = []ActionTarget{{ID: pc.ID, Label: pc.Name + " (self)"}}
		p.Options = []string{fmt.Sprintf("Quantity: %d; regain 2d4 + 2 HP, capped at %d (current %d); self-use only in the two-actor encounter", result.Quantity, pc.MaxHP, pc.HP)}
	}
	// Revalidate membership on every encounter action, including NPC resolution.
	// Never send stale encounter actors/stats when a target is no longer public.
	if encounter != nil {
		members, err := readMartialTargets(ctx, database, campaign, pc, lock)
		if err != nil {
			return nil, err
		}
		var member *MartialTarget
		for i := range members {
			for _, actor := range encounter.Actors {
				if !actor.Player && actor.ID == members[i].ID {
					member = &members[i]
				}
			}
		}
		if member == nil && active {
			p.Encounter = nil
			return unavailable("The encounter target is no longer available; clarify scene membership.")
		}
		if member == nil {
			p.Encounter = nil
		}
		if member != nil && active {
			if (encounter.Bandit == nil) != (member.Bandit == nil) ||
				(encounter.Bandit != nil && member.Bandit != nil && *encounter.Bandit != *member.Bandit) {
				p.Encounter = nil
				return unavailable("The stat-block profile changed; the encounter requires adjudication.")
			}
		}
		if member != nil && (encounter.Space == nil || encounter.Space.Revision != member.Space.Revision) {
			encounter.Space = &member.Space
		}
		// Include membership, current HP and new rulings in stale-preview detection.
		encoded, err := json.Marshal([]any{p.ExpectedRevision, encounter, members})
		if err != nil {
			return nil, err
		}
		p.ExpectedRevision = fmt.Sprintf("space-v1:%x", sha256.Sum256(encoded))
	}
	if !p.Available {
		return result, nil
	}
	if intent.Kind == combat.EncounterStart || intent.Kind == srd521.MartialAttack || intent.Kind == combat.EncounterReact {
		legal := []MartialTarget{}
		p.Targets = []ActionTarget{}
		for _, target := range result.Targets {
			if active {
				belongs := false
				for _, actor := range encounter.Actors {
					if !actor.Player && actor.ID == target.ID {
						belongs = true
					}
				}
				if !belongs {
					continue
				}
				target.Space = *encounter.Space
			}
			option := ActionTarget{ID: target.ID, Label: target.Name, DistanceFeet: target.Space.DistanceFeet, ReachFeet: 5}
			if intent.Kind == combat.EncounterStart && target.Bandit != nil {
				_, action, err := target.Bandit.Record()
				if err != nil {
					return unavailable(err.Error())
				}
				if action.Attack.Kind == "ranged" {
					option.ReachFeet = action.Attack.Range
				}
			}
			var spatialErr error
			if intent.Kind == combat.EncounterStart {
				spatialErr = target.Space.Validate()
				if target.Bandit != nil {
					spatialErr = combat.BanditAttack(target.Bandit, &target.Space)
				}
			} else {
				spatialErr = target.Space.Melee()
			}
			if spatialErr != nil {
				option.UnavailableReason = spatialErr.Error()
			} else {
				legal = append(legal, target)
			}
			p.Targets = append(p.Targets, option)
		}
		result.Targets = legal
		if len(legal) == 0 {
			return unavailable("No legal target. Clarify unresolved scene facts or choose a target within reach.")
		}
	}
	if intent.Kind == combat.EncounterStart {
		if active {
			return unavailable("An encounter is already active.")
		}
		p.Options = []string{"Two-actor encounter; tied initiative: player first. Confirm to roll initiative."}
		for _, target := range result.Targets {
			if target.Bandit != nil {
				m, a, err := target.Bandit.Record()
				if err != nil {
					return unavailable(err.Error())
				}
				p.Options = append(p.Options, fmt.Sprintf("%s: Bandit AC %d; initiative %+d; speed %d feet; selected %s. %s", target.Name, m.AC, m.Initiative.Modifier, m.Speeds["walk"], a.Name, a.Text))
				for _, action := range m.Actions {
					profile := *target.Bandit
					profile.ActionID = action.ID
					status := "Legal: "
					if err := combat.BanditAttack(&profile, &target.Space); err != nil {
						status = "Unavailable: " + err.Error() + " "
					}
					p.Options = append(p.Options, status+action.Text)
				}
			} else {
				p.Options = append(p.Options, target.Name+": ordinary fixture; NPC Unarmed Strike +2, 1 bludgeoning; speed 30 feet.")
			}
		}
		for _, target := range result.Targets {
			if target.Bandit != nil {
				p.SourceIDs = append(p.SourceIDs, srd521.BanditID)
				break
			}
		}
		p.Costs = []ActionCost{}
		p.CostStatus = "encounter"
		return result, nil
	}
	if !active {
		switch intent.Kind {
		case combat.EncounterEndTurn, combat.EncounterMove, combat.EncounterDash, combat.EncounterApproach, combat.EncounterRetreat, combat.EncounterDecline, combat.EncounterReact:
			return unavailable("No supported encounter is active.")
		}
		return result, nil
	}
	if encounter.Reaction.Pending != nil {
		if intent.Kind != combat.EncounterReact && intent.Kind != combat.EncounterDecline {
			return unavailable("Choose or decline the pending reaction before continuing.")
		}
		trigger := encounter.Reaction.Pending
		if intent.Ability != "" || intent.Skill != "" || intent.SlotLevel != 0 || encounter.Active().Player || encounter.Active().ID != trigger.TargetID {
			return unavailable("This reaction trigger is unavailable.")
		}
		p.Options = []string{"Opportunity attack before the visible foe leaves your reach. Declining spends no reaction. The NPC withdrawal then resumes."}
		p.CostStatus = "encounter"
		p.SourceIDs = append(p.SourceIDs, "srd-5.2.1:content/combat/opportunity-attacks", "srd-5.2.1:content/combat/reactions")
		if intent.Kind == combat.EncounterDecline {
			if intent.EntryID != "" {
				return unavailable("Choose the supported reaction response.")
			}
			p.Modifier = nil
			p.Targets = []ActionTarget{}
			p.Costs = []ActionCost{}
			return result, nil
		}
		if intent.EntryID != trigger.WeaponID.String() {
			return unavailable("The trigger requires its prepared equipped longsword.")
		}
		copy := *encounter
		state := *encounter.Reaction
		copy.Reaction = &state
		if err := copy.SpendReaction(); err != nil {
			return unavailable(err.Error())
		}
		p.Costs = []ActionCost{{ResourceID: "reaction", Amount: 1}}
		return result, nil
	}
	if intent.Kind == combat.EncounterReact || intent.Kind == combat.EncounterDecline {
		return unavailable("No eligible reaction is pending.")
	}
	if encounter.Active().ID != pc.ID {
		return unavailable("Wait for the active actor's turn.")
	}
	// Do not pretend arbitrary checks are encounter actions in this slice.
	switch intent.Kind {
	case srd521.MartialAttack:

	case srd521.EndConcentration, srd521.SpellCast, srd521.ConsumableUse:
	case combat.EncounterMove, combat.EncounterDash, combat.EncounterEndTurn, combat.EncounterApproach, combat.EncounterRetreat, combat.EncounterDecline:
		if intent.EntryID != "" || intent.Ability != "" || intent.Skill != "" || intent.SlotLevel != 0 {
			return unavailable("Choose a supported encounter action.")
		}
		p.Modifier = nil
		p.Targets = []ActionTarget{}
		p.Costs = []ActionCost{}
		p.Options = []string{"End turn resolves the NPC turn through saved dice mechanics."}
		if intent.Kind == combat.EncounterMove || intent.Kind == combat.EncounterApproach || intent.Kind == combat.EncounterRetreat {
			p.Costs = []ActionCost{{ResourceID: "movement-feet", Amount: 5}}
			p.Options = []string{"Spend 5 feet of movement without changing distance."}
			if intent.Kind != combat.EncounterMove {
				p.Options = []string{"Change the declared exact distance by 5 feet on a clear path."}
			}
		}
		if intent.Kind == combat.EncounterDash {
			p.Costs = []ActionCost{{ResourceID: "action", Amount: 1}}
			p.Options = []string{"Dash adds 30 feet of movement this turn."}
		}
	default:
		return unavailable("Use a supported encounter action or end your turn.")
	}
	if intent.Kind == combat.EncounterEndTurn && encounter.Bandit != nil {
		if err := combat.BanditAttack(encounter.Bandit, encounter.Space); err != nil {
			return unavailable(err.Error())
		}
		_, a, err := encounter.Bandit.Record()
		if err != nil {
			return unavailable(err.Error())
		}
		p.Options = []string{"End turn resolves one Bandit " + a.Text}
	}
	if intent.Kind == combat.EncounterEndTurn {
		if err := encounter.Space.Validate(); err != nil {
			return unavailable(err.Error())
		}
	}
	// Spend only on a copy so previews remain read-only.
	copy := *encounter
	if intent.Kind == srd521.SpellCast {
		if err := copy.SpendSpell(pc.ID, intent.EntryID == "shield-of-faith"); err != nil {
			return unavailable(err.Error())
		}
	} else if intent.Kind != srd521.EndConcentration {
		if err := copy.Spend(pc.ID, intent.Kind); err != nil {
			return unavailable(err.Error())
		}
	}

	p.CostStatus = "encounter"
	p.SourceIDs = append(p.SourceIDs, "srd-5.2.1:content/combat/initiative", "srd-5.2.1:content/combat/your-turn")
	return result, nil
}
