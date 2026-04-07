package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/PatrickFanella/game-master/internal/llm"
)

const advanceTimeToolName = "advance_time"

// TimeStore provides direct campaign_time operations using raw SQL.
type TimeStore interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

// AdvanceTimeTool returns the advance_time tool definition and JSON schema.
func AdvanceTimeTool() llm.Tool {
	return llm.Tool{
		Name:        advanceTimeToolName,
		Description: "Advance the in-game clock by the given hours and/or minutes. Handles day rollover automatically.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"hours": map[string]any{
					"type":        "integer",
					"description": "Number of hours to advance (0 or more).",
				},
				"minutes": map[string]any{
					"type":        "integer",
					"description": "Number of minutes to advance (0 or more, optional, defaults to 0).",
				},
			},
			"required":             []string{"hours"},
			"additionalProperties": false,
		},
	}
}

// RegisterAdvanceTime registers the advance_time tool and handler.
func RegisterAdvanceTime(reg *Registry, db TimeStore) error {
	if db == nil {
		return errors.New("advance_time database connection is required")
	}
	return reg.Register(AdvanceTimeTool(), NewAdvanceTimeHandler(db).Handle)
}

// AdvanceTimeHandler executes advance_time tool calls.
type AdvanceTimeHandler struct {
	db TimeStore
}

// NewAdvanceTimeHandler creates a new advance_time handler.
func NewAdvanceTimeHandler(db TimeStore) *AdvanceTimeHandler {
	return &AdvanceTimeHandler{db: db}
}

const getCampaignTimeForToolSQL = `
SELECT day, hour, minute FROM campaign_time WHERE campaign_id = $1
`

const upsertCampaignTimeForToolSQL = `
INSERT INTO campaign_time (campaign_id, day, hour, minute)
VALUES ($1, $2, $3, $4)
ON CONFLICT (campaign_id) DO UPDATE
SET day = EXCLUDED.day, hour = EXCLUDED.hour, minute = EXCLUDED.minute, updated_at = now()
RETURNING day, hour, minute
`

// Handle executes the advance_time tool.
func (h *AdvanceTimeHandler) Handle(ctx context.Context, args map[string]any) (*ToolResult, error) {
	if h == nil {
		return nil, errors.New("advance_time handler is nil")
	}

	hours, err := parseIntArg(args, "hours")
	if err != nil {
		return nil, err
	}
	if hours < 0 {
		return nil, errors.New("hours must be 0 or more")
	}

	minutes := 0
	if _, ok := args["minutes"]; ok {
		minutes, err = parseIntArg(args, "minutes")
		if err != nil {
			return nil, err
		}
		if minutes < 0 {
			return nil, errors.New("minutes must be 0 or more")
		}
	}

	if hours == 0 && minutes == 0 {
		return nil, errors.New("at least one of hours or minutes must be greater than 0")
	}

	campaignID, ok := CurrentCampaignIDFromContext(ctx)
	if !ok {
		return nil, errors.New("advance_time requires current campaign id in context")
	}

	pgCID := pgtype.UUID{Bytes: campaignID, Valid: campaignID != uuid.Nil}

	// Read current time (defaults if not set).
	var day, hour, minute int
	err = h.db.QueryRow(ctx, getCampaignTimeForToolSQL, pgCID).Scan(&day, &hour, &minute)
	if errors.Is(err, pgx.ErrNoRows) {
		day, hour, minute = 1, 8, 0
	} else if err != nil {
		return nil, fmt.Errorf("read campaign time: %w", err)
	}

	// Advance.
	totalMinutes := minute + minutes
	hour += totalMinutes / 60
	minute = totalMinutes % 60

	totalHours := hour + hours
	day += totalHours / 24
	hour = totalHours % 24

	// Write back.
	err = h.db.QueryRow(ctx, upsertCampaignTimeForToolSQL, pgCID, day, hour, minute).Scan(&day, &hour, &minute)
	if err != nil {
		return nil, fmt.Errorf("update campaign time: %w", err)
	}

	narrative := fmt.Sprintf("%s pass. It is now Day %d, %02d:%02d.", formatDuration(hours, minutes), day, hour, minute)

	return &ToolResult{
		Success: true,
		Data: map[string]any{
			"day":    day,
			"hour":   hour,
			"minute": minute,
		},
		Narrative: narrative,
	}, nil
}

func formatDuration(hours, minutes int) string {
	parts := make([]string, 0, 2)
	if hours > 0 {
		if hours == 1 {
			parts = append(parts, "1 hour")
		} else {
			parts = append(parts, fmt.Sprintf("%d hours", hours))
		}
	}
	if minutes > 0 {
		if minutes == 1 {
			parts = append(parts, "1 minute")
		} else {
			parts = append(parts, fmt.Sprintf("%d minutes", minutes))
		}
	}
	if len(parts) == 0 {
		return "0 minutes"
	}
	if len(parts) == 2 {
		return parts[0] + " and " + parts[1]
	}
	return parts[0]
}
