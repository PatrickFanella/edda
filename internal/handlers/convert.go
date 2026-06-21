package handlers

import (
	"encoding/json"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
	"github.com/jackc/pgx/v5/pgtype"
)

func campaignToResponse(c statedb.Campaign) api.CampaignResponse {
	themes := c.Themes
	if themes == nil {
		themes = []string{}
	}
	rulesMode := c.RulesMode
	if rulesMode == "" {
		rulesMode = "narrative"
	}
	return api.CampaignResponse{
		ID:          dbutil.FromPgtype(c.ID).String(),
		Name:        c.Name,
		Description: c.Description.String,
		Genre:       c.Genre.String,
		Tone:        c.Tone.String,
		Themes:      themes,
		Status:      c.Status,
		RulesMode:   rulesMode,
		CreatedBy:   dbutil.FromPgtype(c.CreatedBy).String(),
		CreatedAt:   timestamptzValue(c.CreatedAt),
		UpdatedAt:   timestamptzValue(c.UpdatedAt),
	}
}

func playerCharacterToResponse(pc statedb.PlayerCharacter) api.CharacterResponse {
	stats := unmarshalJSONMap(pc.Stats)
	abilities := parseAbilities(pc.Abilities)

	return api.CharacterResponse{
		ID:                dbutil.FromPgtype(pc.ID).String(),
		CampaignID:        dbutil.FromPgtype(pc.CampaignID).String(),
		UserID:            dbutil.FromPgtype(pc.UserID).String(),
		Name:              pc.Name,
		Description:       pc.Description.String,
		Stats:             stats,
		HP:                int(pc.Hp),
		MaxHP:             int(pc.MaxHp),
		Experience:        int(pc.Experience),
		Level:             int(pc.Level),
		Status:            pc.Status,
		Abilities:         abilities,
		CurrentLocationID: optionalUUIDString(pc.CurrentLocationID),
	}
}

func locationToResponse(l statedb.Location, conns []statedb.GetConnectionsFromLocationRow) api.LocationResponse {
	connections := make([]api.LocationConnectionResponse, 0, len(conns))
	for _, c := range conns {
		connections = append(connections, api.LocationConnectionResponse{
			ToLocationID:  dbutil.FromPgtype(c.ConnectedLocationID).String(),
			Description:   c.Description.String,
			Bidirectional: c.Bidirectional,
			TravelTime:    c.TravelTime.String,
		})
	}
	return api.LocationResponse{
		ID:           dbutil.FromPgtype(l.ID).String(),
		CampaignID:   dbutil.FromPgtype(l.CampaignID).String(),
		Name:         l.Name,
		Description:  l.Description.String,
		Region:       l.Region.String,
		LocationType: l.LocationType.String,
		Properties:   unmarshalJSONMap(l.Properties),
		Connections:  connections,
	}
}

func npcToResponse(n statedb.Npc) api.NPCResponse {
	return api.NPCResponse{
		ID:          dbutil.FromPgtype(n.ID).String(),
		CampaignID:  dbutil.FromPgtype(n.CampaignID).String(),
		Name:        n.Name,
		Description: n.Description.String,
		Personality: n.Personality.String,
		Disposition: int(n.Disposition),
		FactionID:   optionalUUIDString(n.FactionID),
		Alive:       n.Alive,
		HP:          optionalInt32Value(n.Hp),
		Stats:       unmarshalJSONMap(n.Stats),
		Properties:  unmarshalJSONMap(n.Properties),
	}
}

func questToResponse(q statedb.Quest, objs []statedb.QuestObjective) api.QuestResponse {
	objectives := make([]api.QuestObjectiveResponse, 0, len(objs))
	for _, o := range objs {
		objectives = append(objectives, api.QuestObjectiveResponse{
			ID:          dbutil.FromPgtype(o.ID).String(),
			Description: o.Description,
			Completed:   o.Completed,
			OrderIndex:  int(o.OrderIndex),
		})
	}
	return api.QuestResponse{
		ID:            dbutil.FromPgtype(q.ID).String(),
		CampaignID:    dbutil.FromPgtype(q.CampaignID).String(),
		ParentQuestID: optionalUUIDString(q.ParentQuestID),
		Title:         q.Title,
		Description:   q.Description.String,
		QuestType:     q.QuestType,
		Status:        q.Status,
		Objectives:    objectives,
	}
}

