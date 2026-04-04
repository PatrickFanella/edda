package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"

	"github.com/PatrickFanella/game-master/internal/config"
	"github.com/PatrickFanella/game-master/internal/dbutil"
	"github.com/PatrickFanella/game-master/internal/domain"
	"github.com/PatrickFanella/game-master/internal/engine"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
	"github.com/PatrickFanella/game-master/tui/narrative"
)

// Compile-time check: App must implement tea.Model.
var _ tea.Model = App{}

// testCfg is a minimal config suitable for unit tests.
var testCfg = config.Config{
	LLM: config.LLMConfig{Provider: "ollama"},
}

// testCampaign is a zero-value campaign used in unit tests.
var testCampaign = statedb.Campaign{}

type mockGameEngine struct {
	processTurnFn     func(context.Context, uuid.UUID, string) (*engine.TurnResult, error)
	loadCampaignFn    func(context.Context, uuid.UUID) error
	inputs            []string
	campaignIDs       []uuid.UUID
	loadedCampaignIDs []uuid.UUID
}

func (m *mockGameEngine) ProcessTurn(ctx context.Context, campaignID uuid.UUID, input string) (*engine.TurnResult, error) {
	m.inputs = append(m.inputs, input)
	m.campaignIDs = append(m.campaignIDs, campaignID)
	if m.processTurnFn != nil {
		return m.processTurnFn(ctx, campaignID, input)
	}
	return &engine.TurnResult{}, nil
}

func (m *mockGameEngine) GetGameState(context.Context, uuid.UUID) (*engine.GameState, error) {
	return &engine.GameState{}, nil
}

func (m *mockGameEngine) NewCampaign(context.Context, uuid.UUID) (*domain.Campaign, error) {
	return nil, errors.New("not implemented")
}

func (m *mockGameEngine) LoadCampaign(ctx context.Context, campaignID uuid.UUID) error {
	m.loadedCampaignIDs = append(m.loadedCampaignIDs, campaignID)
	if m.loadCampaignFn != nil {
		return m.loadCampaignFn(ctx, campaignID)
	}
	return nil
}

func (m *mockGameEngine) ProcessTurnStream(_ context.Context, _ uuid.UUID, _ string) (<-chan engine.StreamEvent, error) {
	ch := make(chan engine.StreamEvent)
	close(ch)
	return ch, nil
}

func keyForView(view ViewState) rune {
	// ViewState is zero-based; user-facing key bindings are 1-based.
	return rune('1' + view)
}

func TestViewStateConstants(t *testing.T) {
	if ViewNarrative != 0 {
		t.Fatalf("ViewNarrative should be 0, got %d", ViewNarrative)
	}
	if ViewCharacterSheet != 1 {
		t.Fatalf("ViewCharacterSheet should be 1, got %d", ViewCharacterSheet)
	}
	if ViewInventory != 2 {
		t.Fatalf("ViewInventory should be 2, got %d", ViewInventory)
	}
	if ViewQuestLog != 3 {
		t.Fatalf("ViewQuestLog should be 3, got %d", ViewQuestLog)
	}

	if ViewLogs != 4 {
		t.Fatalf("ViewLogs should be 4, got %d", ViewLogs)
	}
}

func TestNewAppRegistersAllViews(t *testing.T) {
	app := NewApp(testCfg, testCampaign)
	if app.router.TabCount() != 5 {
		t.Fatalf("expected 5 registered views, got %d", app.router.TabCount())
	}
}

func TestNewAppStartsOnNarrative(t *testing.T) {
	app := NewApp(testCfg, testCampaign)
	if app.ActiveViewState() != ViewNarrative {
		t.Fatalf("expected initial ViewState %d (ViewNarrative), got %d",
			ViewNarrative, app.ActiveViewState())
	}
}

func TestAppTabNamesMatchViewStates(t *testing.T) {
	app := NewApp(testCfg, testCampaign)
	tabs := app.router.Tabs()
	expected := []string{"Narrative", "Character", "Inventory", "Quests", "Logs"}
	for i, name := range expected {
		if tabs[i].Name != name {
			t.Errorf("tab[%d]: expected %q, got %q", i, name, tabs[i].Name)
		}
	}
}

func TestAppInitReturnsNil(t *testing.T) {
	app := NewApp(testCfg, testCampaign)
	if app.Init() != nil {
		t.Fatal("Init() should return nil")
	}
}

func TestAppUpdateQuitCtrlC(t *testing.T) {
	app := NewApp(testCfg, testCampaign)
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected quit command for ctrl+c, got nil")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("expected tea.QuitMsg for ctrl+c")
	}
}

