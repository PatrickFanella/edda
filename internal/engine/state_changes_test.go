package engine

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func applied(tool string, result string) AppliedToolCall {
	return AppliedToolCall{Tool: tool, Result: json.RawMessage(result)}
}

func TestStateChangesFromAppliedToolCalls(t *testing.T) {
	npcID := uuid.New()
	factID := uuid.New()
	questID := uuid.New()
	locationID := uuid.New()
	playerID := uuid.New()
	playerHPID := uuid.New()
	combatID := uuid.New()

	cases := []struct {
		name   string
		call   AppliedToolCall
		entity string
		id     uuid.UUID
		field  string
	}{
		{"create_npc", applied("create_npc", `{"id":"`+npcID.String()+`","campaign_id":"campaign","name":"Seraphine Maw","description":"Seer","personality":"Calm","disposition":0,"alive":true,"stats":{},"properties":{},"location_id":"`+locationID.String()+`"}`), "npc", npcID, "created"},
		{"establish_fact", applied("establish_fact", `{"id":"`+factID.String()+`","campaign_id":"campaign","fact":"The veil is thin here.","category":"lore","source":"established"}`), "world_fact", factID, "created"},
		{"create_quest", applied("create_quest", `{"id":"`+questID.String()+`","campaign_id":"campaign","title":"Escape the Hollow Spire","description":"Leave at dawn.","quest_type":"short_term","status":"active","objectives":[],"related_entities":[],"summary":"Quest"}`), "quest", questID, "created"},
		{"create_location", applied("create_location", `{"id":"`+locationID.String()+`","campaign_id":"campaign","name":"Old Road","description":"A weathered road","region":"Borderlands","location_type":"wilderness","properties":{},"connections":[]}`), "location", locationID, "created"},
		{"reveal_location", applied("reveal_location", `{"location_id":"`+locationID.String()+`","campaign_id":"campaign","name":"Old Road"}`), "location", locationID, "revealed"},
		{"update_player_status", applied("update_player_status", `{"player_character_id":"`+playerID.String()+`","status":"cursed","statuses":[{"status":"cursed"}],"duration":{"unit":"turns","value":"2"}}`), "player_character", playerID, "status_updated"},
		{"update_player_hp", applied("update_player_hp", `{"player_character_id":"`+playerHPID.String()+`","hp":7,"max_hp":20,"old_hp":12,"old_max_hp":20}`), "player_character", playerHPID, "hp_updated"},
		{"initiate_combat", applied("initiate_combat", `{"combat_state_id":"`+combatID.String()+`","initiative_order":[],"environment":"trail","surprise":"players","enemy_count":1,"opening_description":"Combat begins!","mode":"combat"}`), "combat", combatID, "started"},
	}

	for _, tc := range cases {
		changes := StateChangesFromAppliedToolCalls([]AppliedToolCall{tc.call})
		if len(changes) != 1 {
			t.Fatalf("%s: len(changes) = %d, want 1", tc.name, len(changes))
		}
		got := changes[0]
		if got.Entity != tc.entity || got.EntityID != tc.id || got.Field != tc.field {
			t.Fatalf("%s: change = %#v, want entity=%q id=%s field=%q", tc.name, got, tc.entity, tc.id, tc.field)
		}
		if len(got.NewValue) == 0 {
			t.Fatalf("%s: missing NewValue", tc.name)
		}
	}
}

func TestStateChangesFromAppliedToolCalls_MovePlayerEmitsConsumerCompatibleChanges(t *testing.T) {
	locationID := uuid.New()
	playerID := uuid.New()
	changes := StateChangesFromAppliedToolCalls([]AppliedToolCall{
		applied("move_player", `{"location_id":"`+locationID.String()+`","campaign_id":"campaign","player_character_id":"`+playerID.String()+`","name":"Old Road","description":"A weathered road","location_type":"wilderness","travel_time":"1 hour","day":1,"hour":9,"minute":0,"visited_marked":true,"visited_warning":"","time_warning":""}`),
	})
	if len(changes) != 3 {
		t.Fatalf("len(changes) = %d, want 3: %#v", len(changes), changes)
	}
	want := map[string]struct{}{
		"player_character:" + playerID.String() + ":location_updated": {},
		"location:" + locationID.String() + ":updated":                {},
		"location:" + locationID.String() + ":moved":                  {},
	}
	for _, change := range changes {
		key := change.Entity + ":" + change.EntityID.String() + ":" + change.Field
		if _, ok := want[key]; !ok {
			t.Fatalf("unexpected change %#v", change)
		}
		delete(want, key)
	}
	if len(want) != 0 {
		t.Fatalf("missing changes: %#v", want)
	}
}

func TestStateChangesFromAppliedToolCalls_CreateLocationWithMoveEmitsMovementChanges(t *testing.T) {
	locationID := uuid.New()
	playerID := uuid.New()
	changes := StateChangesFromAppliedToolCalls([]AppliedToolCall{
		applied("create_location", `{"id":"`+locationID.String()+`","location_id":"`+locationID.String()+`","campaign_id":"campaign","player_character_id":"`+playerID.String()+`","name":"Daylight Maintenance Ledge","description":"A ledge","region":"Upper Drainage","location_type":"ledge","properties":{},"connections":[],"reused":false,"move_player_here":true,"visited_marked":true}`),
	})
	if len(changes) != 4 {
		t.Fatalf("len(changes) = %d, want 4: %#v", len(changes), changes)
	}
	want := map[string]struct{}{
		"location:" + locationID.String() + ":created":                {},
		"player_character:" + playerID.String() + ":location_updated": {},
		"location:" + locationID.String() + ":updated":                {},
		"location:" + locationID.String() + ":moved":                  {},
	}
	for _, change := range changes {
		key := change.Entity + ":" + change.EntityID.String() + ":" + change.Field
		if _, ok := want[key]; !ok {
			t.Fatalf("unexpected change %#v", change)
		}
		delete(want, key)
	}
	if len(want) != 0 {
		t.Fatalf("missing changes: %#v", want)
	}
}

func TestStateChangesFromAppliedToolCalls_ReusedCreateLocationDoesNotEmitCreated(t *testing.T) {
	locationID := uuid.New()
	changes := StateChangesFromAppliedToolCalls([]AppliedToolCall{
		applied("create_location", `{"id":"`+locationID.String()+`","campaign_id":"campaign","name":"Daylight Maintenance Ledge","description":"A ledge","region":"Upper Drainage","location_type":"ledge","properties":{},"connections":[],"reused":true}`),
	})
	if len(changes) != 0 {
		t.Fatalf("changes = %#v, want none", changes)
	}
}

func TestStateChangesFromAppliedToolCalls_UnknownOrMalformedSkipped(t *testing.T) {
	changes := StateChangesFromAppliedToolCalls([]AppliedToolCall{
		applied("unknown_tool", `{}`),
		applied("create_npc", `{bad json`),
	})
	if len(changes) != 0 {
		t.Fatalf("changes = %#v, want none", changes)
	}
}
