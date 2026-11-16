package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"

	"git.subcult.tv/subculture-collective/edda/internal/combat"
	"git.subcult.tv/subculture-collective/edda/internal/db"
	"git.subcult.tv/subculture-collective/edda/internal/game"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// PreparedAction confirms supported checks and explicitly prepared fixture actions.
// No modifiers, DC, advantage, resource amounts or disclosure flags are accepted.
type PreparedAction struct {
	ActionID         uuid.UUID         `json:"action_id"`
	ExpectedRevision string            `json:"expected_revision"`
	Intent           game.ActionIntent `json:"intent"`
	TargetID         uuid.UUID         `json:"target_id,omitempty"`
}

// ErrActionRefresh requires a fresh preview before another confirmation.
var ErrActionRefresh = errors.New("action changed or unavailable; refresh its preview")

type actionDice struct{}

func (actionDice) Intn(n int) int { return rand.IntN(n) }

// WithActionDice supplies deterministic dice for mechanical fixtures.
func WithActionDice(dice srd521.Dice) Option { return func(e *Engine) { e.actionDice = dice } }

// ConfirmAction revalidates, rolls and records a check and history in one short
// transaction, under the same campaign exclusion used by narrative turns.
// Replays consult the receipt before revision validation and never roll again.
func (e *Engine) ConfirmAction(ctx context.Context, campaignID uuid.UUID, request PreparedAction) (*TurnResult, error) {
	if e.turnDB == nil || request.ActionID == uuid.Nil || request.ExpectedRevision == "" {
		return nil, ErrActionRefresh
	}
	release, err := e.acquireTurn(ctx, campaignID)
	if err != nil {
		return nil, err
	}
	defer release()
	tx, err := e.turnDB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackTurnTx(tx)
	identity, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	input := "Prepared action: " + string(identity)
	var savedCampaign uuid.UUID
	var savedInput, status string
	var data []byte
	err = tx.QueryRow(ctx, `SELECT campaign_id,player_input,status,final_result FROM turn_attempts WHERE id=$1`, db.ToPgUUID(request.ActionID)).Scan(&savedCampaign, &savedInput, &status, &data)
	if err == nil {
		if savedCampaign != campaignID || savedInput != input {
			return nil, ErrActionIDConflict
		}
		if status != "completed" {
			return nil, ErrTurnRecoveryRequired
		}
		var result TurnResult
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	var pending bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM turn_attempts WHERE campaign_id=$1 AND status IN ('pending_mechanics','pending_narration'))`, db.ToPgUUID(campaignID)).Scan(&pending); err != nil {
		return nil, err
	}
	if pending {
		return nil, ErrTurnRecoveryRequired
	}
	campaign, pc, err := game.ReadActionSnapshot(ctx, tx, campaignID, request.Intent.ActorID, uuid.Nil, true)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrActionRefresh
	}
	if err != nil {
		return nil, err
	}
	prepared, err := game.PrepareStoredAction(ctx, tx, campaign, pc, request.Intent, true)
	if err != nil {
		return nil, err
	}
	preview := prepared.Preview
	if !preview.Available || preview.ExpectedRevision != request.ExpectedRevision {
		return nil, ErrActionRefresh
	}
	// Only prepared targets can cross the confirmation boundary.
	if request.Intent.Kind == srd521.SpellCast || request.Intent.Kind == srd521.SavingThrow || request.Intent.Kind == srd521.ConsumableUse {
		if request.TargetID != pc.ID && !((request.Intent.Kind == srd521.ConsumableUse || (request.Intent.Kind == srd521.SpellCast && (request.Intent.EntryID == "cure-wounds" || request.Intent.EntryID == "shield-of-faith"))) && prepared.HealingTarget(request.TargetID) != nil) {
			return nil, ErrActionRefresh
		}
	} else if request.Intent.Kind != srd521.MartialAttack && request.Intent.Kind != combat.EncounterStart && request.Intent.Kind != combat.EncounterReact && request.TargetID != uuid.Nil {
		return nil, ErrActionRefresh
	}
	var inCombat bool
	if err := tx.QueryRow(ctx, `SELECT COALESCE((SELECT COALESCE(call->'Result'->'combat_state'->>'status','active')='active' FROM session_logs l CROSS JOIN LATERAL jsonb_array_elements(COALESCE(l.tool_calls,'[]'::jsonb)) WITH ORDINALITY AS t(call,position) WHERE l.campaign_id=$1 AND call->'Result' ? 'combat_state' ORDER BY l.turn_number DESC,position DESC LIMIT 1),false)`, db.ToPgUUID(campaignID)).Scan(&inCombat); err != nil {
		return nil, err
	}
	if inCombat {
		return nil, ErrActionRefresh
	}
	sheet, err := srd521.Project(pc.Stats, pc.Level, pc.HP, pc.MaxHP, pc.Experience)
	if err != nil {
		return nil, ErrActionRefresh
	}
	if err := game.ApplyFaithAC(ctx, tx, campaignID, pc.ID, prepared.Clock, sheet); err != nil {
		return nil, err
	}
	dice := e.actionDice
	if dice == nil {
		dice = actionDice{}
	}
	label := preview.Intent.Ability
	if preview.Intent.Skill != "" {
		label = preview.Intent.Skill
	}
	var fixtureResolution *srd521.CheckResolution
	var receipt any
	tool := "srd_check"
	narrative := ""
	encounter := prepared.Encounter
	if encounter != nil && encounter.Status == "active" && request.Intent.Kind != combat.EncounterReact && request.Intent.Kind != combat.EncounterDecline && request.Intent.Kind != srd521.EndConcentration {
		var spendErr error
		if request.Intent.Kind == srd521.SpellCast {
			spendErr = encounter.SpendSpell(pc.ID, request.Intent.EntryID == "shield-of-faith")
		} else {
			spendErr = encounter.Spend(pc.ID, request.Intent.Kind)
		}
		if spendErr != nil {
			return nil, ErrActionRefresh
		}

	}
	initialRound := 0
	if encounter != nil {
		initialRound = encounter.Round
	}
	if request.Intent.Kind == combat.EncounterStart {
		var target *game.MartialTarget
		for i := range prepared.Targets {
			if prepared.Targets[i].ID == request.TargetID {
				target = &prepared.Targets[i]
				break
			}
		}
		if target == nil {
			return nil, ErrActionRefresh
		}
		foe := combat.EncounterActor{ID: target.ID, Name: target.Name, HP: target.HP, MaxHP: target.HP}
		if target.Bandit != nil {
			m, _, err := target.Bandit.Record()
			if err != nil {
				return nil, ErrActionRefresh
			}
			foe.Initiative = m.Initiative.Modifier
			foe.Speed = m.Speeds["walk"]
		}
		encounter, err = combat.NewEncounter(combat.EncounterActor{ID: pc.ID, Name: pc.Name, HP: pc.HP, MaxHP: pc.MaxHP}, foe, sheet.Defenses["initiative"].Value, dice)
		if err != nil {
			return nil, err
		}
		encounter.Bandit = target.Bandit
		encounter.Space = &target.Space
		encounter.NPCTactic = target.Tactic
		label = "Initiative"
		tool = "srd_encounter"
		narrative = "Initiative rolled for the two-actor encounter."
	} else if request.Intent.Kind == combat.EncounterEndTurn || request.Intent.Kind == combat.EncounterMove || request.Intent.Kind == combat.EncounterDash || request.Intent.Kind == combat.EncounterApproach || request.Intent.Kind == combat.EncounterRetreat {
		if encounter == nil || encounter.Status != "active" {
			return nil, ErrActionRefresh
		}
		tool = "srd_encounter"
		label = request.Intent.Kind
		narrative = "Spent 5 feet of movement without changing distance."
		if request.Intent.Kind == combat.EncounterApproach || request.Intent.Kind == combat.EncounterRetreat {
			narrative = fmt.Sprintf("Moved 5 feet; target distance is now %d feet.", *encounter.Space.DistanceFeet)
		}
		if request.Intent.Kind == combat.EncounterDash {
			narrative = "Dash: 30 feet added to this turn's movement."
		}
		if request.Intent.Kind == combat.EncounterEndTurn {
			encounter.Advance()
			narrative = "Player turn ended."
		}
	} else if request.Intent.Kind == srd521.ActionSurge {
		var envelope struct {
			Build srd521.Build `json:"srd521"`
		}
		if err := json.Unmarshal(pc.Stats, &envelope); err != nil {
			return nil, err
		}
		surge, err := srd521.UseActionSurge(&envelope.Build)
		if err != nil {
			return nil, ErrActionRefresh
		}
		if _, err := tx.Exec(ctx, `UPDATE player_characters SET stats=jsonb_set(stats,'{srd521,action_surge_spent}','true'::jsonb),updated_at=clock_timestamp() WHERE id=$1 AND campaign_id=$2`, pc.ID, campaignID); err != nil {
			return nil, err
		}
		receipt, tool, label = surge, "srd_action_surge", "Action Surge"
		narrative = "Action Surge saved: one additional Action this turn, except Magic. Bonus Action, Reaction and movement unchanged."
	} else if request.Intent.Kind == srd521.MartialAttack || request.Intent.Kind == combat.EncounterReact {
		var target *game.MartialTarget
		for i := range prepared.Targets {
			if prepared.Targets[i].ID == request.TargetID {
				target = &prepared.Targets[i]
				break
			}
		}
		if target == nil {
			return nil, ErrActionRefresh
		}
		if request.Intent.Kind == combat.EncounterReact {
			if encounter.Reaction.Pending.TargetID != request.TargetID {
				return nil, ErrActionRefresh
			}
			if err := encounter.SpendReaction(); err != nil {
				return nil, ErrActionRefresh
			}
		}
		ac := 10
		if target.Bandit != nil {
			m, _, err := target.Bandit.Record()
			if err != nil {
				return nil, ErrActionRefresh
			}
			ac = m.AC
		}
		faith, err := game.ReadFaithEffect(ctx, tx, campaignID, target.ID, prepared.Clock)
		if err != nil {
			return nil, err
		}
		if faith != nil {
			ac += 2
		}
		attack, err := srd521.ResolveLongsword(sheet, ac+target.Space.CoverBonus(), dice)
		if err != nil {
			return nil, err
		}
		attack.TargetID = target.ID.String()
		attack.Target = target.Name
		if attack.Damage > 0 {
			if _, err := tx.Exec(ctx, `UPDATE npcs SET hp=GREATEST(0,hp-$1),alive=(hp-$1>0),updated_at=clock_timestamp() WHERE id=$2 AND campaign_id=$3`, attack.Damage, db.ToPgUUID(target.ID), db.ToPgUUID(campaignID)); err != nil {
				return nil, err
			}
		}
		if encounter != nil && encounter.Status == "active" {
			encounter.SetHP(target.ID, max(0, target.HP-attack.Damage))
		}
		receipt = attack
		tool = "srd_attack"
		label = "Longsword"
		narrative = fmt.Sprintf("Longsword attack against %s: %s (total %d), %d slashing damage. One fixture Attack action recorded; mastery is not implemented.", target.Name, attack.Outcome, attack.Total, attack.Damage)
		if request.Intent.Kind == combat.EncounterReact {
			narrative = fmt.Sprintf("Opportunity attack against %s: %s (total %d), %d slashing damage. One reaction spent; it recovers at the start of your next turn.", target.Name, attack.Outcome, attack.Total, attack.Damage)
			finishWithdrawal(encounter)
		}
	} else if request.Intent.Kind == combat.EncounterDecline {
		tool = "srd_encounter"
		label = "Decline reaction"
		narrative = "Opportunity attack declined. No reaction spent; the NPC withdrawal resumes."
		finishWithdrawal(encounter)
	} else if srd521.IsEquipment(request.Intent.Kind) {
		if prepared.Equipment == nil || prepared.Build == nil {
			return nil, ErrActionRefresh
		}
		equipment := prepared.Equipment
		tag, err := tx.Exec(ctx, `UPDATE items SET equipped=$1,updated_at=clock_timestamp() WHERE id=$2 AND campaign_id=$3 AND player_character_id=$4 AND equipped=$5 AND quantity=1`, equipment.EquippedAfter, equipment.ItemID, campaignID, pc.ID, equipment.EquippedBefore)
		if err != nil {
			return nil, err
		}
		if tag.RowsAffected() != 1 {
			return nil, ErrActionRefresh
		}
		// Update only the authoritative equipment fields, preserving unknown metadata.
		if _, err := tx.Exec(ctx, `UPDATE player_characters SET stats=jsonb_set(jsonb_set(stats,'{srd521,armor}',to_jsonb($1::text)),'{srd521,shield}',to_jsonb($2::boolean)),updated_at=clock_timestamp() WHERE id=$3 AND campaign_id=$4`, prepared.Build.Armor, prepared.Build.Shield, pc.ID, campaignID); err != nil {
			return nil, err
		}
		if equipment.Expired != nil {
			if err := saveCaster(ctx, tx, campaignID, pc.ID, prepared.Build, pc.HP); err != nil {
				return nil, err
			}
		}
		if err := saveSpellClock(ctx, tx, campaignID, equipment.TimeAfter); err != nil {
			return nil, err
		}
		receipt, tool, label = equipment, "srd_equipment", equipment.Equipment
		narrative = fmt.Sprintf("%s %s saved; AC %d to %d; campaign time advanced %d seconds.", request.Intent.Kind, equipment.Equipment, equipment.ACBefore, equipment.ACAfter, equipment.TimeAfter-equipment.TimeBefore)
		if equipment.Expired != nil {
			narrative += " Shield of Faith expired."
		}
	} else if request.Intent.Kind == srd521.TacticalMind {
		w := prepared.Tactical
		if w == nil {
			return nil, ErrActionRefresh
		}
		r, err := srd521.ResolveTacticalMind(w.Base, w.DC, prepared.Build.SecondWind, request.Intent.EntryID, dice)
		if err != nil {
			return nil, ErrActionRefresh
		}
		r.CheckID = w.CheckID.String()
		if _, err = tx.Exec(ctx, `UPDATE tactical_mind_windows SET resolved=true WHERE actor_id=$1`, pc.ID); err != nil {
			return nil, err
		}
		if _, err = tx.Exec(ctx, `UPDATE player_characters SET stats=jsonb_set(stats,'{srd521,second_wind}',to_jsonb($1::int)),updated_at=clock_timestamp() WHERE id=$2`, r.SecondWindAfter, pc.ID); err != nil {
			return nil, err
		}
		receipt, tool, label = r, "srd_tactical_mind", "Tactical Mind"
		narrative = fmt.Sprintf("Tactical Mind: %s; saved check total %d, d10 %d, total %d. Second Wind %d to %d. Original roll and narration retained.", r.Outcome, r.Base.Total, r.Die, r.Total, r.SecondWindBefore, r.SecondWindAfter)
	} else if request.Intent.Kind == srd521.AdvanceLevel {
		var envelope struct {
			Build srd521.Build `json:"srd521"`
		}
		if err := json.Unmarshal(pc.Stats, &envelope); err != nil {
			return nil, err
		}
		advancement, err := srd521.AdvanceFighter(&envelope.Build, sheet, request.Intent.EntryID, dice)
		if err != nil {
			return nil, ErrActionRefresh
		}
		build, err := json.Marshal(envelope.Build)
		if err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `UPDATE player_characters SET level=2,max_hp=$1,stats=jsonb_set(stats,'{srd521}',$2::jsonb),updated_at=clock_timestamp() WHERE id=$3 AND campaign_id=$4`, advancement.MaxHPAfter, build, pc.ID, campaignID); err != nil {
			return nil, err
		}
		receipt, tool, label = advancement, "srd_advancement", "Fighter level 2"
		narrative = fmt.Sprintf("Fighter level 2 saved: %s HP choice, d10 value %d, maximum HP %d to %d. Current HP unchanged. Defense and weapon mastery choices retained. Action Surge available on your encounter turn; Tactical Mind supports a saved failed check from the selected Edda Easy Strength (Athletics) DC 10 fixture.", advancement.Choice, advancement.Die, advancement.MaxHPBefore, advancement.MaxHPAfter)
	} else if srd521.IsLifecycle(request.Intent.Kind) || (request.Intent.Kind == srd521.SpellCast && request.Intent.EntryID == "shield-of-faith") {
		var envelope struct {
			Build srd521.Build `json:"srd521"`
		}
		if err := json.Unmarshal(pc.Stats, &envelope); err != nil {
			return nil, err
		}
		effect, err := srd521.Lifecycle(&envelope.Build, sheet, request.Intent.Kind, request.Intent.EntryID, request.ActionID.String(), prepared.Clock, dice)
		if err != nil {
			return nil, ErrActionRefresh
		}
		if request.Intent.Kind == srd521.SpellCast && request.TargetID != pc.ID {
			envelope.Build.Casting.Concentration.RecipientID = request.TargetID.String()
			effect.Reason = "Shield of Faith on selected known creature; +2 AC once; one level 1 slot and Bonus Action."
		}
		if err := saveCaster(ctx, tx, campaignID, pc.ID, &envelope.Build, effect.HPAfter); err != nil {
			return nil, err
		}
		if effect.TimeAfter != effect.TimeBefore {
			if err := saveSpellClock(ctx, tx, campaignID, effect.TimeAfter); err != nil {
				return nil, err
			}
		}
		receipt = effect
		tool = "srd_effect"
		label = request.Intent.Kind
		narrative = effect.Reason
		narrative += fmt.Sprintf(" Time advanced %d seconds; HP %d to %d; slots %d to %d; Hit Dice %d to %d.", effect.TimeAfter-effect.TimeBefore, effect.HPBefore, effect.HPAfter, effect.SlotsBefore, effect.SlotsAfter, effect.HitDiceBefore, effect.HitDiceAfter)
		// A same-turn NPC attack must use the newly granted AC, not the preview AC.
		var stats []byte
		if err := tx.QueryRow(ctx, `SELECT stats FROM player_characters WHERE id=$1`, pc.ID).Scan(&stats); err != nil {
			return nil, err
		}
		sheet, err = srd521.Project(stats, pc.Level, effect.HPAfter, pc.MaxHP, pc.Experience)
		if err != nil {
			return nil, err
		}
		if err := game.ApplyFaithAC(ctx, tx, campaignID, pc.ID, effect.TimeAfter, sheet); err != nil {
			return nil, err
		}
	} else if request.Intent.Kind == srd521.SpellCast {
		target := prepared.HealingTarget(request.TargetID)
		hp, maxHP, targetName := pc.HP, pc.MaxHP, pc.Name
		if target != nil {
			hp, maxHP, targetName = target.HP, target.MaxHP, target.Name
		}
		cast, err := srd521.ResolveCureWoundsTarget(sheet, hp, maxHP, dice)
		if err != nil {
			return nil, err
		}
		casterHP := pc.HP
		if target == nil {
			casterHP = cast.HPAfter
		} else {
			if _, err := tx.Exec(ctx, `UPDATE npcs SET hp=$1,updated_at=clock_timestamp() WHERE id=$2 AND campaign_id=$3`, cast.HPAfter, target.ID, campaignID); err != nil {
				return nil, err
			}
		}
		if _, err := tx.Exec(ctx, `UPDATE player_characters SET hp=$1,stats=jsonb_set(stats,'{srd521,casting,slots}',to_jsonb($2::int)),updated_at=clock_timestamp() WHERE id=$3 AND campaign_id=$4`, casterHP, cast.SlotsAfter, db.ToPgUUID(pc.ID), db.ToPgUUID(campaignID)); err != nil {
			return nil, err
		}
		cast.TargetID, cast.Target = request.TargetID.String(), targetName
		if encounter != nil && encounter.Status == "active" {
			encounter.SetHP(pc.ID, casterHP)
		}
		receipt = cast
		tool = "srd_spell"
		label = "Cure Wounds"
		narrativeTarget := "self"
		if target != nil {
			narrativeTarget = targetName
		}
		narrative = fmt.Sprintf("%s casts Cure Wounds on %s using a level 1 slot: regains %d HP (%d to %d). Magic action recorded; %d level 1 slots remain.", pc.Name, narrativeTarget, cast.Healing, cast.HPBefore, cast.HPAfter, cast.SlotsAfter)
	} else if request.Intent.Kind == srd521.ConsumableUse {
		target := prepared.HealingTarget(request.TargetID)
		hp, maxHP, targetName := pc.HP, pc.MaxHP, pc.Name
		if target != nil {
			hp, maxHP, targetName = target.HP, target.MaxHP, target.Name
		}
		healing, err := srd521.ResolveHealingPotion(hp, maxHP, dice)
		if err != nil {
			return nil, err
		}
		itemID, err := uuid.Parse(request.Intent.EntryID)
		if err != nil {
			return nil, ErrActionRefresh
		}
		tag, err := tx.Exec(ctx, `UPDATE items SET quantity=quantity-1,updated_at=clock_timestamp() WHERE id=$1 AND campaign_id=$2 AND player_character_id=$3 AND quantity=$4 AND quantity>0`, db.ToPgUUID(itemID), db.ToPgUUID(campaignID), db.ToPgUUID(pc.ID), prepared.Quantity)
		if err != nil {
			return nil, err
		}
		if tag.RowsAffected() != 1 {
			return nil, ErrActionRefresh
		}
		if target != nil {
			if _, err := tx.Exec(ctx, `UPDATE npcs SET hp=$1,updated_at=clock_timestamp() WHERE id=$2 AND campaign_id=$3`, healing.HPAfter, target.ID, campaignID); err != nil {
				return nil, err
			}
		} else {
			if _, err := tx.Exec(ctx, `UPDATE player_characters SET hp=$1,updated_at=clock_timestamp() WHERE id=$2 AND campaign_id=$3`, healing.HPAfter, db.ToPgUUID(pc.ID), db.ToPgUUID(campaignID)); err != nil {
				return nil, err
			}
		}
		healing.ItemID = itemID.String()
		healing.TargetID = request.TargetID.String()
		healing.Target = targetName
		healing.QuantityBefore = prepared.Quantity
		healing.QuantityAfter = prepared.Quantity - 1
		if target == nil && encounter != nil && encounter.Status == "active" {
			encounter.SetHP(pc.ID, healing.HPAfter)
		}
		receipt = healing
		tool = "srd_consumable"
		label = "Potion of Healing"
		narrative = fmt.Sprintf("%s drinks a Potion of Healing: regains %d HP (%d to %d). One potion consumed; %d remain. Bonus action recorded; a supported encounter tracks its turn budget.", pc.Name, healing.Healing, healing.HPBefore, healing.HPAfter, healing.QuantityAfter)
		if target != nil {
			narrative = fmt.Sprintf("%s administers a Potion of Healing to %s: regains %d HP (%d to %d). One potion consumed; %d remain. Bonus Action recorded.", pc.Name, targetName, healing.Healing, healing.HPBefore, healing.HPAfter, healing.QuantityAfter)
		}
	} else if request.Intent.Kind == srd521.AbilityCheck && request.Intent.EntryID == game.EasyAthleticsEntry {
		fixtureResolution, err = srd521.ResolveCheck(sheet, srd521.Check{Kind: srd521.AbilityCheck, Ability: "strength", Skill: "athletics"}, srd521.Difficulty{Value: game.EasyAthleticsDC, PlayerKnown: true, OutcomeKnown: true}, dice)
		if err != nil {
			return nil, err
		}
		check := fixtureResolution.PlayerReceipt()
		receipt = check
		label = "Easy Strength (Athletics) fixture"
		narrative = fmt.Sprintf("Edda bounded adjudication fixture: Easy Strength (Athletics), DC 10; total %d: %s. No scene effect or time cost applied.", check.Total, check.Outcome)
		if !fixtureResolution.Success() {
			narrative += " Inspect Tactical Mind to explicitly use or decline; no d10 rolled or Second Wind spent."
		}
	} else {
		resolution, err := srd521.ResolveCheck(sheet, srd521.Check{Kind: preview.Intent.Kind, Ability: preview.Intent.Ability, Skill: preview.Intent.Skill}, srd521.Difficulty{}, dice)
		if err != nil {
			return nil, err
		}
		check := resolution.PlayerReceipt()
		receipt = check
		narrative = fmt.Sprintf("%s %s recorded: total %d. The outcome requires adjudication; no scene effect or time cost has been applied.", label, preview.Intent.Kind, check.Total)
	}
	payload := map[string]any{"receipt": receipt}
	if request.Intent.Kind == combat.EncounterReact {
		payload["reaction_cost"] = 1
	}
	if request.Intent.Kind == combat.EncounterDecline {
		payload["reaction_cost"] = 0
	}
	if encounter != nil {
		npcReceipts, concentrationReceipts, err := resolveEncounterNPC(ctx, tx, campaignID, encounter, sheet, dice)
		if err != nil {
			return nil, err
		}
		payload["encounter"] = encounter
		payload["npc_receipts"] = npcReceipts
		payload["concentration_receipts"] = concentrationReceipts
		for _, r := range concentrationReceipts {
			narrative += " " + r.Reason
		}
		if initialRound > 0 && encounter.Round > initialRound {
			// One complete ordinary combat round advances six seconds.
			now := prepared.Clock + 6*(encounter.Round-initialRound)
			if err := saveSpellClock(ctx, tx, campaignID, now); err != nil {
				return nil, err
			}
			var stats []byte
			var hp int
			if err := tx.QueryRow(ctx, `SELECT stats,hp FROM player_characters WHERE id=$1`, pc.ID).Scan(&stats, &hp); err != nil {
				return nil, err
			}
			var envelope struct {
				Build srd521.Build `json:"srd521"`
			}
			if err := json.Unmarshal(stats, &envelope); err != nil {
				return nil, err
			}
			if c := envelope.Build.Casting; c != nil && c.Concentration != nil && c.Concentration.ExpiresAt <= now {
				expiry := &srd521.EffectReceipt{RulesVersion: "srd-5.2.1", Kind: "concentration-expired", Reason: "Shield of Faith expired.", Before: c.Concentration, TimeBefore: prepared.Clock, TimeAfter: now, HPBefore: hp, HPAfter: hp, SlotsBefore: c.Slots, SlotsAfter: c.Slots, HitDiceBefore: envelope.Build.HitDice, HitDiceAfter: envelope.Build.HitDice, SourceIDs: []string{"srd-5.2.1:content/spells/shield-of-faith"}}
				c.Concentration = nil
				if err := saveCaster(ctx, tx, campaignID, pc.ID, &envelope.Build, hp); err != nil {
					return nil, err
				}
				concentrationReceipts = append(concentrationReceipts, expiry)
				payload["concentration_receipts"] = concentrationReceipts
				narrative += " " + expiry.Reason
			}
		}
		for _, attack := range npcReceipts {
			narrative += fmt.Sprintf(" %s: %s (d20 %d + %d = %d), %d %s damage.", attack.Weapon, attack.Outcome, attack.Die, attack.Modifier.Value, attack.Total, attack.Damage, attack.DamageType)
		}
		if encounter.Reaction.Pending != nil {
			narrative += " A visible foe is about to leave your reach. Choose an opportunity attack or decline; no reaction has been spent."
		}
		narrative += fmt.Sprintf(" Round %d; %s. Action: %t; bonus action: %t; movement: %d feet; encounter %s.", encounter.Round, encounter.Active().Name, encounter.Budget.Action, encounter.Budget.BonusAction, encounter.Budget.Movement, encounter.Status)
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	var turn int
	if err := tx.QueryRow(ctx, `SELECT COALESCE(MAX(turn_number),0)+1 FROM session_logs WHERE campaign_id=$1`, db.ToPgUUID(campaignID)).Scan(&turn); err != nil {
		return nil, err
	}
	result := &TurnResult{ActionID: request.ActionID, TurnNumber: turn, Narrative: narrative, AppliedToolCalls: []AppliedToolCall{{Tool: tool, Result: encoded}}}
	resultData, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	calls, err := marshalAppliedToolCalls(result.AppliedToolCalls)
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO turn_attempts(id,campaign_id,turn_number,player_input,status,calls,final_narrative,final_result) VALUES($1,$2,$3,$4,'completed',$5,$6,$7)`, db.ToPgUUID(request.ActionID), db.ToPgUUID(campaignID), turn, input, calls, result.Narrative, resultData); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO session_logs(campaign_id,turn_number,player_input,input_type,llm_response,tool_calls) VALUES($1,$2,$3,'game_action',$4,$5)`, db.ToPgUUID(campaignID), turn, "Confirm "+label+" "+preview.Intent.Kind, result.Narrative, calls); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `UPDATE campaigns SET updated_at=clock_timestamp() WHERE id=$1`, db.ToPgUUID(campaignID)); err != nil {
		return nil, err
	}
	if fixtureResolution != nil && !fixtureResolution.Success() {
		// Snapshot after all final history/campaign writes, under the same transaction
		// and locks: the opening revision must match the next fresh preparation.
		currentCampaign, currentPC, err := game.ReadActionSnapshot(ctx, tx, campaignID, pc.ID, uuid.Nil, true)
		if err != nil {
			return nil, err
		}
		if err := game.OpenTacticalMindWindow(ctx, tx, currentCampaign, currentPC, request.ActionID, fixtureResolution, game.EasyAthleticsDC); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}

// PreviewActionSurge exposes the same read-only preparation to the trusted TUI.
func (e *Engine) PreviewActionSurge(ctx context.Context, campaignID uuid.UUID) (*game.ActionPreview, error) {
	encounter, err := game.ReadEncounter(ctx, e.baseDB, campaignID)
	if err != nil {
		return nil, err
	}
	if encounter == nil || encounter.Status != "active" || !encounter.Active().Player {
		return nil, ErrActionRefresh
	}
	c, pc, err := game.ReadActionSnapshot(ctx, e.baseDB, campaignID, encounter.Active().ID, uuid.Nil, false)
	if err != nil {
		return nil, err
	}
	prepared, err := game.PrepareStoredAction(ctx, e.baseDB, c, pc, game.ActionIntent{ActorID: pc.ID, Kind: srd521.ActionSurge}, false)
	if err != nil {
		return nil, err
	}
	return prepared.Preview, nil
}

// PreviewTacticalMind restores only the saved owned window for the trusted TUI.
func (e *Engine) PreviewTacticalMind(ctx context.Context, campaignID uuid.UUID, choice string) (*game.ActionPreview, error) {
	var actorID uuid.UUID
	if err := e.baseDB.QueryRow(ctx, `SELECT actor_id FROM tactical_mind_windows WHERE campaign_id=$1 AND NOT resolved ORDER BY actor_id LIMIT 1`, campaignID).Scan(&actorID); err != nil {
		return nil, ErrActionRefresh
	}
	c, pc, err := game.ReadActionSnapshot(ctx, e.baseDB, campaignID, actorID, uuid.Nil, false)
	if err != nil {
		return nil, err
	}
	prepared, err := game.PrepareStoredAction(ctx, e.baseDB, c, pc, game.ActionIntent{ActorID: actorID, Kind: srd521.TacticalMind, EntryID: choice}, false)
	if err != nil {
		return nil, err
	}
	return prepared.Preview, nil
}

// PreviewEasyAthletics prepares Edda's fixed DC 10 fixture for the trusted TUI.
func (e *Engine) PreviewEasyAthletics(ctx context.Context, campaignID uuid.UUID) (*game.ActionPreview, error) {
	var actorID uuid.UUID
	if err := e.baseDB.QueryRow(ctx, `SELECT p.id FROM player_characters p JOIN campaigns c ON c.id=p.campaign_id WHERE c.id=$1 AND p.user_id=c.created_by AND p.status='active' ORDER BY p.id LIMIT 1`, campaignID).Scan(&actorID); err != nil {
		return nil, err
	}
	c, pc, err := game.ReadActionSnapshot(ctx, e.baseDB, campaignID, actorID, uuid.Nil, false)
	if err != nil {
		return nil, err
	}
	prepared, err := game.PrepareStoredAction(ctx, e.baseDB, c, pc, game.ActionIntent{ActorID: actorID, Kind: srd521.AbilityCheck, EntryID: game.EasyAthleticsEntry}, false)
	if err != nil {
		return nil, err
	}
	return prepared.Preview, nil
}

// PreviewFaithAction prepares an explicitly selected owned existing caster for
// the TUI. Recipient selection remains a separate explicit confirmation step.
func (e *Engine) PreviewFaithAction(ctx context.Context, campaignID, actorID uuid.UUID, end bool) (*game.ActionPreview, error) {
	if e.baseDB == nil {
		return nil, ErrActionRefresh
	}
	c, p, err := game.ReadActionSnapshot(ctx, e.baseDB, campaignID, actorID, uuid.Nil, false)
	if err != nil {
		return nil, err
	}
	if c.CreatedBy == uuid.Nil || p.UserID != c.CreatedBy {
		return nil, ErrActionRefresh
	}
	intent := game.ActionIntent{ActorID: actorID, Kind: srd521.SpellCast, EntryID: "shield-of-faith", SlotLevel: 1}
	if end {
		intent = game.ActionIntent{ActorID: actorID, Kind: srd521.EndConcentration}
	}
	prepared, err := game.PrepareStoredAction(ctx, e.baseDB, c, p, intent, false)
	if err != nil {
		return nil, err
	}
	return prepared.Preview, nil
}
