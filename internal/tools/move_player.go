package tools

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/PatrickFanella/game-master/internal/llm"
)

const movePlayerToolName = "move_player"

// MovePlayerStore provides persistence and relationship checks for player movement.
type MovePlayerStore interface {
	GetLocation(ctx context.Context, locationID uuid.UUID) (name, description string, err error)
	IsLocationConnected(ctx context.Context, fromLocationID, toLocationID uuid.UUID) (bool, error)
	UpdatePlayerLocation(ctx context.Context, playerCharacterID, locationID uuid.UUID) error
	SetLocationPlayerVisited(ctx context.Context, locationID uuid.UUID) error
	GetConnectionTravelTime(ctx context.Context, fromLocationID, toLocationID uuid.UUID) (string, error)
}

// MovePlayerTool returns the move_player tool definition and JSON schema.
func MovePlayerTool() llm.Tool {
	return llm.Tool{
		Name:        movePlayerToolName,
		Description: "Move the player character to a connected location.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"location_id": map[string]any{
					"type":        "string",
					"description": "Destination location UUID.",
				},
			},
			"required":             []string{"location_id"},
			"additionalProperties": false,
		},
	}
}

// RegisterMovePlayer registers the move_player tool and handler.
func RegisterMovePlayer(reg *Registry, store MovePlayerStore, timeDB ...TimeStore) error {
	if store == nil {
		return errors.New("move_player store is required")
	}
	var ts TimeStore
	if len(timeDB) > 0 {
		ts = timeDB[0]
	}
	return reg.Register(MovePlayerTool(), NewMovePlayerHandler(store, ts).Handle)
}

// MovePlayerHandler executes move_player tool calls.
type MovePlayerHandler struct {
	store  MovePlayerStore
	timeDB TimeStore
}

// NewMovePlayerHandler creates a new move_player handler.
func NewMovePlayerHandler(store MovePlayerStore, timeDB TimeStore) *MovePlayerHandler {
	return &MovePlayerHandler{store: store, timeDB: timeDB}
}

// Handle executes the move_player tool.
func (h *MovePlayerHandler) Handle(ctx context.Context, args map[string]any) (*ToolResult, error) {
	if h == nil {
		return nil, errors.New("move_player handler is nil")
	}
	if h.store == nil {
		return nil, errors.New("move_player store is required")
	}

	targetLocationID, err := parseUUIDArg(args, "location_id")
	if err != nil {
		return nil, err
	}

	locationName, locationDescription, err := h.store.GetLocation(ctx, targetLocationID)
	if err != nil {
		return nil, fmt.Errorf("get target location: %w", err)
	}

	currentLocationID, ok := CurrentLocationIDFromContext(ctx)
	if !ok {
		return nil, errors.New("move_player requires current location id in context")
	}

	isConnected, err := h.store.IsLocationConnected(ctx, currentLocationID, targetLocationID)
	if err != nil {
		return nil, fmt.Errorf("check location connection: %w", err)
	}
	if !isConnected {
		return nil, errors.New("target location is not connected to current location")
	}

	playerCharacterID, ok := CurrentPlayerCharacterIDFromContext(ctx)
	if !ok {
		return nil, errors.New("move_player requires current player character id in context")
	}

	if err := h.store.UpdatePlayerLocation(ctx, playerCharacterID, targetLocationID); err != nil {
		return nil, fmt.Errorf("update player location: %w", err)
	}

	_ = h.store.SetLocationPlayerVisited(ctx, targetLocationID)

	// Deduct travel time if time tracking is available.
	var travelNarrative string
	if h.timeDB != nil {
		travelTime, ttErr := h.store.GetConnectionTravelTime(ctx, currentLocationID, targetLocationID)
		if ttErr == nil && travelTime != "" {
			advHours, advMinutes := parseTravelTime(travelTime)
			if advHours > 0 || advMinutes > 0 {
				campaignID, hasCampaign := CurrentCampaignIDFromContext(ctx)
				if hasCampaign {
					pgCID := pgtype.UUID{Bytes: campaignID, Valid: campaignID != uuid.Nil}
					var day, hour, minute int
					scanErr := h.timeDB.QueryRow(ctx, getCampaignTimeForToolSQL, pgCID).Scan(&day, &hour, &minute)
					if errors.Is(scanErr, pgx.ErrNoRows) {
						day, hour, minute = 1, 8, 0
					} else if scanErr != nil {
						// Non-fatal; skip time advancement.
						goto skipTime
					}

					totalMinutes := minute + advMinutes
					hour += totalMinutes / 60
					minute = totalMinutes % 60
					totalHours := hour + advHours
					day += totalHours / 24
					hour = totalHours % 24

					scanErr = h.timeDB.QueryRow(ctx, upsertCampaignTimeForToolSQL, pgCID, day, hour, minute).Scan(&day, &hour, &minute)
					if scanErr == nil {
						travelNarrative = fmt.Sprintf(" The journey took %s. It is now Day %d, %02d:%02d.", travelTime, day, hour, minute)
					}
				}
			}
		}
	}
skipTime:

	return &ToolResult{
		Success: true,
		Data: map[string]any{
			"location_id": targetLocationID.String(),
			"name":        locationName,
			"description": locationDescription,
		},
		Narrative: fmt.Sprintf("Player moved to %s.%s", locationName, travelNarrative),
	}, nil
}

// parseTravelTime parses a human-readable travel time string into hours and minutes.
// Supports formats like "2 hours", "30 minutes", "1 hour 30 minutes", "2h", "30m".
func parseTravelTime(s string) (hours, minutes int) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return 1, 0 // default: 1 hour
	}

	numRe := regexp.MustCompile(`(\d+)`)

	if strings.Contains(s, "hour") {
		matches := numRe.FindString(s)
		if matches != "" {
			n, _ := strconv.Atoi(matches)
			hours = n
		} else {
			hours = 1
		}
		// Also check for minutes in the same string.
		if idx := strings.Index(s, "minute"); idx >= 0 {
			sub := s[idx-5:]
			if m := numRe.FindString(sub); m != "" {
				n, _ := strconv.Atoi(m)
				minutes = n
			}
		}
		return hours, minutes
	}

	if strings.Contains(s, "minute") || strings.Contains(s, "min") {
		matches := numRe.FindString(s)
		if matches != "" {
			n, _ := strconv.Atoi(matches)
			return 0, n
		}
		return 0, 30 // default: 30 minutes
	}

	// Default: 1 hour.
	return 1, 0
}
