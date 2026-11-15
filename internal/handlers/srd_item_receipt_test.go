package handlers

import (
	"bytes"
	"encoding/json"
	"testing"

	"git.subcult.tv/subculture-collective/edda/internal/engine"
)

func TestSRDItemReceiptsExcludePrivateEnvelopes(t *testing.T) {
	for _, tc := range []struct{ tool, kind string }{
		{"srd_spell", "spell-cast"}, {"srd_attack", "martial-attack"}, {"srd_consumable", "consumable-use"},
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
