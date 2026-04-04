package campaign_test

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackc/pgx/v5/pgtype"

	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
	"github.com/PatrickFanella/game-master/tui/campaign"
)

func makeUUID(b byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{b}, Valid: true}
}

func makeCampaign(id byte, name, status string) statedb.Campaign {
	return statedb.Campaign{
		ID:     makeUUID(id),
		Name:   name,
		Status: status,
		Genre:  pgtype.Text{String: "Fantasy", Valid: true},
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC),
			Valid: true,
		},
		CreatedBy: makeUUID(99),
	}
}

var _ tea.Model = campaign.Picker{}

func TestNewPicker_RendersCampaignsPlusNewOption(t *testing.T) {
	m := campaign.NewPicker([]statedb.Campaign{
		makeCampaign(1, "Alpha", "active"),
		makeCampaign(2, "Beta", "paused"),
	})
	m.SetSize(120, 40)

	view := m.View()
	for _, want := range []string{"Alpha", "Beta", "✦ New Campaign"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q: %q", want, view)
		}
	}
}

func TestNewPicker_EmptyListStillShowsNewOption(t *testing.T) {
	m := campaign.NewPicker(nil)
	m.SetSize(120, 40)

	view := m.View()
	if !strings.Contains(view, "✦ New Campaign") {
		t.Fatalf("view missing new campaign option: %q", view)
	}
}

func TestNewPicker_DescriptionIncludesGenreLastPlayedAndStatus(t *testing.T) {
	m := campaign.NewPicker([]statedb.Campaign{
		makeCampaign(1, "Alpha", "active"),
	})
	m.SetSize(120, 40)

	view := m.View()
	for _, want := range []string{"Genre: Fantasy", "Last played: 2026-03-30", "Status: active"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q: %q", want, view)
		}
	}
}

func TestPicker_InitReturnsNil(t *testing.T) {
	m := campaign.NewPicker(nil)
	if m.Init() != nil {
		t.Fatal("Init() should return nil")
	}
}

func TestPicker_UpdateEnterOnExistingCampaignReturnsSelectedMsg(t *testing.T) {
	c := makeCampaign(1, "My Campaign", "active")
	m := campaign.NewPicker([]statedb.Campaign{c})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected selection command")
	}

	msg := cmd()
	sel, ok := msg.(campaign.SelectedMsg)
	if !ok {
		t.Fatalf("expected SelectedMsg, got %T", msg)
	}
	if sel.Campaign.ID != c.ID {
		t.Fatalf("selected wrong campaign: got %v want %v", sel.Campaign.ID, c.ID)
	}
}

func TestPicker_UpdateEnterOnNewCampaignReturnsNewCampaignMsg(t *testing.T) {
	m := campaign.NewPicker([]statedb.Campaign{makeCampaign(1, "Alpha", "active")})

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if cmd != nil {
		_ = cmd
	}

	picker, ok := next.(campaign.Picker)
	if !ok {
		t.Fatalf("expected Picker after moving selection, got %T", next)
	}

	_, cmd = picker.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected selection command")
	}

	msg := cmd()
	if _, ok := msg.(campaign.NewCampaignMsg); !ok {
		t.Fatalf("expected NewCampaignMsg, got %T", msg)
	}
}

func TestPicker_SetSizeKeepsViewNonEmpty(t *testing.T) {
	m := campaign.NewPicker(nil)
	m.SetSize(80, 24)

	if view := m.View(); strings.TrimSpace(view) == "" {
		t.Fatal("View() should remain non-empty after SetSize")
	}
}
