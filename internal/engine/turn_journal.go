package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"git.subcult.tv/subculture-collective/edda/internal/tools"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrTurnRecoveryRequired = errors.New("recorded turn needs recovery")
var ErrActionIDConflict = errors.New("action_id belongs to a different campaign or input")

var ErrTurnPersistence = errors.New("turn receipt persistence failed")
var errRecoveredTurn = errors.New("recorded turn recovered")

const mechanicalNarrative = "Mechanical outcomes were recorded. Narration could not be completed; no effects were repeated during recovery."

type actionIDKey struct{}
type journalKey struct{}
type toolTxKey struct{}

// WithActionID identifies retries without confusing repeated input with a new action.
func WithActionID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, actionIDKey{}, id)
}

type recoveryError struct {
	id    uuid.UUID
	cause error
}

func (e *recoveryError) Error() string {
	return fmt.Sprintf("%s: retry the original action with action_id %s; effects will not repeat", ErrTurnRecoveryRequired, e.id)
}
func (e *recoveryError) Unwrap() error { return ErrTurnRecoveryRequired }
func RecoveryActionID(err error) string {
	var r *recoveryError
	if errors.As(err, &r) {
		return r.id.String()
	}
	return ""
}

type turnJournal struct {
	id        uuid.UUID
	tc        *TurnContext
	calls     []AppliedToolCall
	uncertain bool
	candidate bool
	failure   error
}

// toolDB routes a tool invocation's queries into its short transaction.
// Provider calls and post-commit memory callbacks use the base connection.
type toolDB struct{ base statedb.DBTX }

func (d toolDB) target(ctx context.Context) statedb.DBTX {
	if tx, ok := ctx.Value(toolTxKey{}).(pgx.Tx); ok {
		return tx
	}
	return d.base
}
func (d toolDB) Exec(ctx context.Context, s string, args ...any) (pgconn.CommandTag, error) {
	return d.target(ctx).Exec(ctx, s, args...)
}
func (d toolDB) Query(ctx context.Context, s string, args ...any) (pgx.Rows, error) {
	return d.target(ctx).Query(ctx, s, args...)
}
func (d toolDB) QueryRow(ctx context.Context, s string, args ...any) pgx.Row {
	return d.target(ctx).QueryRow(ctx, s, args...)
}
func rollbackTurnTx(tx pgx.Tx) {
	cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = tx.Rollback(cleanup)
}

// Mechanical effects and their exact tool receipt commit together. No database
// transaction spans inference. Partial turns remain recoverable in turn_attempts,
// outside completed history, replay, summaries and save points.
func (e *Engine) journalTool(ctx context.Context, name string, args map[string]any, handler tools.Handler) (returnedResult *tools.ToolResult, returnedError error) {
	j, _ := ctx.Value(journalKey{}).(*turnJournal)
	if j != nil && j.failure != nil {
		return nil, j.failure
	}
	defer func() {
		if j != nil && errors.Is(returnedError, ErrTurnPersistence) {
			j.failure = returnedError
		}
	}()
	if j == nil || name == "search_memory" {
		return handler(ctx, args)
	}
	tx, err := e.turnDB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: begin: %v", ErrTurnPersistence, err)
	}
	defer rollbackTurnTx(tx)
	workCtx, flushMemory := tools.WithDeferredMemory(context.WithValue(ctx, toolTxKey{}, tx))
	result, err := handler(workCtx, args)
	if err != nil {
		return nil, err
	}
	call, err := buildAppliedToolCall(llm.ToolCall{Name: name, Arguments: args}, result)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTurnPersistence, err)
	}
	calls := append(append([]AppliedToolCall(nil), j.calls...), call)
	data, err := marshalAppliedToolCalls(calls)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTurnPersistence, err)
	}
	tag, err := tx.Exec(ctx, `UPDATE turn_attempts SET calls=$2,updated_at=now() WHERE id=$1 AND status='pending_mechanics'`, dbutil.ToPgtype(j.id), data)
	if err != nil || tag.RowsAffected() != 1 {
		return nil, fmt.Errorf("%w: record tool receipt: %v", ErrTurnPersistence, err)
	}
	j.uncertain = true
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%w: commit: %v", ErrTurnPersistence, err)
	}
	j.uncertain = false
	j.calls = calls
	flushMemory(ctx)
	return result, nil
}