func itemToResponse(i statedb.Item) api.ItemResponse {
	return api.ItemResponse{
		ID:                dbutil.FromPgtype(i.ID).String(),
		CampaignID:        dbutil.FromPgtype(i.CampaignID).String(),
		PlayerCharacterID: optionalUUIDString(i.PlayerCharacterID),
		Name:              i.Name,
		Description:       i.Description.String,
		ItemType:          i.ItemType,
		Rarity:            i.Rarity,
		Properties:        unmarshalJSONMap(i.Properties),
		Equipped:          i.Equipped,
		Quantity:          int(i.Quantity),
	}
}

func engineStateChangesToAPI(changes []engine.StateChange) []api.StateChange {
	out := make([]api.StateChange, 0, len(changes))
	for _, sc := range changes {
		details := make(map[string]any, 2)
		if sc.OldValue != nil {
			var old any
			if json.Unmarshal(sc.OldValue, &old) == nil {
				details["old_value"] = old
			}
		}
		if sc.NewValue != nil {
			var nv any
			if json.Unmarshal(sc.NewValue, &nv) == nil {
				if fields, ok := nv.(map[string]any); ok {
					for key, value := range fields {
						details[key] = value
					}
				}
				details["new_value"] = nv
			}
		}
		out = append(out, api.StateChange{
			EntityType: sc.Entity,
			EntityID:   sc.EntityID.String(),
			ChangeType: sc.Field,
			Details:    details,
		})
	}
	return out
}

func engineTurnResultToAPI(tr *engine.TurnResult) api.TurnResponse {
	return api.TurnResponse{
		Narrative:    tr.Narrative,
		StateChanges: engineStateChangesToAPI(tr.StateChanges),
		CombatActive: tr.CombatActive,
	}
}

func sessionLogToEntry(sl statedb.SessionLog) api.SessionLogEntry {
	choices := openingChoicesFromToolCalls(sl.ToolCalls)
	return api.SessionLogEntry{
		TurnNumber:  int(sl.TurnNumber),
		PlayerInput: sl.PlayerInput,
		InputType:   sl.InputType,
		LLMResponse: sl.LlmResponse,
		Choices:     choices,
		CreatedAt:   sl.CreatedAt.Time,
	}
}

func openingChoicesFromToolCalls(toolCalls []byte) []string {
	if len(toolCalls) == 0 {
		return nil
	}

	var payload []struct {
		Type    string   `json:"type"`
		Choices []string `json:"choices"`
	}
	if err := json.Unmarshal(toolCalls, &payload); err != nil {
		return nil
	}
	for _, call := range payload {
		if call.Type == "opening_choices" {
			if len(call.Choices) == 0 {
				return nil
			}
			return call.Choices
		}
	}
	return nil
}

// unmarshalJSONMap decodes raw JSON bytes into a map; returns empty map on
// nil/empty input or decode failure.
func unmarshalJSONMap(data []byte) map[string]any {
	if len(data) == 0 {
		return map[string]any{}
	}
	var m map[string]any
	if json.Unmarshal(data, &m) != nil {
		return map[string]any{}
	}
	return m
}

// parseAbilities decodes raw JSON bytes into a slice of CharacterAbility.
func parseAbilities(data []byte) []api.CharacterAbility {
	if len(data) == 0 {
		return []api.CharacterAbility{}
	}
	var abs []api.CharacterAbility
	if json.Unmarshal(data, &abs) != nil {
		return []api.CharacterAbility{}
	}
	return abs
}

