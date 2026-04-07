// Package saves provides persistence and HTTP handlers for campaign save points
// and campaign time tracking.
package saves

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/PatrickFanella/game-master/internal/db"
)

// SavePoint represents a row in the save_points table.
type SavePoint struct {
	ID         uuid.UUID
	CampaignID uuid.UUID
	Name       string
	TurnNumber int
	IsAuto     bool
	CreatedAt  time.Time
}

// CampaignTime represents a row in the campaign_time table.
type CampaignTime struct {
	CampaignID uuid.UUID
	Day        int
	Hour       int
	Minute     int
	UpdatedAt  time.Time
}

// Store provides direct save_points and campaign_time operations using raw SQL.
type Store struct {
	db db.DBTX
}

// NewStore creates a Store backed by the given database connection.
func NewStore(conn db.DBTX) *Store {
	return &Store{db: conn}
}

const createSavePointSQL = `
INSERT INTO save_points (campaign_id, name, turn_number, is_auto)
VALUES ($1, $2, $3, $4)
RETURNING id, campaign_id, name, turn_number, is_auto, created_at
`

// CreateSavePoint inserts a new save point and returns it.
func (s *Store) CreateSavePoint(ctx context.Context, campaignID uuid.UUID, name string, turnNumber int, isAuto bool) (SavePoint, error) {
	pgCID := db.ToPgUUID(campaignID)
	row := s.db.QueryRow(ctx, createSavePointSQL, pgCID, name, turnNumber, isAuto)
	var sp SavePoint
	var pgID, pgCampaignID pgtype.UUID
	var pgCreatedAt pgtype.Timestamptz
	err := row.Scan(&pgID, &pgCampaignID, &sp.Name, &sp.TurnNumber, &sp.IsAuto, &pgCreatedAt)
	if err != nil {
		return SavePoint{}, err
	}
	sp.ID = db.FromPgUUID(pgID)
	sp.CampaignID = db.FromPgUUID(pgCampaignID)
	if pgCreatedAt.Valid {
		sp.CreatedAt = pgCreatedAt.Time
	}
	return sp, nil
}

const listSavePointsSQL = `
SELECT id, campaign_id, name, turn_number, is_auto, created_at
FROM save_points
WHERE campaign_id = $1
ORDER BY created_at DESC
`

