package quest

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Compile-time checks: *Model must satisfy tea.Model and the sub-model
// interface (tea.Model + SetSize) expected by the Router. Importing the parent
// tui package here would create a cycle, so the interface is inlined.
var _ tea.Model = (*Model)(nil)
var _ interface {
	tea.Model
	SetSize(width, height int)
} = (*Model)(nil)

func TestNewReturnsModel(t *testing.T) {
	m := New()
	if m.width != 0 || m.height != 0 {
		t.Fatal("expected zero dimensions on a freshly created model")
	}
}

func TestNewHasPlaceholderQuest(t *testing.T) {
	m := New()
	if len(m.Quests) == 0 {
		t.Fatal("expected at least one placeholder quest")
	}
}

func TestSetSizeUpdatesDimensions(t *testing.T) {
	m := New()
	m.SetSize(80, 24)
	if m.width != 80 || m.height != 24 {
		t.Fatalf("expected 80x24, got %dx%d", m.width, m.height)
	}
}

func TestInitReturnsNil(t *testing.T) {
	m := New()
	if m.Init() != nil {
		t.Fatal("Init() should return nil")
	}
}

func TestEscEmitsNavigateBackMsg(t *testing.T) {
	m := New()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected a command on Esc, got nil")
	}
	msg := cmd()
	if _, ok := msg.(NavigateBackMsg); !ok {
		t.Fatalf("expected NavigateBackMsg, got %T", msg)
	}
}

func TestOtherKeyReturnsNoCmd(t *testing.T) {
	m := New()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("expected no command for non-Esc keys")
	}
}

func TestViewContainsQuestLogHeader(t *testing.T) {
	m := New()
	m.SetSize(80, 24)
	got := m.View()
	if !strings.Contains(got, "Quest Log") {
		t.Fatalf("expected View to contain %q, got:\n%s", "Quest Log", got)
	}
}

func TestViewContainsPlaceholderQuest(t *testing.T) {
	m := New()
	m.SetSize(80, 24)
	got := m.View()
	if !strings.Contains(got, "Missing Merchant") {
		t.Fatalf("expected View to contain placeholder quest name, got:\n%s", got)
	}
}

func TestViewEmptyQuestsShowsNoActiveQuests(t *testing.T) {
	m := New()
	m.Quests = nil
	m.SetSize(80, 24)
	got := m.View()
	if !strings.Contains(got, "No active quests") {
		t.Fatalf("expected View to contain %q when no quests, got:\n%s", "No active quests", got)
	}
}
