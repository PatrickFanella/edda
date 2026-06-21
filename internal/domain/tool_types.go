package domain

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"git.subcult.tv/subculture-collective/edda/internal/memory"
)

// CreateLanguageParams holds the parameters for creating a language.
type CreateLanguageParams struct {
	CampaignID         uuid.UUID
	Name               string
	Description        string
	PhonologicalRules  json.RawMessage
	NamingConventions  json.RawMessage
	SampleVocabulary   json.RawMessage
	SpokenByFactionIDs []uuid.UUID
	SpokenByCultureIDs []uuid.UUID
}

// CreateMemoryParams holds the parameters for creating a semantic memory.
type CreateMemoryParams struct {
	CampaignID uuid.UUID
	Content    string
	Embedding  []float32
	MemoryType string
	Metadata   json.RawMessage
}

// PlayerItem represents a player's item stack.
type PlayerItem struct {
	ID                uuid.UUID
	PlayerCharacterID uuid.UUID
	Name              string
	Description       string
	ItemType          string
	Rarity            string
	Properties        map[string]any
	Equipped          bool
	Quantity          int
}

// CreateNPCParams holds the parameters for creating an NPC.
type CreateNPCParams struct {
	CampaignID  uuid.UUID
	Name        string
	Description string
	Personality string
	Disposition int
	LocationID  *uuid.UUID
	FactionID   *uuid.UUID
	Stats       json.RawMessage
	Properties  json.RawMessage
}

// NPCDialogueLogEntry captures the dialogue event to persist in the session log.
type NPCDialogueLogEntry struct {
	CampaignID        uuid.UUID
	LocationID        uuid.UUID
	NPCID             uuid.UUID
	Dialogue          string
	Emotion           *string
	FormattedDialogue string
}

// InitiateCombatNPCParams contains NPC fields required when creating enemy records.
type InitiateCombatNPCParams struct {
	CampaignID  uuid.UUID
	Name        string
	Description string
	LocationID  *uuid.UUID
	HP          int
	Stats       json.RawMessage
	Abilities   json.RawMessage
}

// InitiateCombatLogEntry contains persisted context for a newly-started encounter.
type InitiateCombatLogEntry struct {
	CampaignID             uuid.UUID
	LocationID             *uuid.UUID
	EnemyNPCIDs            []uuid.UUID
	EnvironmentDescription string
	OpeningDescription     string
}

// StatModifierResolver resolves a character's modifier for a given skill/stat.
type StatModifierResolver interface {
	GetStatModifier(ctx context.Context, characterID uuid.UUID, skill string) (int, error)
}

// AddExperienceStore provides persistence required for add_experience.
type AddExperienceStore interface {
	GetPlayerCharacterByID(ctx context.Context, playerCharacterID uuid.UUID) (*PlayerCharacter, error)
	UpdatePlayerExperience(ctx context.Context, playerCharacterID uuid.UUID, experience, level int) error
}

// LevelUpStore provides persistence required for level_up.
type LevelUpStore interface {
	GetPlayerCharacterByID(ctx context.Context, playerCharacterID uuid.UUID) (*PlayerCharacter, error)
	UpdatePlayerLevel(ctx context.Context, playerCharacterID uuid.UUID, level int) error
	UpdatePlayerStats(ctx context.Context, playerCharacterID uuid.UUID, stats json.RawMessage) error
	UpdatePlayerAbilities(ctx context.Context, playerCharacterID uuid.UUID, abilities json.RawMessage) error
	UpdatePlayerHP(ctx context.Context, playerCharacterID uuid.UUID, hp, maxHP int) error
}

// SearchMemorySearcher is the narrow interface required by the search_memory tool.
type SearchMemorySearcher interface {
	SearchSimilar(ctx context.Context, campaignID uuid.UUID, query string, limit int) ([]memory.MemoryResult, error)
}

// FeatBonusDB is an optional database interface for looking up feat bonuses
// during skill checks. When nil, feat bonuses are skipped.
type FeatBonusDB interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}
