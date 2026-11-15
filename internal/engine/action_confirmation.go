package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"

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
	// This slice has no scene-target resolver or combat action economy. A save
	// targets the actor; an ability/skill check has no selected scene target.
	if request.Intent.Kind == srd521.SavingThrow {
		if request.TargetID != pc.ID {
			return nil, ErrActionRefresh
		}
	} else if request.Intent.Kind != srd521.MartialAttack && request.TargetID != uuid.Nil {
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
	dice := e.actionDice
	if dice == nil {
		dice = actionDice{}
	}
	label := preview.Intent.Ability
	if preview.Intent.Skill != "" {
		label = preview.Intent.Skill
	}
	var receipt any
	tool := "srd_check"
	narrative := ""
	if request.Intent.Kind == srd521.MartialAttack {
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
		attack, err := srd521.ResolveLongsword(sheet, 10, dice)
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
		receipt = attack
		tool = "srd_attack"
		label = "Longsword"
		narrative = fmt.Sprintf("Longsword attack against %s: %s (total %d), %d slashing damage. One fixture Attack action recorded; initiative, mastery and reactions are not implemented.", target.Name, attack.Outcome, attack.Total, attack.Damage)
	} else {
		resolution, err := srd521.ResolveCheck(sheet, srd521.Check{Kind: preview.Intent.Kind, Ability: preview.Intent.Ability, Skill: preview.Intent.Skill}, srd521.Difficulty{}, dice)
		if err != nil {
			return nil, err
		}
		check := resolution.PlayerReceipt()
		receipt = check
		narrative = fmt.Sprintf("%s %s recorded: total %d. The outcome requires adjudication; no scene effect or time cost has been applied.", label, preview.Intent.Kind, check.Total)
	}
	encoded, err := json.Marshal(map[string]any{"receipt": receipt})
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
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}