func TestAppUpdateQuitQOutsideSuppressedView(t *testing.T) {
	app := NewApp(testCfg, testCampaign)
	app.router.GoToTab(int(ViewCharacterSheet))

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatal("expected quit command for q outside narrative input, got nil")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("expected tea.QuitMsg for q outside narrative input")
	}
}

func TestAppUpdateFocusedNarrativeSuppressesConflictingRunes(t *testing.T) {
	app := NewApp(testCfg, testCampaign)
	sized, _ := app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	updated := sized.(App)

	for _, r := range []rune{'h', '1', 'l', 'q', '5'} {
		model, cmd := updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		if cmd != nil {
			if _, ok := cmd().(tea.QuitMsg); ok {
				t.Fatalf("expected %q to reach focused narrative input, got quit command", r)
			}
		}
		updated = model.(App)
		if updated.ActiveViewState() != ViewNarrative {
			t.Fatalf("expected narrative to stay active after %q, got %d", r, updated.ActiveViewState())
		}
	}

	if !strings.Contains(updated.View(), "h1lq5") {
		t.Fatalf("expected focused narrative input to keep conflicting runes, view=%q", updated.View())
	}
}

func TestAppUpdateFocusedNarrativeSuppressesArrowNavigation(t *testing.T) {
	app := NewApp(testCfg, testCampaign)
	sized, _ := app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	updated := sized.(App)

	for _, r := range []rune{'a', 'b'} {
		model, _ := updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		updated = model.(App)
	}

	model, _ := updated.Update(tea.KeyMsg{Type: tea.KeyLeft})
	updated = model.(App)
	if updated.ActiveViewState() != ViewNarrative {
		t.Fatalf("expected narrative to stay active after left arrow, got %d", updated.ActiveViewState())
	}

	model, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	updated = model.(App)

	model, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRight})
	updated = model.(App)
	if updated.ActiveViewState() != ViewNarrative {
		t.Fatalf("expected narrative to stay active after right arrow, got %d", updated.ActiveViewState())
	}

	model, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	updated = model.(App)

	if !strings.Contains(updated.View(), "acbd") {
		t.Fatalf("expected arrow keys to move the narrative cursor, view=%q", updated.View())
	}
}

func TestAppUpdateTabStillCyclesWhenNarrativeInputFocused(t *testing.T) {
	app := NewApp(testCfg, testCampaign)
	sized, _ := app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	updated := sized.(App)

	model, _ := updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	updated = model.(App)

	model, cmd := updated.Update(tea.KeyMsg{Type: tea.KeyTab})
	if cmd != nil {
		t.Fatal("expected tab to remain a global shortcut")
	}
	updated = model.(App)
	if updated.ActiveViewState() != ViewCharacterSheet {
		t.Fatalf("expected tab to switch to character sheet, got %d", updated.ActiveViewState())
	}
}

func TestAppUpdateTabNextView(t *testing.T) {
	app := NewApp(testCfg, testCampaign)
	m, _ := app.Update(tea.KeyMsg{Type: tea.KeyTab})
	updated := m.(App)
	if updated.ActiveViewState() != ViewCharacterSheet {
		t.Fatalf("expected ViewCharacterSheet after tab, got %d", updated.ActiveViewState())
	}
}

func TestAppUpdateTabWrapsAround(t *testing.T) {
	app := NewApp(testCfg, testCampaign)
	// Advance to the last view (Logs = index 4).
	app.router.GoToTab(4)
	m, _ := app.Update(tea.KeyMsg{Type: tea.KeyTab})
	updated := m.(App)
	if updated.ActiveViewState() != ViewNarrative {
		t.Fatalf("expected ViewNarrative after wrapping tab, got %d", updated.ActiveViewState())
	}
}

func TestAppUpdateShiftTabCyclesBackward(t *testing.T) {
	app := NewApp(testCfg, testCampaign)
	// shift+tab from Narrative wraps to Logs.
	m, _ := app.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	updated := m.(App)
	if updated.ActiveViewState() != ViewLogs {
		t.Fatalf("expected ViewLogs after shift+tab wrap, got %d", updated.ActiveViewState())
	}
}

func TestAppUpdateShiftTabPrevView(t *testing.T) {
	app := NewApp(testCfg, testCampaign)
	app.router.GoToTab(2)
	m, _ := app.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	updated := m.(App)
	if updated.ActiveViewState() != ViewCharacterSheet {
		t.Fatalf("expected ViewCharacterSheet after shift+tab, got %d", updated.ActiveViewState())
	}
}

