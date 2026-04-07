// Package export provides read-only data retrieval and HTTP handlers for
// campaign export in JSON and Markdown formats.
package export

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// DBTX is the database interface satisfied by *pgxpool.Pool, pgx.Conn, and pgx.Tx.
type DBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

// Store provides raw SQL queries to gather full campaign data for export.
type Store struct {
	db DBTX
}

// NewStore creates a Store backed by the given database connection.
func NewStore(db DBTX) *Store {
	return &Store{db: db}
}

// CampaignMeta holds campaign metadata for export.
type CampaignMeta struct {
	ID          uuid.UUID
	Name        string
	Description string
	Genre       string
	Tone        string
	Themes      []string
	Status      string
	RulesMode   string
	CreatedAt   time.Time
}

const getCampaignMetaSQL = `
SELECT id, name, COALESCE(description, ''), COALESCE(genre, ''), COALESCE(tone, ''),
       COALESCE(themes, '{}'::text[]), status, COALESCE(rules_mode, 'narrative'), created_at
FROM campaigns WHERE id = $1
`

// GetCampaignMeta returns campaign metadata.
func (s *Store) GetCampaignMeta(ctx context.Context, campaignID uuid.UUID) (CampaignMeta, error) {
	pgCID := pgtype.UUID{Bytes: campaignID, Valid: campaignID != uuid.Nil}
	var m CampaignMeta
	var pgID pgtype.UUID
	var pgCreatedAt pgtype.Timestamptz
	err := s.db.QueryRow(ctx, getCampaignMetaSQL, pgCID).Scan(
		&pgID, &m.Name, &m.Description, &m.Genre, &m.Tone,
		&m.Themes, &m.Status, &m.RulesMode, &pgCreatedAt,
	)
	if err != nil {
		return CampaignMeta{}, err
	}
	m.ID = uuid.UUID(pgID.Bytes)
	if pgCreatedAt.Valid {
		m.CreatedAt = pgCreatedAt.Time
	}
	return m, nil
}

// ExportNPC holds NPC data for export.
type ExportNPC struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Personality string `json:"personality"`
	Alive       bool   `json:"alive"`
}

const listAllNPCsSQL = `
SELECT id, name, COALESCE(description, ''), COALESCE(personality, ''), alive
FROM npcs WHERE campaign_id = $1 ORDER BY created_at
`

// ListAllNPCs returns every NPC in a campaign.
func (s *Store) ListAllNPCs(ctx context.Context, campaignID uuid.UUID) ([]ExportNPC, error) {
	pgCID := pgtype.UUID{Bytes: campaignID, Valid: campaignID != uuid.Nil}
	rows, err := s.db.Query(ctx, listAllNPCsSQL, pgCID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []ExportNPC
	for rows.Next() {
		var n ExportNPC
		var pgID pgtype.UUID
		if err := rows.Scan(&pgID, &n.Name, &n.Description, &n.Personality, &n.Alive); err != nil {
			return nil, err
		}
		n.ID = uuid.UUID(pgID.Bytes).String()
		results = append(results, n)
	}
	return results, rows.Err()
}

// ExportLocation holds location data for export.
type ExportLocation struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Region       string `json:"region"`
	LocationType string `json:"location_type"`
}

const listAllLocationsSQL = `
SELECT id, name, COALESCE(description, ''), COALESCE(region, ''), COALESCE(location_type, '')
FROM locations WHERE campaign_id = $1 ORDER BY created_at
`