// ListSavePoints returns all save points for a campaign, newest first.
func (s *Store) ListSavePoints(ctx context.Context, campaignID uuid.UUID) ([]SavePoint, error) {
	pgCID := db.ToPgUUID(campaignID)
	rows, err := s.db.Query(ctx, listSavePointsSQL, pgCID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SavePoint
	for rows.Next() {
		var sp SavePoint
		var pgID, pgCampaignID pgtype.UUID
		var pgCreatedAt pgtype.Timestamptz
		if err := rows.Scan(&pgID, &pgCampaignID, &sp.Name, &sp.TurnNumber, &sp.IsAuto, &pgCreatedAt); err != nil {
			return nil, err
		}
		sp.ID = db.FromPgUUID(pgID)
		sp.CampaignID = db.FromPgUUID(pgCampaignID)
		if pgCreatedAt.Valid {
			sp.CreatedAt = pgCreatedAt.Time
		}
		results = append(results, sp)
	}
	return results, rows.Err()
}

const deleteOldAutoSavesSQL = `
DELETE FROM save_points
WHERE campaign_id = $1
  AND is_auto = true
  AND id NOT IN (
    SELECT id FROM save_points
    WHERE campaign_id = $1 AND is_auto = true
    ORDER BY created_at DESC
    LIMIT 3
  )
`

// DeleteOldAutoSaves removes auto-save points beyond the 3 most recent.
func (s *Store) DeleteOldAutoSaves(ctx context.Context, campaignID uuid.UUID) error {
	pgCID := db.ToPgUUID(campaignID)
	_, err := s.db.Exec(ctx, deleteOldAutoSavesSQL, pgCID)
	return err
}

const getLatestTurnNumberSQL = `
SELECT COALESCE(MAX(turn_number), 0) FROM session_logs WHERE campaign_id = $1
`

// GetLatestTurnNumber returns the highest turn number for a campaign's session logs.
func (s *Store) GetLatestTurnNumber(ctx context.Context, campaignID uuid.UUID) (int, error) {
	pgCID := db.ToPgUUID(campaignID)
	var turnNumber int
	err := s.db.QueryRow(ctx, getLatestTurnNumberSQL, pgCID).Scan(&turnNumber)
	return turnNumber, err
}

const deleteSessionLogsByCampaignSQL = `DELETE FROM session_logs WHERE campaign_id = $1`
const deleteSavePointsByCampaignSQL = `DELETE FROM save_points WHERE campaign_id = $1`
const deleteCampaignTimeSQL = `DELETE FROM campaign_time WHERE campaign_id = $1`

// StartOver deletes all session logs, save points, and campaign time for a campaign.
func (s *Store) StartOver(ctx context.Context, campaignID uuid.UUID) error {
	pgCID := db.ToPgUUID(campaignID)
	if _, err := s.db.Exec(ctx, deleteSessionLogsByCampaignSQL, pgCID); err != nil {
		return err
	}
	if _, err := s.db.Exec(ctx, deleteSavePointsByCampaignSQL, pgCID); err != nil {
		return err
	}
	_, err := s.db.Exec(ctx, deleteCampaignTimeSQL, pgCID)
	return err
}

const getCampaignTimeSQL = `
SELECT campaign_id, day, hour, minute, updated_at
FROM campaign_time
WHERE campaign_id = $1
`

// GetCampaignTime returns the current campaign time. Returns a default (day 1, 08:00)
// if no row exists.
func (s *Store) GetCampaignTime(ctx context.Context, campaignID uuid.UUID) (CampaignTime, error) {
	pgCID := db.ToPgUUID(campaignID)
	row := s.db.QueryRow(ctx, getCampaignTimeSQL, pgCID)
	var ct CampaignTime
	var pgCampaignID pgtype.UUID
	var pgUpdatedAt pgtype.Timestamptz
	err := row.Scan(&pgCampaignID, &ct.Day, &ct.Hour, &ct.Minute, &pgUpdatedAt)
	if err != nil {
		return CampaignTime{
			CampaignID: campaignID,
			Day:        1,
			Hour:       8,
			Minute:     0,
		}, err
	}
	ct.CampaignID = db.FromPgUUID(pgCampaignID)
	if pgUpdatedAt.Valid {
		ct.UpdatedAt = pgUpdatedAt.Time
	}
	return ct, nil
}

const upsertCampaignTimeSQL = `
INSERT INTO campaign_time (campaign_id, day, hour, minute)
VALUES ($1, $2, $3, $4)
ON CONFLICT (campaign_id) DO UPDATE
SET day = EXCLUDED.day, hour = EXCLUDED.hour, minute = EXCLUDED.minute, updated_at = now()
RETURNING campaign_id, day, hour, minute, updated_at
`

// UpsertCampaignTime creates or updates the campaign time.
func (s *Store) UpsertCampaignTime(ctx context.Context, campaignID uuid.UUID, day, hour, minute int) (CampaignTime, error) {
	pgCID := db.ToPgUUID(campaignID)
	row := s.db.QueryRow(ctx, upsertCampaignTimeSQL, pgCID, day, hour, minute)
	var ct CampaignTime
	var pgCampaignID pgtype.UUID
	var pgUpdatedAt pgtype.Timestamptz
	err := row.Scan(&pgCampaignID, &ct.Day, &ct.Hour, &ct.Minute, &pgUpdatedAt)
	if err != nil {
		return CampaignTime{}, err
	}
	ct.CampaignID = db.FromPgUUID(pgCampaignID)
	if pgUpdatedAt.Valid {
		ct.UpdatedAt = pgUpdatedAt.Time
	}
	return ct, nil
}
