package tui

import (
	"context"
	"testing"

	"git.subcult.tv/subculture-collective/edda/internal/game"
	"git.subcult.tv/subculture-collective/edda/tui/narrative"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type faithMock struct{ tacticalMock }

func (m *faithMock) PreviewFaithAction(_ context.Context, _ uuid.UUID, actor uuid.UUID, end bool) (*game.ActionPreview, error) {
	p := &game.ActionPreview{Available: true, ExpectedRevision: "faith", Intent: game.ActionIntent{ActorID: actor, Kind: "spell-cast", EntryID: "shield-of-faith", SlotLevel: 1}, Targets: []game.ActionTarget{{ID: actor, Label: "self"}}, Options: []string{"one slot; Bonus Action; +2 AC once"}}
	if end {
		p.Intent = game.ActionIntent{ActorID: actor, Kind: "concentration-end"}
		p.Targets = nil
	}
	return p, nil
}
func TestTUIFaithExplicitConfirmationRetry(t *testing.T) {
	m := &faithMock{}
	m.fail = true
	a := NewAppWithEngine(testCfg, testCampaign, context.Background(), m, nil)
	model, _ := a.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	a = model.(App)
	send := func(input string) {
		t.Helper()
		model, cmd := a.Update(narrative.SubmitMsg{Input: input})
		a = model.(App)
		if cmd != nil {
			model, _ = a.Update(cmd())
			a = model.(App)
		}
	}
	actor := uuid.New()
	send("/confirm-faith")
	if len(m.requests) != 0 {
		t.Fatal("unprepared confirmation")
	}
	send("/faith " + actor.String() + " " + actor.String())
	if a.faithRequest == nil || len(m.requests) != 0 {
		t.Fatal("preview mutated or unavailable")
	}
	send("/confirm-faith")
	if a.faithRequest == nil || len(m.requests) != 1 {
		t.Fatal("uncertain request lost")
	}
	m.fail = false
	send("/confirm-faith")
	if len(m.requests) != 2 || m.requests[0] != m.requests[1] || a.faithRequest != nil {
		t.Fatal("retry identity changed")
	}
	send("/end-faith " + actor.String())
	if a.faithRequest == nil || a.faithRequest.TargetID != uuid.Nil {
		t.Fatal("ending not prepared")
	}
	send("/cancel-faith")
	if a.faithRequest != nil || len(m.requests) != 2 {
		t.Fatal("cancel committed")
	}
}
