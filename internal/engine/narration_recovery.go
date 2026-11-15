package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"git.subcult.tv/subculture-collective/edda/internal/db"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// MaxNarrationAttempts bounds recovery inference across retries and restarts.
// Persistence retries of an already validated candidate do not spend attempts.
const MaxNarrationAttempts = 3

// ErrNarrationAttemptsExhausted leaves a receipt pending for operator recovery.
var ErrNarrationAttemptsExhausted = errors.New("narration recovery attempt limit reached")

func (e *Engine) recoverNarration(ctx context.Context, tc *TurnContext, a storedAttempt) error {
	tc.Recovered = true
	j := &turnJournal{id: a.id, tc: tc, calls: tc.Applied, candidate: a.narrationReady}
	tc.Ctx = context.WithValue(ctx, journalKey{}, j)
	ctx = tc.Ctx
	if len(a.candidateResult) > 0 {
		var result TurnResult
		if err := json.Unmarshal(a.candidateResult, &result); err != nil {
			return &recoveryError{id: a.id, cause: err}
		}
		tc.Narrative, tc.Choices = result.Narrative, result.Choices
		tc.StateChanges, tc.CombatActive = result.StateChanges, result.CombatActive
		if err := e.finalizeAttempt(ctx, tc, a.id, tc.Narrative); err != nil {
			return &recoveryError{id: a.id, cause: err}
		}
		e.snapshotQuestsIfNeeded(ctx, tc.CampaignID, tc.Applied)
		e.autoSaveIfNeeded(ctx, tc.CampaignID, tc.TurnNumber)
		e.autoSummarizeIfNeeded(ctx, tc.CampaignID, tc.TurnNumber)
		return errRecoveredTurn
	}
	if !a.narrationReady {
		if err := e.claimNarrationAttempt(ctx, a.id); err != nil {
			return &recoveryError{id: a.id, cause: err}
		}
		messages := e.assembler.AssembleContext(tc.State, tc.RecentLogs, tc.PlayerInput)
		receipts, err := marshalAppliedToolCalls(tc.Applied)
		if err != nil {
			return &recoveryError{id: a.id, cause: err}
		}
		messages = append(messages, llm.Message{Role: llm.RoleUser, Content: fmt.Sprintf("Recover narration for the ORIGINAL action above. Its mechanical processing has stopped. The following receipts are already committed and are the authoritative outcomes: %s\nDescribe only these outcomes and the current player-visible scene. Preserve every recorded die and cost; do not roll, repeat tools, spend resources, or invent additional effects or hidden information. Unfinished effects did not happen. Return narrative with a Choices section; call no tools.", receipts)})
		resp, err := e.provider.Complete(ctx, messages, nil)
		if err != nil {
			return &recoveryError{id: a.id, cause: err}
		}
		if err := validateRecoveryNarration(resp, tc.Applied); err != nil {
			return &recoveryError{id: a.id, cause: err}
		}
		tc.Narrative = resp.Content
		if err := e.recordCandidate(ctx, tc); err != nil {
			return &recoveryError{id: a.id, cause: err}
		}
	} else {
		tc.Narrative = a.candidate
	}
	// An interrupted normal completion may have prose without choices. Generate
	// those through the existing choice provider, under the same durable budget.
	_, choices, err := extractChoicesStrict(tc.Narrative)
	if err != nil {
		return &recoveryError{id: a.id, cause: err}
	}
	if len(choices) == 0 && e.choiceProvider != nil {
		if err := e.claimNarrationAttempt(ctx, a.id); err != nil {
			return &recoveryError{id: a.id, cause: err}
		}
		choices, err = NewTurnCompleter(e, e.choiceProvider).generateChoices(ctx, tc.Narrative)
		if err != nil || len(choices) == 0 {
			return &recoveryError{id: a.id, cause: ErrEmptyTurnResponse}
		}
		for i, choice := range choices {
			tc.Narrative += fmt.Sprintf("\n%d. %s", i+1, choice.Text)
		}
		if err := e.recordCandidate(ctx, tc); err != nil {
			return &recoveryError{id: a.id, cause: err}
		}
	}
	if err := NewTurnCompleter(e, nil).CompleteAndPersist(ctx, tc); err != nil {
		return &recoveryError{id: a.id, cause: err}
	}
	return errRecoveredTurn
}

