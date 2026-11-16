package handlers

import (
	"encoding/json"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	"strings"
	"testing"
)

func TestActionSurgePublicReceipt(t *testing.T) {
	payload := json.RawMessage(`{"receipt":{"rules_version":"srd-5.2.1","kind":"action-surge","uses_before":1,"uses_after":0,"additional_actions":1,"excluded_action":"Magic","source_ids":["srd-5.2.1:content/fighter/action-surge"],"secret":"hidden"},"encounter":{"npc_tactic":"secret"}}`)
	events := engineResolutionEventsToAPI([]engine.AppliedToolCall{{Tool: "srd_action_surge", Result: payload}})
	if len(events) != 1 {
		t.Fatal(events)
	}
	if _, ok := events[0].Details["receipt"].(*srd521.ActionSurgeReceipt); !ok {
		t.Fatal("untyped receipt")
	}
	data, err := json.Marshal(events)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "secret") || strings.Contains(string(data), "hidden") {
		t.Fatal("private payload disclosed")
	}
}
