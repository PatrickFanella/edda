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
	case "add_item", "create_item":
		return one(entityStateChange("inventory_item", data, "item_id", "created"))
	case "modify_item":
		return one(entityStateChange("inventory_item", data, "item_id", "updated"))
	case "remove_item":
		return one(entityStateChange("inventory_item", data, "item_id", "removed"))
	case "create_npc":
		return one(entityCreatedStateChange("npc", data))
	case "establish_fact":
		return one(entityCreatedStateChange("world_fact", data))
	case "create_quest":
		return one(entityCreatedStateChange("quest", data))
	case "update_quest":
		return one(entityStateChange("quest", data, "id", "updated"))
	case "complete_objective":
		changes := make([]StateChange, 0, 2)
		if change, ok := entityStateChange("objective", data, "objective_id", "completed"); ok {
			changes = append(changes, change)
		}
		if change, ok := entityStateChange("quest", data, "quest_id", "updated"); ok {
			changes = append(changes, change)
		}
		return changes
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
	case "add_experience":
		return one(entityStateChange("player_character", data, "player_character_id", "experience_updated"))
	case "level_up":
		return one(entityStateChange("player_character", data, "player_character_id", "level_updated"))
	case "update_npc":
		return one(entityStateChange("npc", data, "npc_id", "updated"))
	case "initiate_combat":
		return one(entityStateChange("combat", data, "combat_state_id", "started"))
	case "resolve_combat":
		if change, ok := nestedEntityStateChange("combat", data, "combat_state", "id", "resolved"); ok {
			return []StateChange{change}
		}
		return nil
	case "advance_time":
		// No campaign_id is present in the current tool result shape, so there is
		// no stable projection target for a campaign_time state change.
		return nil
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

func nestedEntityStateChange(entity string, data map[string]any, nestedKey, idKey, changeName string) (StateChange, bool) {
	raw, ok := data[nestedKey]
	if !ok {
		return StateChange{}, false
	}
	nested, ok := raw.(map[string]any)
	if !ok {
		return StateChange{}, false
	}
	return entityStateChange(entity, nested, idKey, changeName)
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
