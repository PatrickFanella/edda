package handlers

import (
	"bytes"
	"encoding/json"
	"testing"

	"git.subcult.tv/subculture-collective/edda/internal/engine"
)

func TestSRDItemReceiptsExcludePrivateEnvelopes(t *testing.T) {
	for _, tc := range []struct{ tool, kind string }{
		{"srd_equipment", "equipment-equip"}, {"srd_effect", "rest-long"}, {"srd_spell", "spell-cast"}, {"srd_attack", "martial-attack"}, {"srd_consumable", "consumable-use"},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			call := engine.AppliedToolCall{Tool: tc.tool, Arguments: json.RawMessage(`{"secret":"hidden ambusher"}`), Result: mustJSON(t, map[string]any{"secret": "hidden ambusher", "receipt": map[string]any{"rules_version": "srd-5.2.1", "kind": tc.kind, "outcome": "hit", "secret": "hidden ambusher"}})}
			events := engineResolutionEventsToAPI([]engine.AppliedToolCall{call})
			if len(events) != 1 {
				t.Fatalf("events=%+v", events)
			}
			if bytes.Contains(mustJSON(t, events), []byte("hidden ambusher")) {
				t.Fatal("private data leaked")
			}
			history := engineResolutionEventsToAPI(appliedToolCallsFromSessionLog(mustJSON(t, []engine.AppliedToolCall{call})))
			if !bytes.Equal(mustJSON(t, events), mustJSON(t, history)) {
				t.Fatal("history receipt differs")
			}
		})
	}
}

func TestAllySpellReceiptProjectionAndHistory(t *testing.T) {
	call := engine.AppliedToolCall{Tool: "srd_spell", Arguments: json.RawMessage(`{"secret":"private notes"}`), Result: json.RawMessage(`{"secret":"private notes","receipt":{"rules_version":"srd-5.2.1","kind":"spell-cast","spell":"Cure Wounds","target_id":"known-ally","target":"Rowan","slot_level":1,"slots_before":2,"slots_after":1,"dice":[2,2],"rolled_healing":7,"healing":2,"hp_before":8,"hp_after":10,"secret":"private notes"}}`)}
	live := engineResolutionEventsToAPI([]engine.AppliedToolCall{call})
	history := engineResolutionEventsToAPI(appliedToolCallsFromSessionLog(mustJSON(t, []engine.AppliedToolCall{call})))
	data := mustJSON(t, live)
	if len(live) != 1 || !bytes.Equal(data, mustJSON(t, history)) || !bytes.Contains(data, []byte(`"target":"Rowan"`)) || !bytes.Contains(data, []byte(`"dice":[2,2]`)) || bytes.Contains(data, []byte("private notes")) {
		t.Fatalf("unsafe or changed ally receipt: %s", data)
	}
}
