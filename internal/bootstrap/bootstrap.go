// Package bootstrap handles first-boot setup for the Game Master TUI.
// On first run it creates a default local user. On subsequent runs it returns
// existing campaigns for the user so the TUI can show a selection list.
package bootstrap

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
)

const (
	// DefaultUserName is the name given to the auto-created local user.
	DefaultUserName = "Player"

	// DefaultCampaignName is the default name used when creating a campaign
	// through engine-level flows.
	DefaultCampaignName = "The Beginning"
)

// Result holds the user and available campaigns returned by Run.
type Result struct {
	User      statedb.User
	Campaigns []statedb.Campaign
}

// Run ensures a default user exists in the database and returns all campaigns
// currently owned by that user. If no user named DefaultUserName is found, it
// creates one.
func Run(ctx context.Context, q statedb.Querier) (Result, error) {
	user, err := findOrCreateUser(ctx, q, DefaultUserName)
	if err != nil {
		return Result{}, fmt.Errorf("bootstrap user: %w", err)
	}

	campaigns, err := q.ListCampaignsByUser(ctx, user.ID)
	if err != nil {
		return Result{}, fmt.Errorf("list campaigns: %w", err)
	}

	return Result{User: user, Campaigns: campaigns}, nil
}

// findOrCreateUser returns the user matching name, or creates one.
// It uses GetUserByName for efficiency rather than a full table scan.
func findOrCreateUser(ctx context.Context, q statedb.Querier, name string) (statedb.User, error) {
	u, err := q.GetUserByName(ctx, name)
	if err == nil {
		return u, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return statedb.User{}, fmt.Errorf("get user by name: %w", err)
	}
	return q.CreateUser(ctx, name)
}
