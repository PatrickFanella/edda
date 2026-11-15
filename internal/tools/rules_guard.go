package tools

import (
	"context"
	"errors"
	"fmt"

	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type campaignRulesReader interface {
	GetCampaignByID(context.Context, pgtype.UUID) (statedb.Campaign, error)
}

// RequireLegacyCampaign guards execution of this registry's legacy tools using
// persisted campaign metadata, even if tool filtering or engine gathering is bypassed.
// Call it before sharing the registry with any concurrent callers.
func (r *Registry) RequireLegacyCampaign(q campaignRulesReader) {
	r.beforeExecute = func(ctx context.Context) error {
		id, ok := CurrentCampaignIDFromContext(ctx)
		if !ok {
			return errors.New("current campaign is required for legacy tool execution")
		}
		campaign, err := q.GetCampaignByID(ctx, dbutil.ToPgtype(id))
		if err != nil {
			return fmt.Errorf("load tool campaign rules: %w", err)
		}
		return domain.RequireLegacyRules(domain.RulesVersion(campaign.RulesVersion))
	}
}