type storedAttempt struct {
	result                          []byte
	id                              uuid.UUID
	turn                            int
	input, status, candidate, final string
	calls                           []byte
	recovered                       bool
}

func scanAttempt(row pgx.Row) (storedAttempt, error) {
	var a storedAttempt
	err := row.Scan(&a.id, &a.turn, &a.input, &a.status, &a.candidate, &a.final, &a.calls, &a.recovered, &a.result)
	return a, err
}

const attemptColumns = `id,turn_number,player_input,status,candidate_narrative,final_narrative,calls,recovered,final_result`

func (e *Engine) startJournal(ctx context.Context, tc *TurnContext) error {
	if e.turnDB == nil {
		return nil
	}
	id, _ := ctx.Value(actionIDKey{}).(uuid.UUID)
	var a storedAttempt
	var err error
	if id != uuid.Nil {
		a, err = scanAttempt(e.baseDB.QueryRow(ctx, `SELECT `+attemptColumns+` FROM turn_attempts WHERE id=$1 AND campaign_id=$2`, dbutil.ToPgtype(id), dbutil.ToPgtype(tc.CampaignID)))
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if err == nil && a.input != tc.PlayerInput {
			return ErrActionIDConflict
		}
	}
	if a.id == uuid.Nil || a.status == "failed" {
		pending, pendingErr := scanAttempt(e.baseDB.QueryRow(ctx, `SELECT `+attemptColumns+` FROM turn_attempts WHERE campaign_id=$1 AND status IN ('pending_mechanics','pending_narration')`, dbutil.ToPgtype(tc.CampaignID)))
		if pendingErr == nil {
			if id != uuid.Nil && id != pending.id || strings.TrimSpace(pending.input) != strings.TrimSpace(tc.PlayerInput) {
				return &recoveryError{id: pending.id}
			}
			a = pending
		} else if !errors.Is(pendingErr, pgx.ErrNoRows) {
			return pendingErr
		}
	}
	if a.id != uuid.Nil && a.status != "failed" {
		if err := json.Unmarshal(a.calls, &tc.Applied); err != nil {
			return err
		}
		tc.PlayerInput = a.input
		tc.ActionID = a.id
		tc.TurnNumber = a.turn
		tc.Recovered = a.recovered
		tc.StateChanges = StateChangesFromAppliedToolCalls(tc.Applied)
		tc.CombatActive = combatActiveAfterApplied(tc.State.CombatActive, tc.Applied)
		if a.status == "completed" {
			var result TurnResult
			if err := json.Unmarshal(a.result, &result); err != nil {
				return fmt.Errorf("decode recorded turn result: %w", err)
			}
			tc.Narrative = result.Narrative
			tc.Choices = result.Choices
			tc.Applied = result.AppliedToolCalls
			tc.StateChanges = result.StateChanges
			tc.CombatActive = result.CombatActive
			tc.Recovered = result.Recovered
			return errRecoveredTurn
		}
		// With no committed receipt there is no effect to replay; restart safely.
		if len(tc.Applied) == 0 && a.status == "pending_mechanics" {
			id = a.id
		} else {
			tc.Narrative = mechanicalNarrative
			tc.Recovered = true
			if err := e.finalizeAttempt(ctx, tc, a.id, mechanicalNarrative); err != nil {
				return &recoveryError{id: a.id, cause: err}
			}
			return errRecoveredTurn
		}
	}
	if id == uuid.Nil {
		id = uuid.New()
	}
	tc.ActionID = id
	tc.TurnNumber = nextTurnNumber(tc.RecentLogs)
	tag, err := e.baseDB.Exec(ctx, `INSERT INTO turn_attempts(id,campaign_id,turn_number,player_input,status) VALUES($1,$2,$3,$4,'pending_mechanics')
 ON CONFLICT(id) DO UPDATE SET status='pending_mechanics',updated_at=now() WHERE turn_attempts.status IN ('failed','pending_mechanics') AND turn_attempts.campaign_id=EXCLUDED.campaign_id AND turn_attempts.player_input=EXCLUDED.player_input`, dbutil.ToPgtype(id), dbutil.ToPgtype(tc.CampaignID), tc.TurnNumber, tc.PlayerInput)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrActionIDConflict
	}
	tc.Ctx = context.WithValue(tc.Ctx, journalKey{}, &turnJournal{id: id, tc: tc})
	return nil
}
func (e *Engine) recordCandidate(ctx context.Context, tc *TurnContext) error {
	j, _ := ctx.Value(journalKey{}).(*turnJournal)
	if j == nil {
		return nil
	}
	_, err := e.baseDB.Exec(ctx, `UPDATE turn_attempts SET candidate_narrative=$2,status='pending_narration',updated_at=now() WHERE id=$1`, dbutil.ToPgtype(j.id), tc.Narrative)
	if err != nil {
		return fmt.Errorf("%w: candidate: %v", ErrTurnPersistence, err)
	}
	j.candidate = true
	return nil
}
func (e *Engine) finalizeAttempt(ctx context.Context, tc *TurnContext, id uuid.UUID, narrative string) error {
	tx, err := e.turnDB.Begin(ctx)
	if err != nil {
		return err
	}
	defer rollbackTurnTx(tx)
	data, err := marshalAppliedToolCalls(tc.Applied)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO session_logs(campaign_id,turn_number,player_input,input_type,llm_response,tool_calls,location_id)
 VALUES($1,$2,$3,$4,$5,$6,$7)`, dbutil.ToPgtype(tc.CampaignID), tc.TurnNumber, tc.PlayerInput, string(domain.Classify(tc.PlayerInput)), narrative, data, optionalJournalLocation(tc))
	if err != nil {
		return err
	}
	resultJSON, err := json.Marshal(resultFromTurn(tc))
	if err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `UPDATE turn_attempts SET status='completed',final_narrative=$2,recovered=$3,final_result=$4,updated_at=now() WHERE id=$1 AND status IN ('pending_mechanics','pending_narration')`, dbutil.ToPgtype(id), narrative, tc.Recovered, resultJSON)
	if err != nil || tag.RowsAffected() != 1 {
		return fmt.Errorf("finalize attempt: %v", err)
	}
	return tx.Commit(ctx)
}
func optionalJournalLocation(tc *TurnContext) any {
	location := finalLocationIDFromApplied(tc.State.Player.CurrentLocationID, tc.Applied)
	if location == nil {
		return nil
	}
	return dbutil.ToPgtype(*location)
}
func (e *Engine) persistCompletedTurn(ctx context.Context, tc *TurnContext, log domain.SessionLog) error {
	j, _ := ctx.Value(journalKey{}).(*turnJournal)
	if j == nil {
		return e.state.SaveSessionLog(ctx, log)
	}
	tc.Applied = j.calls
	return e.finalizeAttempt(ctx, tc, j.id, log.LLMResponse)
}
func (e *Engine) turnFailure(ctx context.Context, err error) error {
	if ctx == nil {
		return err
	}
	j, _ := ctx.Value(journalKey{}).(*turnJournal)
	if j == nil {
		return err
	}
	if len(j.calls) > 0 || j.candidate || j.uncertain || errors.Is(err, ErrTurnPersistence) {
		return &recoveryError{id: j.id, cause: err}
	}
	cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = e.baseDB.Exec(cleanup, `UPDATE turn_attempts SET status='failed',updated_at=now() WHERE id=$1 AND calls='[]'::jsonb AND status<>'completed'`, dbutil.ToPgtype(j.id))
	return err
}

func (e *Engine) recordRawCandidate(ctx context.Context, text string) error {
	j, _ := ctx.Value(journalKey{}).(*turnJournal)
	if j == nil {
		return nil
	}
	_, err := e.baseDB.Exec(ctx, `UPDATE turn_attempts SET candidate_narrative=$2,updated_at=now() WHERE id=$1`, dbutil.ToPgtype(j.id), text)
	if err != nil {
		return fmt.Errorf("%w: raw narrative: %v", ErrTurnPersistence, err)
	}
	return nil
}

func turnPersistenceFailure(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	j, _ := ctx.Value(journalKey{}).(*turnJournal)
	if j != nil {
		return j.failure
	}
	return nil
}

func resultFromTurn(tc *TurnContext) *TurnResult {
	return &TurnResult{ActionID: tc.ActionID, Recovered: tc.Recovered, TurnNumber: tc.TurnNumber, Narrative: tc.Narrative, AppliedToolCalls: tc.Applied, StateChanges: tc.StateChanges, Choices: tc.Choices, CombatActive: tc.CombatActive}
}
