//go:build integration

package handlers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"git.subcult.tv/subculture-collective/edda/internal/tools"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type turnBarrierProvider struct {
	campaign uuid.UUID
	entered  chan struct{}
	release  chan struct{}
}

func (p *turnBarrierProvider) Complete(ctx context.Context, _ []llm.Message, _ []llm.Tool) (*llm.Response, error) {
	id, _ := tools.CurrentCampaignIDFromContext(ctx)
	if id == p.campaign {
		select {
		case p.entered <- struct{}{}:
		default:
		}
		select {
		case <-p.release:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return &llm.Response{Content: "The wind continues."}, nil
}
func (*turnBarrierProvider) Stream(context.Context, []llm.Message, []llm.Tool) (<-chan llm.StreamChunk, error) {
	return nil, errors.New("unused")
}

func TestIntegrationCampaignTurnOrder(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	q := statedb.New(testPool)
	create := func(name string) uuid.UUID {
		c, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: name, Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := q.DeleteCampaign(context.Background(), c.ID); err != nil {
				t.Error(err)
			}
		})
		return dbutil.FromPgtype(c.ID)
	}
	first, other := create("Overlap"), create("Independent")
	p := &turnBarrierProvider{campaign: first, entered: make(chan struct{}, 1), release: make(chan struct{})}
	poolCfg := testPool.Config()
	poolCfg.MaxConns = 1
	turnPool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer turnPool.Close()
	// Separate engines exercise process-independent PostgreSQL exclusion.
	e1, err := engine.New(turnPool, p, config.LLMConfig{})
	if err != nil {
		t.Fatal(err)
	}
	e2, err := engine.New(turnPool, p, config.LLMConfig{})
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := e1.ProcessTurn(ctx, first, "Wait."); done <- err }()
	defer func() { cancel() }()
	select {
	case <-p.entered:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if _, err := e2.ProcessTurn(ctx, first, "Also wait."); !errors.Is(err, engine.ErrTurnInProgress) {
		t.Fatalf("overlap = %v", err)
	}
	stream, err := e2.ProcessTurnStream(ctx, first, "Stream overlap.")
	if err != nil {
		t.Fatal(err)
	}
	var streamErr error
	for ev := range stream {
		if ev.Type == "error" {
			streamErr = ev.Err
		}
	}
	if !errors.Is(streamErr, engine.ErrTurnInProgress) {
		t.Fatalf("stream overlap = %v", streamErr)
	}
	if _, err := e2.ProcessTurn(ctx, other, "Wait independently."); err != nil {
		t.Fatal(err)
	}
	close(p.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	// The next accepted turn must see the completed first turn.
	if _, err := e2.ProcessTurn(ctx, first, "Wait again."); err != nil {
		t.Fatal(err)
	}
	logs, err := q.ListSessionLogsByCampaign(ctx, dbutil.ToPgtype(first))
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 2 || logs[0].TurnNumber != 1 || logs[1].TurnNumber != 2 {
		t.Fatalf("turns = %+v", logs)
	}
	// Bypassing the engine still cannot duplicate a turn number.
	_, err = q.CreateSessionLog(ctx, statedb.CreateSessionLogParams{CampaignID: dbutil.ToPgtype(first), TurnNumber: 1, PlayerInput: "duplicate", InputType: "game_action", LlmResponse: "duplicate"})
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Fatalf("duplicate insert = %v", err)
	}
}
