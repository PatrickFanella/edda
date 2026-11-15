package engine

import (
	"context"
	"errors"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	"git.subcult.tv/subculture-collective/edda/internal/tools"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"testing"
)

type uncertainReceiptTx struct {
	pgx.Tx
	committed bool
}

func (tx *uncertainReceiptTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag("UPDATE 1"), nil
}
func (tx *uncertainReceiptTx) Commit(context.Context) error {
	tx.committed = true
	return errors.New("commit reply lost")
}
func (tx *uncertainReceiptTx) Rollback(context.Context) error { return nil }

type receiptBeginner struct{ tx pgx.Tx }

func (b receiptBeginner) Begin(context.Context) (pgx.Tx, error) { return b.tx, nil }
func TestUncertainToolCommitStopsProviderRecoveryAndRequiresReceiptRead(t *testing.T) {
	tx := &uncertainReceiptTx{}
	e := &Engine{turnDB: receiptBeginner{tx}}
	j := &turnJournal{id: uuid.New(), tc: &TurnContext{}}
	ctx := context.WithValue(context.Background(), journalKey{}, j)
	reg := tools.NewRegistry()
	reg.AroundExecute(e.journalTool)
	executions := 0
	if err := reg.Register(llm.Tool{Name: "probe", Parameters: map[string]any{"type": "object"}}, func(context.Context, map[string]any) (*tools.ToolResult, error) {
		executions++
		return &tools.ToolResult{Success: true}, nil
	}); err != nil {
		t.Fatal(err)
	}
	p := newScriptedProvider(t, scriptedResponse{resp: &llm.Response{Content: "A pause.", ToolCalls: []llm.ToolCall{{ID: "probe", Name: "probe", Arguments: map[string]any{}}}}})
	tp := NewTurnProcessor(p, reg, tools.NewValidator(reg), testLogger())
	_, _, err := tp.ProcessWithRecovery(ctx, nil, reg.List())
	if !errors.Is(err, ErrTurnPersistence) || executions != 1 || p.callCount != 1 || !j.uncertain || !tx.committed {
		t.Fatalf("uncertain commit retried: err=%v executions=%d provider=%d journal=%+v", err, executions, p.callCount, j)
	}
	if !errors.Is(e.turnFailure(ctx, err), ErrTurnRecoveryRequired) {
		t.Fatal("uncertain commit was reported as a safe new action")
	}
}
