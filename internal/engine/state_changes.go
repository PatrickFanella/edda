package engine

import (
	"encoding/json"

	"github.com/google/uuid"
)

// StateChangesFromAppliedToolCalls projects durable tool results into API-visible state changes.
func StateChangesFromAppliedToolCalls(applied []AppliedToolCall) []StateChange {
	changes := make([]StateChange, 0, len(applied))
	for _, call := range applied {
		changes = append(changes, stateChangesFromAppliedToolCall(call)...)
	}
	return changes
}

func stateChangesFromAppliedToolCall(call AppliedToolCall) []StateChange {
	var data map[string]any
	if err := json.Unmarshal(call.Result, &data); err != nil {
		return nil
	}

	switch call.Tool {
	case "create_npc":
		return one(entityCreatedStateChange("npc", data))
	case "establish_fact":
		return one(entityCreatedStateChange("world_fact", data))
	case "create_quest":
		return one(entityCreatedStateChange("quest", data))
	case "create_location":
		changes := make([]StateChange, 0, 4)
		if reused, _ := data["reused"].(bool); !reused {
			if change, ok := entityCreatedStateChange("location", data); ok {
				changes = append(changes, change)
			}
		}
		if moved, _ := data["move_player_here"].(bool); moved {
			if change, ok := entityStateChange("player_character", data, "player_character_id", "location_updated"); ok {
				changes = append(changes, change)
			}
			if change, ok := entityStateChange("location", data, "location_id", "updated"); ok {
				changes = append(changes, change)
			}
			if change, ok := entityStateChange("location", data, "location_id", "moved"); ok {
				changes = append(changes, change)
			}
		}
		return changes
	case "reveal_location":
		return one(entityStateChange("location", data, "location_id", "revealed"))
	case "move_player":
		changes := make([]StateChange, 0, 3)
		if change, ok := entityStateChange("player_character", data, "player_character_id", "location_updated"); ok {
			changes = append(changes, change)
		}
		if change, ok := entityStateChange("location", data, "location_id", "updated"); ok {
			changes = append(changes, change)
		}
		if change, ok := entityStateChange("location", data, "location_id", "moved"); ok {
			changes = append(changes, change)
		}
		return changes
	case "update_player_status":
		return one(entityStateChange("player_character", data, "player_character_id", "status_updated"))
	case "update_player_hp":
		return one(entityStateChange("player_character", data, "player_character_id", "hp_updated"))
	case "initiate_combat":
		return one(entityStateChange("combat", data, "combat_state_id", "started"))
	default:
		return nil
	}
}

func one(change StateChange, ok bool) []StateChange {
	if !ok {
		return nil
	}
	return []StateChange{change}
}

func entityCreatedStateChange(entity string, data map[string]any) (StateChange, bool) {
	return entityStateChange(entity, data, "id", "created")
}

func entityStateChange(entity string, data map[string]any, idKey, changeName string) (StateChange, bool) {
	id, ok := uuidField(data, "id")
	if idKey != "id" {
		id, ok = uuidField(data, idKey)
	}
	if !ok {
		return StateChange{}, false
	}
	return StateChange{Entity: entity, EntityID: id, Field: changeName, NewValue: mustJSON(data)}, true
}

func uuidField(data map[string]any, key string) (uuid.UUID, bool) {
	raw, ok := data[key]
	if !ok {
		return uuid.UUID{}, false
	}
	str, ok := raw.(string)
	if !ok {
		return uuid.UUID{}, false
	}
	id, err := uuid.Parse(str)
	return id, err == nil
}

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}