func TestAppUpdateNumberKeysOutsideSuppressedView(t *testing.T) {
	tests := []struct {
		key      rune
		expected ViewState
	}{
		{'1', ViewNarrative},
		{'2', ViewCharacterSheet},
		{'3', ViewInventory},
		{'4', ViewQuestLog},
		{'5', ViewLogs},
	}
	for _, tt := range tests {
		app := NewApp(testCfg, testCampaign)
		app.router.GoToTab(int(ViewCharacterSheet))
		m, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}})
		updated := m.(App)
		if updated.ActiveViewState() != tt.expected {
			t.Errorf("key %q from non-narrative view: expected ViewState %d, got %d", tt.key, tt.expected, updated.ActiveViewState())
		}
	}
}

func TestAppUpdateViewSwitchingPreservesState(t *testing.T) {
	// Sub-model state should not be reset when switching between views.
	app := NewApp(testCfg, testCampaign)

	// The narrative view was seeded with entries in NewApp; verify the router
	// still holds those entries after switching away and back.
	narrativeViewBefore := app.router.Tabs()[0].View

	// Switch away to character sheet, then back to narrative, using the always-global tab bindings.
	m, _ := app.Update(tea.KeyMsg{Type: tea.KeyTab})
	updated := m.(App)
	if updated.ActiveViewState() != ViewCharacterSheet {
		t.Fatalf("expected ViewCharacterSheet, got %d", updated.ActiveViewState())
	}

	m2, _ := updated.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	backToNarrative := m2.(App)

	narrativeViewAfter := backToNarrative.router.Tabs()[0].View
	if narrativeViewBefore != narrativeViewAfter {
		t.Fatal("narrative sub-model was replaced when switching views (state not preserved)")
	}
}

func TestAppWindowSizeUpdatesState(t *testing.T) {
	app := NewApp(testCfg, testCampaign)
	m, _ := app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	updated := m.(App)
	if updated.width != 120 || updated.height != 40 {
		t.Fatalf("expected 120x40, got %dx%d", updated.width, updated.height)
	}
}

func TestAppViewReturnsNonEmpty(t *testing.T) {
	app := NewApp(testCfg, testCampaign)
	v := app.View()
	if v == "" {
		t.Fatal("View() should return non-empty string")
	}
}

func TestStatusBarShowsViewsHintsAndActiveView(t *testing.T) {
	app := NewApp(testCfg, testCampaign)
	_, statusBar := app.chrome()

	for _, label := range []string{"Narrative", "Character", "Inventory", "Quests", "Logs"} {
		if !strings.Contains(statusBar, label) {
			t.Fatalf("expected status bar to include %q", label)
		}
	}
	if !strings.Contains(statusBar, "[Narrative]") {
		t.Fatal("expected status bar to highlight the active narrative view")
	}
	if !strings.Contains(statusBar, statusBarHints) {
		t.Fatal("expected status bar to include view switching key hints")
	}
}

func TestStatusBarUpdatesImmediatelyOnViewSwitch(t *testing.T) {
	app := NewApp(testCfg, testCampaign)
	targetInventoryKey := keyForView(ViewInventory)

	// Move off the focused narrative input first so the numeric shortcut is handled
	// globally rather than being consumed by text entry.
	m, _ := app.Update(tea.KeyMsg{Type: tea.KeyTab})
	updated := m.(App)
	if updated.ActiveViewState() != ViewCharacterSheet {
		t.Fatalf("expected tab to move to character sheet first, got %d", updated.ActiveViewState())
	}

	m, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{targetInventoryKey}})
	updated = m.(App)
	_, statusBar := updated.chrome()

	if !strings.Contains(statusBar, "[Inventory]") {
		t.Fatalf("expected status bar to highlight inventory after pressing %q", targetInventoryKey)
	}
	if strings.Contains(statusBar, "[Narrative]") {
		t.Fatalf("expected narrative to no longer be active after pressing %q", targetInventoryKey)
	}

	m2, _ := updated.Update(tea.KeyMsg{Type: tea.KeyTab})
	updated2 := m2.(App)
	_, statusBar2 := updated2.chrome()
	if !strings.Contains(statusBar2, "[Quests]") || !strings.Contains(statusBar2, "Logs") {
		t.Fatal("expected status bar to highlight quests after tab cycling")
	}
}

