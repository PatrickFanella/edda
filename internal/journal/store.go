// Package journal provides persistence and HTTP handlers for session summaries
// and player journal entries.
package journal

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

// Summary represents a row in the session_summaries table.
type Summary struct {
	ID         uuid.UUID
	CampaignID uuid.UUID
	FromTurn   int
	ToTurn     int
	Summary    string
	CreatedAt  time.Time
}

// Entry represents a row in the player_journal_entries table.
type Entry struct {
	ID         uuid.UUID
	CampaignID uuid.UUID
	Title      string
	Content    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Store provides direct session_summaries and player_journal_entries operations using raw SQL.
type Store struct {
	db DBTX
}

// NewStore creates a Store backed by the given database connection.
func NewStore(db DBTX) *Store {
	return &Store{db: db}
}

const listSummariesSQL = `
SELECT id, campaign_id, from_turn, to_turn, summary, created_at
FROM session_summaries
WHERE campaign_id = $1
ORDER BY created_at DESC
`

// ListSummaries returns all session summaries for a campaign, newest first.
func (s *Store) ListSummaries(ctx context.Context, campaignID uuid.UUID) ([]Summary, error) {
	pgCID := pgtype.UUID{Bytes: campaignID, Valid: campaignID != uuid.Nil}
	rows, err := s.db.Query(ctx, listSummariesSQL, pgCID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Summary
	for rows.Next() {
		var sm Summary
		var pgID, pgCampaignID pgtype.UUID
		var pgCreatedAt pgtype.Timestamptz
		if err := rows.Scan(&pgID, &pgCampaignID, &sm.FromTurn, &sm.ToTurn, &sm.Summary, &pgCreatedAt); err != nil {
			return nil, err
		}
		sm.ID = uuid.UUID(pgID.Bytes)
		sm.CampaignID = uuid.UUID(pgCampaignID.Bytes)
		if pgCreatedAt.Valid {
			sm.CreatedAt = pgCreatedAt.Time
		}
		results = append(results, sm)
	}
	return results, rows.Err()
}

const listEntriesSQL = `
SELECT id, campaign_id, title, content, created_at, updated_at
FROM player_journal_entries
WHERE campaign_id = $1
ORDER BY created_at DESC
`

// ListEntries returns all journal entries for a campaign, newest first.
func (s *Store) ListEntries(ctx context.Context, campaignID uuid.UUID) ([]Entry, error) {
	pgCID := pgtype.UUID{Bytes: campaignID, Valid: campaignID != uuid.Nil}
	rows, err := s.db.Query(ctx, listEntriesSQL, pgCID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Entry
	for rows.Next() {
		var e Entry
		var pgID, pgCampaignID pgtype.UUID
		var pgCreatedAt, pgUpdatedAt pgtype.Timestamptz
		if err := rows.Scan(&pgID, &pgCampaignID, &e.Title, &e.Content, &pgCreatedAt, &pgUpdatedAt); err != nil {
			return nil, err
		}
		e.ID = uuid.UUID(pgID.Bytes)
		e.CampaignID = uuid.UUID(pgCampaignID.Bytes)
		if pgCreatedAt.Valid {
			e.CreatedAt = pgCreatedAt.Time
		}
		if pgUpdatedAt.Valid {
			e.UpdatedAt = pgUpdatedAt.Time
		}
		results = append(results, e)
	}
	return results, rows.Err()
}

const createEntrySQL = `
INSERT INTO player_journal_entries (campaign_id, title, content)
VALUES ($1, $2, $3)
RETURNING id, campaign_id, title, content, created_at, updated_at
`

// CreateEntry inserts a new journal entry and returns it.
func (s *Store) CreateEntry(ctx context.Context, campaignID uuid.UUID, title, content string) (Entry, error) {
	pgCID := pgtype.UUID{Bytes: campaignID, Valid: campaignID != uuid.Nil}
	row := s.db.QueryRow(ctx, createEntrySQL, pgCID, title, content)
	var e Entry
	var pgID, pgCampaignID pgtype.UUID
	var pgCreatedAt, pgUpdatedAt pgtype.Timestamptz
	err := row.Scan(&pgID, &pgCampaignID, &e.Title, &e.Content, &pgCreatedAt, &pgUpdatedAt)
	if err != nil {
		return Entry{}, err
	}
	e.ID = uuid.UUID(pgID.Bytes)
	e.CampaignID = uuid.UUID(pgCampaignID.Bytes)
	if pgCreatedAt.Valid {
		e.CreatedAt = pgCreatedAt.Time
	}
	if pgUpdatedAt.Valid {
		e.UpdatedAt = pgUpdatedAt.Time
	}
	return e, nil
}

const deleteEntrySQL = `DELETE FROM player_journal_entries WHERE id = $1`

// DeleteEntry removes a journal entry by ID.
func (s *Store) DeleteEntry(ctx context.Context, entryID uuid.UUID) error {
	pgID := pgtype.UUID{Bytes: entryID, Valid: entryID != uuid.Nil}
	_, err := s.db.Exec(ctx, deleteEntrySQL, pgID)
	return err
}
