package handlers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
)

func TestConvertResponsesGoldenShapes(t *testing.T) {
	t.Parallel()

	campID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	locID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	playerID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
	combatID := uuid.MustParse("88888888-8888-8888-8888-888888888888")
	factionID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	questID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	factSupersededBy := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	moveLocationID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	moveApplied := engine.AppliedToolCall{Tool: "move_player", Result: mustJSON(t, map[string]any{"location_id": moveLocationID.String(), "campaign_id": campID.String(), "player_character_id": playerID.String(), "name": "Old Road", "description": "A weathered road", "location_type": "wilderness", "travel_time": "1 hour", "day": 1, "hour": 9, "minute": 0, "visited_marked": true, "visited_warning": "", "time_warning": ""})}
	combatApplied := engine.AppliedToolCall{Tool: "resolve_combat", Result: mustJSON(t, map[string]any{"xp_earned": 150, "loot": []any{}, "dead_npc_ids": []any{}, "combat_state": map[string]any{"id": combatID.String(), "campaign_id": campID.String(), "round_number": 2, "status": "completed", "narrative": "", "initiative_order": []any{}, "environment": map[string]any{"description": "trail"}, "combatants": []any{}}, "outcome_type": "victory"})}
	now := time.Date(2026, 6, 19, 12, 34, 56, 0, time.UTC)
	conv := map[string]any{
		"campaign":  campaignToResponse(statedb.Campaign{ID: dbutil.ToPgtype(campID), Name: "The Iron Road", Description: pgtype.Text{String: "A frontier campaign", Valid: true}, Genre: pgtype.Text{String: "fantasy", Valid: true}, Tone: pgtype.Text{String: "gritty", Valid: true}, Themes: []string{"exploration", "survival"}, Status: "active", CreatedBy: dbutil.ToPgtype(userID), CreatedAt: pgtype.Timestamptz{Time: now, Valid: true}, UpdatedAt: pgtype.Timestamptz{Time: now.Add(time.Hour), Valid: true}, RulesMode: "narrative"}),
		"character": playerCharacterToResponse(statedb.PlayerCharacter{ID: dbutil.ToPgtype(playerID), CampaignID: dbutil.ToPgtype(campID), UserID: dbutil.ToPgtype(userID), Name: "Vera", Description: pgtype.Text{String: "Scout", Valid: true}, Stats: mustJSON(t, map[string]any{"strength": 12, "agility": 14}), Hp: 20, MaxHp: 25, Experience: 120, Level: 3, Status: "healthy", Abilities: mustJSON(t, []api.CharacterAbility{{Name: "Track", Description: "Follow signs"}}), CurrentLocationID: dbutil.ToPgtype(locID)}),
		"quest":     questToResponse(statedb.Quest{ID: dbutil.ToPgtype(questID), CampaignID: dbutil.ToPgtype(campID), ParentQuestID: dbutil.ToPgtype(uuid.MustParse("88888888-8888-8888-8888-888888888888")), Title: "Secure the pass", Description: pgtype.Text{String: "Defeat raiders", Valid: true}, QuestType: "short_term", Status: "active"}, []statedb.QuestObjective{{ID: dbutil.ToPgtype(uuid.MustParse("99999999-9999-9999-9999-999999999999")), Description: "Scout the pass", Completed: false, OrderIndex: 1}}),
		"fact":      factToResponse(statedb.WorldFact{ID: dbutil.ToPgtype(uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")), CampaignID: dbutil.ToPgtype(campID), Fact: "The road is cursed.", Category: "lore", Source: "gm", SupersededBy: dbutil.ToPgtype(factSupersededBy), PlayerKnown: true, CreatedAt: pgtype.Timestamptz{Time: now, Valid: true}}),
		"location":  locationToResponse(statedb.Location{ID: dbutil.ToPgtype(locID), CampaignID: dbutil.ToPgtype(campID), Name: "Gatehouse", Description: pgtype.Text{String: "Stone fort", Valid: true}, Region: pgtype.Text{String: "North", Valid: true}, LocationType: pgtype.Text{String: "city", Valid: true}, Properties: mustJSON(t, map[string]any{"danger": "low"})}, []statedb.GetConnectionsFromLocationRow{{ConnectedLocationID: dbutil.ToPgtype(uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")), Description: pgtype.Text{String: "Road east", Valid: true}, Bidirectional: true, TravelTime: pgtype.Text{String: "15m", Valid: true}}}),
		"npc":       npcToResponse(statedb.Npc{ID: dbutil.ToPgtype(uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")), CampaignID: dbutil.ToPgtype(campID), Name: "Captain Rowan", Description: pgtype.Text{String: "Guard commander", Valid: true}, Personality: pgtype.Text{String: "stern", Valid: true}, Disposition: 40, FactionID: dbutil.ToPgtype(factionID), Alive: true, Hp: pgtype.Int4{Int32: 22, Valid: true}, Stats: mustJSON(t, map[string]any{"intelligence": 14}), Properties: mustJSON(t, map[string]any{"rank": "captain"})}),
		"turn":      engineTurnResultToAPI(&engine.TurnResult{Narrative: "You arrive at the gatehouse.", StateChanges: append(engine.StateChangesFromAppliedToolCalls([]engine.AppliedToolCall{moveApplied}), engine.StateChangesFromAppliedToolCalls([]engine.AppliedToolCall{combatApplied})...), CombatActive: false}),
	}

	want, err := os.ReadFile(filepath.Join("testdata", "convert_response_shapes.golden.json"))
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	got, err := json.MarshalIndent(conv, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got = append(got, '\n')
	if string(got) != string(want) {
		t.Fatalf("golden mismatch\nwant:\n%s\ngot:\n%s", string(want), string(got))
	}
}

func TestConversionHelpersDefaulting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  any
		want any
	}{
		{name: "empty json map", got: unmarshalJSONMap(nil), want: map[string]any{}},
		{name: "invalid json map", got: unmarshalJSONMap([]byte(`{bad`)), want: map[string]any{}},
		{name: "empty json slice", got: parseAbilities(nil), want: []api.CharacterAbility{}},
		{name: "invalid json slice", got: parseAbilities([]byte(`not json`)), want: []api.CharacterAbility{}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got, want := mustJSON(t, tt.got), mustJSON(t, tt.want); string(got) != string(want) {
				t.Fatalf("mismatch\nwant: %s\ngot:  %s", string(want), string(got))
			}
		})
	}

	if got := optionalUUIDString(pgtype.UUID{}); got != nil {
		t.Fatalf("optionalUUIDString() = %v, want nil", got)
	}
	if got := optionalInt32Value(pgtype.Int4{}); got != nil {
		t.Fatalf("optionalInt32Value() = %v, want nil", got)
	}
}

func TestEngineStateChangesToAPIMergesObjectNewValueIntoDetails(t *testing.T) {
	t.Parallel()

	locID := uuid.New()
	changes := engineStateChangesToAPI([]engine.StateChange{{
		Entity:   "location",
		EntityID: locID,
		Field:    "updated",
		NewValue: json.RawMessage(`{"location_id":"loc-1","name":"Old Road","location_type":"wilderness"}`),
	}})

	if len(changes) != 1 {
		t.Fatalf("len(changes) = %d, want 1", len(changes))
	}
	if got := changes[0].Details["location_type"]; got != "wilderness" {
		t.Fatalf("details.location_type = %v, want wilderness", got)
	}
	if got := changes[0].Details["name"]; got != "Old Road" {
		t.Fatalf("details.name = %v, want Old Road", got)
	}
	if _, ok := changes[0].Details["new_value"]; !ok {
		t.Fatalf("details.new_value missing: %#v", changes[0].Details)
	}
}

func TestEngineTurnResultToAPIPreservesStateChanges(t *testing.T) {
	t.Parallel()

	questID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	playerID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	locationID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	factID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	itemID := uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
	objectiveID := uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")
	combatID := uuid.MustParse("11111111-2222-3333-4444-555555555555")

	resp := engineTurnResultToAPI(&engine.TurnResult{
		Narrative: "The world shifts.",
		StateChanges: []engine.StateChange{
			{Entity: "quest", EntityID: questID, Field: "created", NewValue: json.RawMessage(`{"id":"` + questID.String() + `","title":"Secure the pass","status":"active"}`)},
			{Entity: "player_character", EntityID: playerID, Field: "location_updated", NewValue: json.RawMessage(`{"location_id":"` + locationID.String() + `","name":"Gatehouse"}`)},
			{Entity: "world_fact", EntityID: factID, Field: "created", NewValue: json.RawMessage(`{"id":"` + factID.String() + `","fact":"The road is cursed."}`)},
			{Entity: "player_character", EntityID: playerID, Field: "hp_updated", NewValue: json.RawMessage(`{"hp":7,"max_hp":20}`)},
			{Entity: "inventory_item", EntityID: itemID, Field: "created", NewValue: json.RawMessage(`{"item_id":"` + itemID.String() + `","name":"Potion"}`)},
			{Entity: "objective", EntityID: objectiveID, Field: "completed", NewValue: json.RawMessage(`{"objective_id":"` + objectiveID.String() + `","quest_id":"` + questID.String() + `","objective_completed":true}`)},
			{Entity: "quest", EntityID: questID, Field: "updated", NewValue: json.RawMessage(`{"id":"` + questID.String() + `","status":"completed"}`)},
			{Entity: "combat", EntityID: combatID, Field: "started", NewValue: json.RawMessage(`{"combat_state_id":"` + combatID.String() + `","opening_description":"Combat begins!"}`)},
			{Entity: "combat", EntityID: combatID, Field: "resolved", NewValue: json.RawMessage(`{"id":"` + combatID.String() + `","status":"completed","outcome_type":"victory"}`)},
		},
	})

	if resp.Narrative != "The world shifts." {
		t.Fatalf("narrative = %q, want %q", resp.Narrative, "The world shifts.")
	}
	if len(resp.StateChanges) != 9 {
		t.Fatalf("len(state_changes) = %d, want 9", len(resp.StateChanges))
	}

	assertChange := func(i int, entityType, changeType, entityID string, wantKeys ...string) {
		t.Helper()
		got := resp.StateChanges[i]
		if got.EntityType != entityType || got.ChangeType != changeType || got.EntityID != entityID {
			t.Fatalf("state_changes[%d] = %#v", i, got)
		}
		for _, key := range wantKeys {
			if _, ok := got.Details[key]; !ok {
				t.Fatalf("state_changes[%d].details missing %q: %#v", i, key, got.Details)
			}
		}
	}

	assertChange(0, "quest", "created", questID.String(), "id", "title", "status", "new_value")
	assertChange(1, "player_character", "location_updated", playerID.String(), "location_id", "name", "new_value")
	assertChange(2, "world_fact", "created", factID.String(), "id", "fact", "new_value")
	assertChange(3, "player_character", "hp_updated", playerID.String(), "hp", "max_hp", "new_value")
	assertChange(4, "inventory_item", "created", itemID.String(), "item_id", "name", "new_value")
	assertChange(5, "objective", "completed", objectiveID.String(), "objective_id", "quest_id", "objective_completed", "new_value")
	assertChange(6, "quest", "updated", questID.String(), "id", "status", "new_value")
	assertChange(7, "combat", "started", combatID.String(), "combat_state_id", "opening_description", "new_value")
	assertChange(8, "combat", "resolved", combatID.String(), "id", "status", "outcome_type", "new_value")
}

func TestOpeningChoicesFromToolCalls(t *testing.T) {
	t.Parallel()

	toolCalls := mustJSON(t, []map[string]any{{"type": "opening_choices", "choices": []string{"Go east", "Rest"}}})
	if got, want := openingChoicesFromToolCalls(toolCalls), []string{"Go east", "Rest"}; !equalStrings(got, want) {
		t.Fatalf("openingChoicesFromToolCalls() = %v, want %v", got, want)
	}

	if got := openingChoicesFromToolCalls(nil); got != nil {
		t.Fatalf("openingChoicesFromToolCalls(nil) = %v, want nil", got)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}