func TestAppOtherKeyDelegatedToSubView(t *testing.T) {
	// Non-global keys should be forwarded to the active sub-view.
	// The active view is a real narrative.Model; pressing Enter doesn't crash.
	app := NewApp(testCfg, testCampaign)
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// No panic = pass; the sub-view received the message.
}

func TestAppUnknownMsgForwardedToSubView(t *testing.T) {
	// Non-key, non-window messages (e.g. custom command results) must be
	// forwarded to the active sub-view rather than silently dropped.
	type customMsg struct{ value string }
	app := NewApp(testCfg, testCampaign)
	// Sending an unrecognised message type must not panic and must return a
	// well-formed model (not nil).
	m, _ := app.Update(customMsg{"hello"})
	if m == nil {
		t.Fatal("Update should return a non-nil model for unknown message types")
	}
	if _, ok := m.(App); !ok {
		t.Fatal("Update should return an App model for unknown message types")
	}
}

func TestAppSubmitCallsEngineAndStreamsNarrativeWithChoices(t *testing.T) {
	campaignID := uuid.New()
	mockEngine := &mockGameEngine{
		processTurnFn: func(_ context.Context, gotCampaignID uuid.UUID, gotInput string) (*engine.TurnResult, error) {
			if gotCampaignID != campaignID {
				t.Fatalf("expected campaign id %s, got %s", campaignID, gotCampaignID)
			}
			if gotInput != "open the door" {
				t.Fatalf("expected input %q, got %q", "open the door", gotInput)
			}
			return &engine.TurnResult{
				Narrative: "The heavy oak door swings inward.",
				Choices: []engine.Choice{
					{ID: "enter", Text: "Step inside"},
					{ID: "listen", Text: "Listen at the threshold"},
				},
			}, nil
		},
	}

	app := NewAppWithEngine(testCfg, statedb.Campaign{ID: dbutil.ToPgtype(campaignID)}, context.Background(), mockEngine, nil)
	sized, _ := app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	app = sized.(App)
	var cmd tea.Cmd
	model, _ := app.Update(narrative.SubmitMsg{Input: "open the door"})
	updated := model.(App)

	if !updated.turnBusy {
		t.Fatal("expected app to remain busy while the turn is in flight")
	}
	if !strings.Contains(updated.View(), "Thinking…") {
		t.Fatal("expected loading indicator while processing")
	}

	msg := updated.processTurn("open the door")()
	model, cmd = updated.Update(msg)
	updated = model.(App)
	if cmd == nil {
		t.Fatal("expected narrative streaming command")
	}

	for updated.turnBusy {
		msg = cmd()
		model, cmd = updated.Update(msg)
		updated = model.(App)
	}

	if len(mockEngine.inputs) != 1 || mockEngine.inputs[0] != "open the door" {
		t.Fatalf("expected engine to be called once with player input, got %#v", mockEngine.inputs)
	}
	view := updated.View()
	if !strings.Contains(view, "The heavy oak door swings inward.") {
		t.Fatal("expected streamed narrative to appear in the view")
	}
	if !strings.Contains(view, "Suggested choices") || !strings.Contains(view, "Step inside") {
		t.Fatal("expected suggested choices to render below the narrative")
	}
}

func TestAppTurnErrorAddsSystemMessage(t *testing.T) {
	mockEngine := &mockGameEngine{
		processTurnFn: func(context.Context, uuid.UUID, string) (*engine.TurnResult, error) {
			return nil, errors.New("connection failed")
		},
	}

	app := NewAppWithEngine(testCfg, testCampaign, context.Background(), mockEngine, nil)
	sized, _ := app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	app = sized.(App)
	model, _ := app.Update(narrative.SubmitMsg{Input: "talk to innkeeper"})
	updated := model.(App)

	model, _ = updated.Update(updated.processTurn("talk to innkeeper")())
	updated = model.(App)

	if updated.turnBusy {
		t.Fatal("expected busy state to clear after error")
	}
	if !strings.Contains(updated.View(), "Error: connection failed") {
		t.Fatal("expected error state to be shown in the narrative viewport")
	}
}

func TestNextNarrativeChunkPreservesUTF8Runes(t *testing.T) {
	text := "🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂"

	chunk, remaining := nextNarrativeChunk(text)

	if chunk == "" || remaining == "" {
		t.Fatal("expected text to be split into two non-empty chunks")
	}
	if chunk != "🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂🙂" {
		t.Fatalf("unexpected first chunk: %q", chunk)
	}
	if remaining != "🙂" {
		t.Fatalf("unexpected remaining chunk: %q", remaining)
	}
}