func factToResponse(f statedb.WorldFact) api.FactResponse {
	return api.FactResponse{
		ID:           dbutil.FromPgtype(f.ID).String(),
		CampaignID:   dbutil.FromPgtype(f.CampaignID).String(),
		Fact:         f.Fact,
		Category:     f.Category,
		Source:       f.Source,
		SupersededBy: optionalUUIDString(f.SupersededBy),
		PlayerKnown:  f.PlayerKnown,
		CreatedAt:    timestamptzRFC3339(f.CreatedAt),
	}
}

func relationshipToResponse(r statedb.EntityRelationship) api.RelationshipResponse {
	return api.RelationshipResponse{
		ID:               dbutil.FromPgtype(r.ID).String(),
		CampaignID:       dbutil.FromPgtype(r.CampaignID).String(),
		SourceEntityType: r.SourceEntityType,
		SourceEntityID:   dbutil.FromPgtype(r.SourceEntityID).String(),
		TargetEntityType: r.TargetEntityType,
		TargetEntityID:   dbutil.FromPgtype(r.TargetEntityID).String(),
		RelationshipType: r.RelationshipType,
		Description:      r.Description.String,
		Strength:         optionalInt32Value(r.Strength),
		PlayerAware:      r.PlayerAware,
		CreatedAt:        timestamptzRFC3339(r.CreatedAt),
	}
}

func languageToResponse(l statedb.Language) api.LanguageResponse {
	return api.LanguageResponse{
		ID:          dbutil.FromPgtype(l.ID).String(),
		CampaignID:  dbutil.FromPgtype(l.CampaignID).String(),
		Name:        l.Name,
		Description: l.Description,
		PlayerKnown: l.PlayerKnown,
		CreatedAt:   timestamptzRFC3339(l.CreatedAt),
	}
}

func cultureToResponse(c statedb.Culture) api.CultureResponse {
	return api.CultureResponse{
		ID:             dbutil.FromPgtype(c.ID).String(),
		CampaignID:     dbutil.FromPgtype(c.CampaignID).String(),
		Name:           c.Name,
		LanguageID:     optionalUUIDString(c.LanguageID),
		BeliefSystemID: optionalUUIDString(c.BeliefSystemID),
		PlayerKnown:    c.PlayerKnown,
		CreatedAt:      timestamptzRFC3339(c.CreatedAt),
	}
}

func beliefSystemToResponse(b statedb.BeliefSystem) api.BeliefSystemResponse {
	return api.BeliefSystemResponse{
		ID:          dbutil.FromPgtype(b.ID).String(),
		CampaignID:  dbutil.FromPgtype(b.CampaignID).String(),
		Name:        b.Name,
		PlayerKnown: b.PlayerKnown,
		CreatedAt:   timestamptzRFC3339(b.CreatedAt),
	}
}

func economicSystemToResponse(e statedb.EconomicSystem) api.EconomicSystemResponse {
	return api.EconomicSystemResponse{
		ID:          dbutil.FromPgtype(e.ID).String(),
		CampaignID:  dbutil.FromPgtype(e.CampaignID).String(),
		Name:        e.Name,
		PlayerKnown: e.PlayerKnown,
		CreatedAt:   timestamptzRFC3339(e.CreatedAt),
	}
}

func mapLocationToResponse(l statedb.Location) api.MapLocationResponse {
	return api.MapLocationResponse{
		ID:            dbutil.FromPgtype(l.ID).String(),
		CampaignID:    dbutil.FromPgtype(l.CampaignID).String(),
		Name:          l.Name,
		Description:   l.Description.String,
		Region:        l.Region.String,
		LocationType:  l.LocationType.String,
		PlayerVisited: l.PlayerVisited,
		PlayerKnown:   l.PlayerKnown,
	}
}

func optionalUUIDString(id pgtype.UUID) *string {
	if !id.Valid {
		return nil
	}
	s := dbutil.FromPgtype(id).String()
	return &s
}

func optionalInt32Value(v pgtype.Int4) *int {
	if !v.Valid {
		return nil
	}
	i := int(v.Int32)
	return &i
}

func timestamptzValue(ts pgtype.Timestamptz) time.Time {
	return ts.Time
}

func timestamptzRFC3339(ts pgtype.Timestamptz) string {
	return ts.Time.Format(time.RFC3339)
}
