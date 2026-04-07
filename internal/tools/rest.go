package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/PatrickFanella/game-master/internal/llm"
)

const restToolName = "rest"

// RestTool returns the rest tool definition and JSON schema.
func RestTool() llm.Tool {
	return llm.Tool{
		Name:        restToolName,
		Description: "Rest the player character for a number of hours. Short rest (1 hour) restores 25% max HP; long rest (up to 8 hours) restores full HP. Advances in-game time accordingly.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"hours": map[string]any{
					"type":        "integer",
					"description": "Number of hours to rest (1-8).",
					"minimum":     1,
					"maximum":     8,
				},
				"short_rest": map[string]any{
					"type":        "boolean",
					"description": "If true, perform a short rest (1 hour, 25% HP recovery). Defaults to false (long rest, full HP recovery).",
				},
			},
			"required":             []string{"hours"},
			"additionalProperties": false,
		},
	}
}

// RegisterRest registers the rest tool and handler.
func RegisterRest(reg *Registry, db TimeStore) error {
	if db == nil {
		return errors.New("rest database connection is required")
	}
	return reg.Register(RestTool(), NewRestHandler(db).Handle)
}

// RestHandler executes rest tool calls.
type RestHandler struct {
	db TimeStore
}

// NewRestHandler creates a new rest handler.
func NewRestHandler(db TimeStore) *RestHandler {
	return &RestHandler{db: db}
}

const getPlayerHPSQL = `
SELECT hp, max_hp FROM player_characters WHERE campaign_id = $1 LIMIT 1
`

const updatePlayerHPSQL = `
UPDATE player_characters SET hp = $1 WHERE campaign_id = $2
`

// Handle executes the rest tool.
func (h *RestHandler) Handle(ctx context.Context, args map[string]any) (*ToolResult, error) {
	if h == nil {
		return nil, errors.New("rest handler is nil")
	}

	hours, err := parseIntArg(args, "hours")
	if err != nil {
		return nil, err
	}
	if hours < 1 || hours > 8 {
		return nil, errors.New("hours must be between 1 and 8")
	}

	shortRest, err := parseBoolArg(args, "short_rest")
	if err != nil {
		return nil, err
	}

	campaignID, ok := CurrentCampaignIDFromContext(ctx)
	if !ok {
		return nil, errors.New("rest requires current campaign id in context")
	}

	pgCID := pgtype.UUID{Bytes: campaignID, Valid: campaignID != uuid.Nil}

	// Read current HP.
	var hp, maxHP int
	err = h.db.QueryRow(ctx, getPlayerHPSQL, pgCID).Scan(&hp, &maxHP)
	if err != nil {
		return nil, fmt.Errorf("read player HP: %w", err)
	}

	// Calculate healing.
	var newHP int
	if shortRest {
		healing := maxHP / 4
		if healing < 1 {
			healing = 1
		}
		newHP = hp + healing
		if newHP > maxHP {
			newHP = maxHP
		}
	} else {
		newHP = maxHP
	}

	// Update HP.
	_, err = h.db.Exec(ctx, updatePlayerHPSQL, newHP, pgCID)
	if err != nil {
		return nil, fmt.Errorf("update player HP: %w", err)
	}

	// Advance time.
	var day, hour, minute int
	err = h.db.QueryRow(ctx, getCampaignTimeForToolSQL, pgCID).Scan(&day, &hour, &minute)
	if errors.Is(err, pgx.ErrNoRows) {
		day, hour, minute = 1, 8, 0
	} else if err != nil {
		return nil, fmt.Errorf("read campaign time: %w", err)
	}

	totalHours := hour + hours
	day += totalHours / 24
	hour = totalHours % 24

	err = h.db.QueryRow(ctx, upsertCampaignTimeForToolSQL, pgCID, day, hour, minute).Scan(&day, &hour, &minute)
	if err != nil {
		return nil, fmt.Errorf("update campaign time: %w", err)
	}

	restType := "long rest"
	if shortRest {
		restType = "short rest"
	}

	narrative := fmt.Sprintf("You take a %s for %s. HP restored to %d/%d. It is now Day %d, %02d:%02d.",
		restType, formatDuration(hours, 0), newHP, maxHP, day, hour, minute)

	return &ToolResult{
		Success: true,
		Data: map[string]any{
			"hp":        newHP,
			"max_hp":    maxHP,
			"rest_type": restType,
			"day":       day,
			"hour":      hour,
			"minute":    minute,
		},
		Narrative: narrative,
	}, nil
}