// claimNarrationAttempt charges before inference; a crash cannot reset its budget.
func (e *Engine) claimNarrationAttempt(ctx context.Context, id uuid.UUID) error {
	tag, err := e.baseDB.Exec(ctx, `UPDATE turn_attempts SET status='pending_narration',narration_attempts=narration_attempts+1,updated_at=now() WHERE id=$1 AND narration_attempts<$2 AND status IN ('pending_mechanics','pending_narration')`, db.ToPgUUID(id), MaxNarrationAttempts)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrNarrationAttemptsExhausted
	}
	return nil
}

// Save the completed projection before final history persistence, retaining
// generated choices and derived state even if delivery or finalization fails.
func (e *Engine) recordCompletionCandidate(ctx context.Context, tc *TurnContext) error {
	j, _ := ctx.Value(journalKey{}).(*turnJournal)
	if j == nil {
		return nil
	}
	data, err := json.Marshal(resultFromTurn(tc))
	if err != nil {
		return err
	}
	tag, err := e.baseDB.Exec(ctx, `UPDATE turn_attempts SET candidate_result=$2,updated_at=now() WHERE id=$1 AND status='pending_narration'`, db.ToPgUUID(j.id), data)
	if err != nil || tag.RowsAffected() != 1 {
		return fmt.Errorf("%w: completion candidate: %v", ErrTurnPersistence, err)
	}
	return nil
}

func validateRecoveryNarration(resp *llm.Response, applied []AppliedToolCall) error {
	if resp == nil || strings.TrimSpace(resp.Content) == "" {
		return ErrEmptyTurnResponse
	}
	// A hallucinated tool call is rejected, never dispatched to the registry.
	if len(resp.ToolCalls) != 0 || len(AuditDurableClaims(resp.Content, applied, nil)) != 0 {
		return ErrUnresolvedDurableClaims
	}
	cleaned, _, err := extractChoicesStrict(resp.Content)
	if err == nil && strings.TrimSpace(stripDanglingChoicesMarker(cleaned)) == "" {
		return ErrEmptyTurnResponse
	}
	return err
}

// ActionReceipt is a durable read of an action's lifecycle and exact effects.
// Pending candidates are withheld until they have passed turn completion.
type ActionReceipt struct {
	ActionID          uuid.UUID
	PlayerInput       string
	Status            string
	TurnNumber        int
	NarrationAttempts int
	AppliedToolCalls  []AppliedToolCall
	Result            *TurnResult
}

// GetActionReceipt returns a campaign-scoped receipt, or nil when absent.
// A nil actionID selects the campaign's pending action for reconnect discovery.
// API adapters must enforce campaign ownership before exposing this read.
func (e *Engine) GetActionReceipt(ctx context.Context, campaignID, actionID uuid.UUID) (*ActionReceipt, error) {
	if e.baseDB == nil {
		return nil, nil
	}
	query := `SELECT ` + attemptColumns + ` FROM turn_attempts WHERE campaign_id=$1 AND id=$2`
	args := []any{db.ToPgUUID(campaignID), db.ToPgUUID(actionID)}
	if actionID == uuid.Nil {
		query = `SELECT ` + attemptColumns + ` FROM turn_attempts WHERE campaign_id=$1 AND status IN ('pending_mechanics','pending_narration')`
		args = args[:1]
	}
	a, err := scanAttempt(e.baseDB.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r := &ActionReceipt{ActionID: a.id, PlayerInput: a.input, Status: a.status, TurnNumber: a.turn, NarrationAttempts: a.narrationAttempts}
	if err := json.Unmarshal(a.calls, &r.AppliedToolCalls); err != nil {
		return nil, err
	}
	// A process may have died between committing a tool and marking narration
	// pending. Receipts, rather than the former connection, determine the state.
	if r.Status == "pending_mechanics" && len(r.AppliedToolCalls) > 0 {
		r.Status = "pending_narration"
	}
	if a.status == "completed" {
		if err := json.Unmarshal(a.result, &r.Result); err != nil {
			return nil, err
		}
	}
	return r, nil
}
