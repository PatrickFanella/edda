package srd521

import (
	"encoding/json"
	"testing"
)

func TestActionSurgeScopeAndRecovery(t *testing.T) {
	creation, err := SupportedBuildForConcept("Human Fighter", "Soldier")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := creation.Stats()
	if err != nil {
		t.Fatal(err)
	}
	var e map[string]any
	if err := json.Unmarshal(raw, &e); err != nil {
		t.Fatal(err)
	}
	var b Build
	if err := json.Unmarshal(mustSurgeJSON(t, e["srd521"]), &b); err != nil {
		t.Fatal(err)
	}
	sheet, err := Project(raw, 1, 12, 12, 300)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AdvanceFighter(&b, sheet, "fixed", nil); err != nil {
		t.Fatal(err)
	}
	e["srd521"] = b
	raw = mustSurgeJSON(t, e)
	if !SupportsActionSurge(raw, 2) || SupportsActionSurge(raw, 1) || SupportsActionSurge(raw, 3) {
		t.Fatal("level scope")
	}
	projected, err := Project(raw, 2, 12, 20, 300)
	if err != nil {
		t.Fatal(err)
	}
	if projected.Resources[len(projected.Resources)-1].Current != 1 {
		t.Fatal("legacy initial use")
	}
	if _, err := UseActionSurge(&b); err != nil {
		t.Fatal(err)
	}
	if _, err := UseActionSurge(&b); err == nil {
		t.Fatal("duplicate use")
	}
	for _, kind := range []string{ShortRest, LongRest} {
		b.ActionSurgeSpent = true
		r, err := RestActionSurge(&b, projected, kind, "", 28800)
		if err != nil || r.ActionSurgeBefore != 0 || r.ActionSurgeAfter != 1 || b.ActionSurgeSpent {
			t.Fatalf("rest %+v %v", r, err)
		}
	}
	if _, err := RestActionSurge(&b, projected, LongRest, "", 57600); err == nil {
		t.Fatal("early rest")
	}
	for _, kind := range []string{WaitMinute, EndConcentration, "unsupported"} {
		if _, err := RestActionSurge(&b, projected, kind, "", 100000); err == nil {
			t.Fatal("unsupported recovery")
		}
	}
	e["srd521_creation"].(map[string]any)["background"] = "Acolyte"
	if SupportsActionSurge(mustSurgeJSON(t, e), 2) {
		t.Fatal("background scope")
	}
}
func mustSurgeJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
