package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type stubView struct {
	width, height int
	viewText      string
}

func (s *stubView) Init() tea.Cmd                           { return nil }
func (s *stubView) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return s, nil }
func (s *stubView) View() string                            { return s.viewText }
func (s *stubView) SetSize(w, h int)                        { s.width = w; s.height = h }

func TestRouterRegisterAndCount(t *testing.T) {
	r := NewRouter()
	if r.TabCount() != 0 {
		t.Fatal("expected 0 tabs")
	}
	r.Register("A", &stubView{})
	r.Register("B", &stubView{})
	if r.TabCount() != 2 {
		t.Fatalf("expected 2 tabs, got %d", r.TabCount())
	}
}

func TestRouterNextTab(t *testing.T) {
	r := NewRouter()
	r.Register("A", &stubView{})
	r.Register("B", &stubView{})
	r.Register("C", &stubView{})

	if r.ActiveTab() != 0 {
		t.Fatal("expected 0")
	}
	r.NextTab()
	if r.ActiveTab() != 1 {
		t.Fatal("expected 1")
	}
	r.NextTab()
	if r.ActiveTab() != 2 {
		t.Fatal("expected 2")
	}
	r.NextTab()
	if r.ActiveTab() != 0 {
		t.Fatal("expected wrap to 0")
	}
}

func TestRouterPrevTab(t *testing.T) {
	r := NewRouter()
	r.Register("A", &stubView{})
	r.Register("B", &stubView{})

	r.PrevTab()
	if r.ActiveTab() != 1 {
		t.Fatal("expected wrap to 1")
	}
	r.PrevTab()
	if r.ActiveTab() != 0 {
		t.Fatal("expected 0")
	}
}

func TestRouterGoToTab(t *testing.T) {
	r := NewRouter()
	r.Register("A", &stubView{})
	r.Register("B", &stubView{})
	r.Register("C", &stubView{})

	r.GoToTab(2)
	if r.ActiveTab() != 2 {
		t.Fatal("expected 2")
	}
	r.GoToTab(-1) // out of range, no-op
	if r.ActiveTab() != 2 {
		t.Fatal("expected still 2")
	}
	r.GoToTab(99) // out of range, no-op
	if r.ActiveTab() != 2 {
		t.Fatal("expected still 2")
	}
}

func TestRouterSetSizePropagates(t *testing.T) {
	s1 := &stubView{}
	s2 := &stubView{}
	r := NewRouter()
	r.Register("A", s1)
	r.Register("B", s2)

	r.SetSize(120, 40)
	if s1.width != 120 || s1.height != 40 {
		t.Fatalf("s1 size: %dx%d", s1.width, s1.height)
	}
	if s2.width != 120 || s2.height != 40 {
		t.Fatalf("s2 size: %dx%d", s2.width, s2.height)
	}
}

func TestRouterViewRendersActive(t *testing.T) {
	r := NewRouter()
	r.Register("A", &stubView{viewText: "view-a"})
	r.Register("B", &stubView{viewText: "view-b"})

	if r.View() != "view-a" {
		t.Fatalf("expected view-a, got %q", r.View())
	}
	r.NextTab()
	if r.View() != "view-b" {
		t.Fatalf("expected view-b, got %q", r.View())
	}
}

func TestRouterViewEmpty(t *testing.T) {
	r := NewRouter()
	if r.View() != "" {
		t.Fatal("expected empty string for no tabs")
	}
}

func TestRouterTabs(t *testing.T) {
	r := NewRouter()
	r.Register("Narrative", &stubView{})
	r.Register("Character", &stubView{})

	tabs := r.Tabs()
	if len(tabs) != 2 {
		t.Fatalf("expected 2 tabs, got %d", len(tabs))
	}
	if tabs[0].Name != "Narrative" {
		t.Fatalf("expected Narrative, got %q", tabs[0].Name)
	}
	if tabs[1].Name != "Character" {
		t.Fatalf("expected Character, got %q", tabs[1].Name)
	}
}
