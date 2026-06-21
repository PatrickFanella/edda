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
	objectiveID := uuid.New()
	locationID := uuid.New()
	playerID := uuid.New()
	playerXPID := uuid.New()
	playerLevelID := uuid.New()
	itemID := uuid.New()
	modifiedItemID := uuid.New()
	removedItemID := uuid.New()
	updatedQuestID := uuid.New()
	updatedNPCID := uuid.New()
	combatStateID := uuid.New()
	combatID := uuid.New()

	cases := []struct {
		name   string
		call   AppliedToolCall
		entity string
		id     uuid.UUID
		field  string
		count  int
	}{
		{"create_npc", applied("create_npc", `{"id":"`+npcID.String()+`","campaign_id":"campaign","name":"Seraphine Maw","description":"Seer","personality":"Calm","disposition":0,"alive":true,"stats":{},"properties":{},"location_id":"`+locationID.String()+`"}`), "npc", npcID, "created", 1},
		{"establish_fact", applied("establish_fact", `{"id":"`+factID.String()+`","campaign_id":"campaign","fact":"The veil is thin here.","category":"lore","source":"established"}`), "world_fact", factID, "created", 1},
		{"create_quest", applied("create_quest", `{"id":"`+questID.String()+`","campaign_id":"campaign","title":"Escape the Hollow Spire","description":"Leave at dawn.","quest_type":"short_term","status":"active","objectives":[],"related_entities":[],"summary":"Quest"}`), "quest", questID, "created", 1},
		{"update_quest", applied("update_quest", `{"id":"`+updatedQuestID.String()+`","campaign_id":"campaign","parent_quest_id":null,"title":"Escape the Hollow Spire","description":"Leave at dawn.","quest_type":"short_term","status":"completed","added_objectives":[],"objectives":[],"cascaded_subquests":[]}`), "quest", updatedQuestID, "updated", 1},
		{"complete_objective", applied("complete_objective", `{"quest_id":"`+questID.String()+`","quest_title":"Escape the Hollow Spire","quest_status":"active","objective_id":"`+objectiveID.String()+`","objective_description":"Find the key","objective_completed":true,"objectives_completed":1,"objectives_total":3,"progress":"1/3","all_objectives_complete":false,"quest_ready_for_completion":false,"quest_auto_completed":false}`), "objective", objectiveID, "completed", 2},
		{"add_item", applied("add_item", `{"item_id":"`+itemID.String()+`","player_character_id":"`+playerID.String()+`","name":"Potion","description":"Restores health","item_type":"consumable","quantity":1}`), "inventory_item", itemID, "created", 1},
		{"modify_item", applied("modify_item", `{"item_id":"`+modifiedItemID.String()+`","player_character_id":"`+playerID.String()+`","name":"Amulet","description":"A tuned focus","item_type":"misc","rarity":"rare","properties":{"effects":["clarity"]},"formatted_description":"Amulet"}`), "inventory_item", modifiedItemID, "updated", 1},
		{"remove_item", applied("remove_item", `{"item_id":"`+removedItemID.String()+`","player_character_id":"`+playerID.String()+`","name":"Torch","removed_quantity":1,"remaining_quantity":0,"deleted":true}`), "inventory_item", removedItemID, "removed", 1},
		{"create_location", applied("create_location", `{"id":"`+locationID.String()+`","campaign_id":"campaign","name":"Old Road","description":"A weathered road","region":"Borderlands","location_type":"wilderness","properties":{},"connections":[]}`), "location", locationID, "created", 1},
		{"reveal_location", applied("reveal_location", `{"location_id":"`+locationID.String()+`","campaign_id":"campaign","name":"Old Road"}`), "location", locationID, "revealed", 1},
		{"update_player_status", applied("update_player_status", `{"player_character_id":"`+playerID.String()+`","status":"cursed","statuses":[{"status":"cursed"}],"duration":{"unit":"turns","value":"2"}}`), "player_character", playerID, "status_updated", 1},
		{"add_experience", applied("add_experience", `{"player_character_id":"`+playerXPID.String()+`","amount":50,"reason":"defeating the bandit","old_experience":980,"new_experience":1030,"current_level":1,"level_up_threshold":1000,"level_up_available":true,"experience_to_next":0}`), "player_character", playerXPID, "experience_updated", 1},
		{"level_up", applied("level_up", `{"player_character_id":"`+playerLevelID.String()+`","old_level":1,"new_level":2,"hp_gain":5,"new_max_hp":25,"updated_stats":{"strength":12},"new_abilities_added":["Power Strike"]}`), "player_character", playerLevelID, "level_updated", 1},
		{"update_npc", applied("update_npc", `{"npc_id":"`+updatedNPCID.String()+`","disposition":15,"alive":true,"description":"Updated description","location_id":"`+locationID.String()+`"}`), "npc", updatedNPCID, "updated", 1},
		{"update_player_hp", applied("update_player_hp", `{"player_character_id":"`+playerID.String()+`","hp":7,"max_hp":20,"old_hp":12,"old_max_hp":20}`), "player_character", playerID, "hp_updated", 1},
		{"initiate_combat", applied("initiate_combat", `{"combat_state_id":"`+combatID.String()+`","initiative_order":[],"environment":"trail","surprise":"players","enemy_count":1,"opening_description":"Combat begins!","mode":"combat"}`), "combat", combatID, "started", 1},
		{"resolve_combat", applied("resolve_combat", `{"xp_earned":150,"loot":[],"dead_npc_ids":[],"combat_state":{"id":"`+combatStateID.String()+`","campaign_id":"campaign","round_number":2,"status":"completed","narrative":"","initiative_order":[],"environment":{"description":"trail"},"combatants":[]},"outcome_type":"victory"}`), "combat", combatStateID, "resolved", 1},
	}

	for _, tc := range cases {
		changes := StateChangesFromAppliedToolCalls([]AppliedToolCall{tc.call})
		if len(changes) != tc.count {
			t.Fatalf("%s: len(changes) = %d, want %d", tc.name, len(changes), tc.count)
		}
		if tc.count == 2 {
			if changes[0].Entity != "objective" || changes[0].EntityID != objectiveID || changes[0].Field != "completed" {
				t.Fatalf("%s: first change = %#v", tc.name, changes[0])
			}
			if changes[1].Entity != "quest" || changes[1].EntityID != questID || changes[1].Field != "updated" {
				t.Fatalf("%s: second change = %#v", tc.name, changes[1])
			}
			continue
		}
		got := changes[0]
		if got.Entity != tc.entity || got.EntityID != tc.id || got.Field != tc.field {
			t.Fatalf("%s: change = %#v, want entity=%q id=%s field=%q", tc.name, got, tc.entity, tc.id, tc.field)
		}
		if len(got.NewValue) == 0 {
			t.Fatalf("%s: missing NewValue", tc.name)
		}
		if tc.name == "resolve_combat" {
			var payload map[string]any
			if err := json.Unmarshal(got.NewValue, &payload); err != nil {
				t.Fatalf("%s: unmarshal new value: %v", tc.name, err)
			}
			if _, ok := payload["outcome_type"]; ok {
				t.Fatalf("%s: unexpected outcome_type in projected new value: %#v", tc.name, payload)
			}
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

func TestStateChangesFromAppliedToolCalls_AdvanceTimeDoesNotEmitProjectionWithoutCampaignID(t *testing.T) {
	changes := StateChangesFromAppliedToolCalls([]AppliedToolCall{
		applied("advance_time", `{"day":2,"hour":9,"minute":30}`),
	})
	if len(changes) != 0 {
		t.Fatalf("changes = %#v, want none", changes)
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
