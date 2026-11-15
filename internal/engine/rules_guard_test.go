package engine

import (
	"context"
	"errors"
	"testing"

	"git.subcult.tv/subculture-collective/edda/internal/domain"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"git.subcult.tv/subculture-collective/edda/internal/tools"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type rulesGuardQuerier struct {
	statedb.Querier
	version string
}

func (q rulesGuardQuerier) GetCampaignByID(context.Context, pgtype.UUID) (statedb.Campaign, error) {
	return statedb.Campaign{RulesVersion: q.version}, nil
}

func TestProductionToolsRefuseVersionedCampaigns(t *testing.T) {
	for _, version := range []string{"srd-5.2.1", "future-version"} {
		t.Run(version, func(t *testing.T) {
			registry := tools.NewRegistry()
			if err := registerAllTools(registry, rulesGuardQuerier{version: version}, nil, nil); err != nil {
				t.Fatal(err)
			}
			ctx := tools.WithCurrentCampaignID(context.Background(), uuid.New())
			for _, tool := range registry.List() {
				if _, err := registry.Execute(ctx, tool.Name, nil); !errors.Is(err, domain.ErrLegacyRulesMutation) {
					t.Errorf("%s = %v", tool.Name, err)
				}
			}
		})
	}
}