// ListAllLocations returns every location in a campaign.
func (s *Store) ListAllLocations(ctx context.Context, campaignID uuid.UUID) ([]ExportLocation, error) {
	pgCID := pgtype.UUID{Bytes: campaignID, Valid: campaignID != uuid.Nil}
	rows, err := s.db.Query(ctx, listAllLocationsSQL, pgCID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []ExportLocation
	for rows.Next() {
		var l ExportLocation
		var pgID pgtype.UUID
		if err := rows.Scan(&pgID, &l.Name, &l.Description, &l.Region, &l.LocationType); err != nil {
			return nil, err
		}
		l.ID = uuid.UUID(pgID.Bytes).String()
		results = append(results, l)
	}
	return results, rows.Err()
}

// ExportQuest holds quest data for export.
type ExportQuest struct {
	ID          string                `json:"id"`
	Title       string                `json:"title"`
	Description string                `json:"description"`
	QuestType   string                `json:"quest_type"`
	Status      string                `json:"status"`
	Objectives  []ExportQuestObjective `json:"objectives"`
}

// ExportQuestObjective holds a quest objective for export.
type ExportQuestObjective struct {
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

const listAllQuestsSQL = `
SELECT id, title, COALESCE(description, ''), quest_type, status
FROM quests WHERE campaign_id = $1 ORDER BY created_at
`

const listObjectivesByQuestSQL = `
SELECT description, completed FROM quest_objectives WHERE quest_id = $1 ORDER BY order_index
`

// ListAllQuests returns every quest with its objectives in a campaign.
func (s *Store) ListAllQuests(ctx context.Context, campaignID uuid.UUID) ([]ExportQuest, error) {
	pgCID := pgtype.UUID{Bytes: campaignID, Valid: campaignID != uuid.Nil}
	rows, err := s.db.Query(ctx, listAllQuestsSQL, pgCID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var quests []ExportQuest
	for rows.Next() {
		var q ExportQuest
		var pgID pgtype.UUID
		if err := rows.Scan(&pgID, &q.Title, &q.Description, &q.QuestType, &q.Status); err != nil {
			return nil, err
		}
		q.ID = uuid.UUID(pgID.Bytes).String()
		quests = append(quests, q)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Fetch objectives for each quest.
	for i := range quests {
		qID, _ := uuid.Parse(quests[i].ID)
		pgQID := pgtype.UUID{Bytes: qID, Valid: qID != uuid.Nil}
		objRows, err := s.db.Query(ctx, listObjectivesByQuestSQL, pgQID)
		if err != nil {
			return nil, err
		}
		for objRows.Next() {
			var obj ExportQuestObjective
			if err := objRows.Scan(&obj.Description, &obj.Completed); err != nil {
				objRows.Close()
				return nil, err
			}
			quests[i].Objectives = append(quests[i].Objectives, obj)
		}
		objRows.Close()
		if quests[i].Objectives == nil {
			quests[i].Objectives = []ExportQuestObjective{}
		}
	}
	return quests, nil
}

// ExportSessionLog holds a session log entry for export.
type ExportSessionLog struct {
	TurnNumber  int       `json:"turn_number"`
	PlayerInput string    `json:"player_input"`
	LLMResponse string    `json:"llm_response"`
	CreatedAt   time.Time `json:"created_at"`
}

const listAllSessionLogsSQL = `
SELECT turn_number, player_input, llm_response, created_at
FROM session_logs WHERE campaign_id = $1 ORDER BY turn_number
`

// ListAllSessionLogs returns every session log for a campaign.
func (s *Store) ListAllSessionLogs(ctx context.Context, campaignID uuid.UUID) ([]ExportSessionLog, error) {
	pgCID := pgtype.UUID{Bytes: campaignID, Valid: campaignID != uuid.Nil}
	rows, err := s.db.Query(ctx, listAllSessionLogsSQL, pgCID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []ExportSessionLog
	for rows.Next() {
		var sl ExportSessionLog
		var pgCreatedAt pgtype.Timestamptz
		if err := rows.Scan(&sl.TurnNumber, &sl.PlayerInput, &sl.LLMResponse, &pgCreatedAt); err != nil {
			return nil, err
		}
		if pgCreatedAt.Valid {
			sl.CreatedAt = pgCreatedAt.Time
		}
		results = append(results, sl)
	}
	return results, rows.Err()
}

// ExportWorldFact holds a world fact for export.
type ExportWorldFact struct {
	Fact     string `json:"fact"`
	Category string `json:"category"`
	Source   string `json:"source"`
}

const listAllWorldFactsSQL = `
SELECT fact, category, source FROM world_facts
WHERE campaign_id = $1 AND superseded_by IS NULL ORDER BY created_at
`

// ListAllWorldFacts returns all active world facts for a campaign.
func (s *Store) ListAllWorldFacts(ctx context.Context, campaignID uuid.UUID) ([]ExportWorldFact, error) {
	pgCID := pgtype.UUID{Bytes: campaignID, Valid: campaignID != uuid.Nil}
	rows, err := s.db.Query(ctx, listAllWorldFactsSQL, pgCID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []ExportWorldFact
	for rows.Next() {
		var f ExportWorldFact
		if err := rows.Scan(&f.Fact, &f.Category, &f.Source); err != nil {
			return nil, err
		}
		results = append(results, f)
	}
	return results, rows.Err()
}

// ExportCharacter holds player character data for export.
type ExportCharacter struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	HP          int    `json:"hp"`
	MaxHP       int    `json:"max_hp"`
	Experience  int    `json:"experience"`
	Level       int    `json:"level"`
	Status      string `json:"status"`
	Abilities   []byte `json:"abilities_raw"`
}

const getPlayerCharacterSQL = `
SELECT id, name, COALESCE(description, ''), hp, max_hp, experience, level, status, COALESCE(abilities, '[]'::jsonb)
FROM player_characters WHERE campaign_id = $1 ORDER BY created_at DESC LIMIT 1
`

// GetPlayerCharacter returns the most recent player character for a campaign.
func (s *Store) GetPlayerCharacter(ctx context.Context, campaignID uuid.UUID) (ExportCharacter, error) {
	pgCID := pgtype.UUID{Bytes: campaignID, Valid: campaignID != uuid.Nil}
	var c ExportCharacter
	var pgID pgtype.UUID
	err := s.db.QueryRow(ctx, getPlayerCharacterSQL, pgCID).Scan(
		&pgID, &c.Name, &c.Description, &c.HP, &c.MaxHP, &c.Experience, &c.Level, &c.Status, &c.Abilities,
	)
	if err != nil {
		return ExportCharacter{}, err
	}
	c.ID = uuid.UUID(pgID.Bytes).String()
	return c, nil
}

// ExportItem holds an inventory item for export.
type ExportItem struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Equipped bool   `json:"equipped"`
}

const listPlayerItemsSQL = `
SELECT name, quantity, equipped
FROM items WHERE campaign_id = $1 AND player_character_id IS NOT NULL ORDER BY created_at
`

// ListPlayerItems returns all items belonging to the player character.
func (s *Store) ListPlayerItems(ctx context.Context, campaignID uuid.UUID) ([]ExportItem, error) {
	pgCID := pgtype.UUID{Bytes: campaignID, Valid: campaignID != uuid.Nil}
	rows, err := s.db.Query(ctx, listPlayerItemsSQL, pgCID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []ExportItem
	for rows.Next() {
		var item ExportItem
		if err := rows.Scan(&item.Name, &item.Quantity, &item.Equipped); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, rows.Err()
}
