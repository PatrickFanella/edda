//go:build integration

package handlers_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"git.subcult.tv/subculture-collective/edda/internal/game"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"github.com/google/uuid"
)

// TestIntegrationDeclaredSliceQualification joins the supported sheet action
// with reloadable projections. Free-text SRD turns are intentionally a failing
// #89 acceptance path: the real engine rejects them before model or dice work.
func TestIntegrationDeclaredSliceQualification(t *testing.T) {
	ctx := context.Background()
	q := statedb.New(testPool)

	create := func(t *testing.T, concept, background string) (uuid.UUID, uuid.UUID, statedb.PlayerCharacter) {
		t.Helper()
		build, err := srd521.SupportedBuildForConcept(concept, background)
		if err != nil {
			t.Fatal(err)
		}
		stats, err := build.Stats()
		if err != nil {
			t.Fatal(err)
		}
		hp := int32(10)
		if concept == "Human Fighter" {
			hp = 12
		}
		c, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Wave 14 " + concept, RulesVersion: string(domain.RulesVersionSRD521), Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _, _ = testPool.Exec(ctx, `DELETE FROM campaigns WHERE id=$1`, c.ID) })
		pc, err := q.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{CampaignID: c.ID, UserID: dbutil.ToPgtype(testUserID), Name: concept, Stats: stats, Hp: hp, MaxHp: hp, Level: 1, Status: "active", Abilities: []byte(`[]`)})
		if err != nil {
			t.Fatal(err)
		}
		return dbutil.FromPgtype(c.ID), dbutil.FromPgtype(pc.ID), pc
	}
	assertProjection := func(t *testing.T, pc statedb.PlayerCharacter, scope string) {
		t.Helper()
		reloaded, err := q.GetPlayerCharacterByID(ctx, pc.ID)
		if err != nil {
			t.Fatal(err)
		}
		sheet, err := srd521.Project(reloaded.Stats, int(reloaded.Level), int(reloaded.Hp), int(reloaded.MaxHp), int(reloaded.Experience))
		if err != nil {
			t.Fatal(err)
		}
		if sheet.Scope != scope || sheet.Coverage.Status != "partial" || sheet.Coverage.FullSupport || sheet.Coverage.PlannedCharacterScopes != 344 || sheet.Coverage.PlannedWorldScopes != 335 || sheet.Coverage.PlannedSpellItemScopes != 753 {
			t.Fatalf("sheet=%+v", sheet)
		}
	}

	t.Run("fighter typed ability check persists and replays", func(t *testing.T) {
		campaign, actor, pc := create(t, "Human Fighter", "Soldier")
		assertProjection(t, pc, "fighter-1-chain-mail-defense-read-only")
		campaignData, character, err := game.ReadActionSnapshot(ctx, testPool, campaign, actor, testUserID, false)
		if err != nil {
			t.Fatal(err)
		}
		preview, err := game.PrepareCharacterAction(ctx, campaignData, character, game.ActionIntent{ActorID: actor, Kind: "ability-check", Skill: "athletics"})
		if err != nil || !preview.Available {
			t.Fatalf("preview=%+v err=%v", preview, err)
		}
		dice := &confirmationDice{}
		newEngine := func() *engine.Engine {
			e, err := engine.New(testPool, nil, config.LLMConfig{}, engine.WithActionDice(dice))
			if err != nil {
				t.Fatal(err)
			}
			return e
		}
		request := engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: preview.ExpectedRevision, Intent: preview.Intent}
		first, err := newEngine().ConfirmAction(ctx, campaign, request)
		if err != nil {
			t.Fatal(err)
		}
		replay, err := newEngine().ConfirmAction(ctx, campaign, request)
		if err != nil || !sameTurnResult(first, replay) || dice.rolls.Load() != 1 {
			t.Fatalf("replay=%+v err=%v rolls=%d", replay, err, dice.rolls.Load())
		}
		receipt, err := newEngine().GetActionReceipt(ctx, campaign, request.ActionID)
		if err != nil || receipt.Status != "completed" || !sameTurnResult(first, receipt.Result) {
			t.Fatalf("receipt=%+v err=%v", receipt, err)
		}
		assertProjection(t, pc, "fighter-1-chain-mail-defense-read-only")
		var before, after string
		if err := testPool.QueryRow(ctx, `SELECT row_to_json(p)::text FROM player_characters p WHERE id=$1`, pc.ID).Scan(&before); err != nil {
			t.Fatal(err)
		}
		var writesBefore, writesAfter int
		if err := testPool.QueryRow(ctx, `SELECT (SELECT count(*) FROM session_logs WHERE campaign_id=$1)+(SELECT count(*) FROM turn_attempts WHERE campaign_id=$1)`, dbutil.ToPgtype(campaign)).Scan(&writesBefore); err != nil {
			t.Fatal(err)
		}
		provider := &qualificationProvider{}
		guardDice := &confirmationDice{}
		e, err := engine.New(testPool, provider, config.LLMConfig{}, engine.WithActionDice(guardDice))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := e.ProcessTurn(ctx, campaign, "Inspect the gate"); !errors.Is(err, domain.ErrLegacyRulesMutation) {
			t.Fatalf("turn error=%v", err)
		}
		stream, err := e.ProcessTurnStream(ctx, campaign, "Inspect the gate")
		if err != nil {
			t.Fatal(err)
		}
		var streamErr error
		for event := range stream {
			if event.Type == "error" {
				streamErr = event.Err
			}
		}
		if !errors.Is(streamErr, domain.ErrLegacyRulesMutation) || provider.calls.Load() != 0 || guardDice.rolls.Load() != 0 {
			t.Fatalf("stream=%v model=%d rolls=%d", streamErr, provider.calls.Load(), guardDice.rolls.Load())
		}
		if err := testPool.QueryRow(ctx, `SELECT (SELECT count(*) FROM session_logs WHERE campaign_id=$1)+(SELECT count(*) FROM turn_attempts WHERE campaign_id=$1)`, dbutil.ToPgtype(campaign)).Scan(&writesAfter); err != nil || writesAfter != writesBefore {
			t.Fatalf("writes before=%d after=%d err=%v", writesBefore, writesAfter, err)
		}
		if err := testPool.QueryRow(ctx, `SELECT row_to_json(p)::text FROM player_characters p WHERE id=$1`, pc.ID).Scan(&after); err != nil || after != before {
			t.Fatalf("state changed before=%q after=%q err=%v", before, after, err)
		}
	})

	t.Run("cleric projection reloads in its own campaign", func(t *testing.T) {
		_, _, pc := create(t, "Human Cleric", "Acolyte")
		assertProjection(t, pc, "cleric-1-cure-wounds-self-fixture")
	})
}

type qualificationProvider struct{ calls atomic.Int32 }

func (p *qualificationProvider) Complete(context.Context, []llm.Message, []llm.Tool) (*llm.Response, error) {
	p.calls.Add(1)
	return nil, errors.New("unexpected model call")
}

func (p *qualificationProvider) Stream(context.Context, []llm.Message, []llm.Tool) (<-chan llm.StreamChunk, error) {
	p.calls.Add(1)
	return nil, errors.New("unexpected model call")
}
