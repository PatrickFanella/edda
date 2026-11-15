package handlers

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
)

func TestSRDReceiptProjectionFiltersPrivateEnvelopes(t *testing.T) {
	data, err := os.ReadFile("../../frontend/tests/fixtures/srd-check-sheet.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct{ Receipts []api.CheckReceipt }
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, receipt := range fixture.Receipts {
		safe := mustJSON(t, receipt)
		var extra map[string]any
		if err := json.Unmarshal(safe, &extra); err != nil {
			t.Fatal(err)
		}
		extra["hidden_fact"] = "secret ambusher"
		extra["margin"] = -4
		call := engine.AppliedToolCall{
			Tool:      "srd_check",
			Arguments: mustJSON(t, map[string]any{"difficulty": 17, "reason": "secret ambusher", "hidden_entity_id": "secret-id"}),
			Result:    mustJSON(t, map[string]any{"receipt": extra, "dc": 17, "success": false, "margin": -4, "facts": "secret ambusher"}),
		}
		out := engineTurnResultToAPI(&engine.TurnResult{AppliedToolCalls: []engine.AppliedToolCall{call}})
		if len(out.ResolutionEvents) != 1 {
			t.Fatalf("missing SRD event: %+v", out.ResolutionEvents)
		}
		event := out.ResolutionEvents[0]
		if event.Type != "srd_check" || event.Outcome != receipt.Outcome || len(event.Details) != 1 {
			t.Fatalf("unexpected event: %+v", event)
		}
		got := mustJSON(t, event.Details["receipt"])
		if !bytes.Equal(safe, got) {
			t.Fatalf("receipt changed at projection boundary:\ngot %s\nwant %s", got, safe)
		}
		encoded := mustJSON(t, out)
		for _, secret := range [][]byte{[]byte("secret ambusher"), []byte("secret-id"), []byte(`"margin"`), []byte(`"difficulty"`), []byte(`"success":false`)} {
			if bytes.Contains(encoded, secret) {
				t.Fatalf("private context leaked through turn projection: %s", encoded)
			}
		}
		// History and the action receipt use the same event projection.
		history := engineResolutionEventsToAPI(appliedToolCallsFromSessionLog(mustJSON(t, []engine.AppliedToolCall{call})))
		if !bytes.Equal(mustJSON(t, history), mustJSON(t, out.ResolutionEvents)) {
			t.Fatal("history projection differs from live turn projection")
		}
	}
}

func TestSRDReceiptProjectionRejectsMissingOrUnsupportedReceipts(t *testing.T) {
	for _, raw := range []string{
		`null`, `{}`, `{"receipt":null}`, `{"receipt":{}}`,
		`{"receipt":{"rules_version":"legacy","kind":"ability-check","outcome":"success"}}`,
		`{"receipt":{"rules_version":"srd-5.2.1","kind":"attack","outcome":"success"}}`,
		`{"receipt":{"rules_version":"srd-5.2.1","kind":"ability-check","outcome":"secret ambusher"}}`,
		`not json`,
	} {
		if events := engineResolutionEventsToAPI([]engine.AppliedToolCall{{Tool: "srd_check", Result: json.RawMessage(raw)}}); len(events) != 0 {
			t.Fatalf("unsupported receipt accepted: %s", raw)
		}
	}
}
