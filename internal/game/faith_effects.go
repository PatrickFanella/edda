package game

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"git.subcult.tv/subculture-collective/edda/internal/combat"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"github.com/google/uuid"
)

// FaithEffect is one durable cast and its independent recipient link. Only the
// selected effect is needed to derive AC; suppressed casts retain their clocks.
type FaithEffect struct {
	ID          uuid.UUID
	CasterID    uuid.UUID
	RecipientID uuid.UUID
	StartedAt   int
	ExpiresAt   int
}

// ReadFaithEffect chooses the latest live fixed +2 effect by persisted cast order.
// The campaign lock serializes confirmations, including casts in the same second.
func ReadFaithEffect(ctx context.Context, database statedb.DBTX, campaign, recipient uuid.UUID, now int) (*FaithEffect, error) {
	rows, err := database.Query(ctx, `SELECT c.id,c.caster_id,r.recipient_id,c.started_at,c.expires_at FROM concentration_casts c JOIN concentration_recipients r ON r.cast_id=c.id JOIN player_characters p ON p.id=c.caster_id AND p.campaign_id=c.campaign_id WHERE c.campaign_id=$1 AND r.recipient_id=$2 AND c.started_at<=$3 AND c.expires_at>$3 AND p.hp>0 AND p.status='active' AND p.stats#>>'{srd521,casting,concentration,id}'=c.id::text ORDER BY c.cast_order DESC LIMIT 1`, campaign, recipient, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, rows.Err()
	}
	var e FaithEffect
	if err := rows.Scan(&e.ID, &e.CasterID, &e.RecipientID, &e.StartedAt, &e.ExpiresAt); err != nil {
		return nil, err
	}
	return &e, rows.Err()
}

// ApplyFaithAC replaces only the named spell term, preserving ordinary equipment
// and every other derivation. Legacy self-only saves remain readable.
func ApplyFaithAC(ctx context.Context, database statedb.DBTX, campaign, actor uuid.UUID, now int, sheet *srd521.CharacterSheet) error {
	e, err := ReadFaithEffect(ctx, database, campaign, actor, now)
	if err != nil {
		return err
	}
	ac := sheet.Defenses["armor-class"]
	ac.Derivation = append([]srd521.Term(nil), ac.Derivation...)
	terms := ac.Derivation[:0]
	for _, t := range ac.Derivation {
		if t.Label == "Shield of Faith" {
			ac.Value -= t.Value
		} else {
			terms = append(terms, t)
		}
	}
	ac.Derivation = terms
	legacy := sheet.Concentration
	if legacy != nil && legacy.ExpiresAt <= now {
		sheet.Concentration = nil
	}
	if e != nil || (legacy != nil && legacy.RecipientID == "" && legacy.StartedAt <= now && legacy.ExpiresAt > now) {
		ac.Value += 2
		// Keep the source-derived term shape owned by the rules package.
		ac.Derivation = append(ac.Derivation, srd521.FaithACTerm())
	}
	sheet.Defenses["armor-class"] = ac
	return nil
}

func prepareFaithTargets(ctx context.Context, database statedb.DBTX, campaign domain.Campaign, pc domain.PlayerCharacter, result *StoredAction, lock bool) (*StoredAction, error) {
	if pc.UserID != campaign.CreatedBy {
		result.Preview.Available = false
		result.Preview.Reason = "Shield of Faith requires an owned same-campaign caster."
		return result, nil
	}
	var envelope struct {
		Build srd521.Build `json:"srd521"`
	}
	if err := json.Unmarshal(pc.Stats, &envelope); err != nil {
		return nil, err
	}
	if !envelope.Build.Casting.IsolatedRecipientEffects {
		return result, nil
	}
	locking := ""
	if lock {
		locking = " FOR UPDATE OF n"
	}
	rows, err := database.Query(ctx, `SELECT n.id,n.name,n.hp,n.updated_at,n.properties->'srd521_faith' FROM npcs n WHERE n.campaign_id=$1 AND n.location_id=$2 AND n.alive AND n.hp>0 AND EXISTS(SELECT 1 FROM session_logs l WHERE l.campaign_id=n.campaign_id AND n.id=ANY(l.npcs_involved)) ORDER BY n.id`+locking, campaign.ID, pc.CurrentLocationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var target HealingTarget
		var raw []byte
		if err := rows.Scan(&target.ID, &target.Name, &target.HP, &target.UpdatedAt, &raw); err != nil {
			return nil, err
		}
		var facts struct {
			Profile  string      `json:"profile"`
			Actors   []uuid.UUID `json:"actor_ids"`
			Location uuid.UUID   `json:"location_id"`
			Hidden   *bool       `json:"hidden"`
			Distance *int        `json:"distance_feet"`
			Sight    *bool       `json:"line_of_sight"`
			Cover    string      `json:"cover"`
		}
		var fields map[string]json.RawMessage
		if json.Unmarshal(raw, &fields) != nil || len(fields) != 7 {
			continue
		}
		valid := true
		for _, key := range []string{"profile", "actor_ids", "location_id", "hidden", "distance_feet", "line_of_sight", "cover"} {
			if len(fields[key]) == 0 || string(fields[key]) == "null" {
				valid = false
			}
		}
		if !valid {
			continue
		}
		if json.Unmarshal(raw, &facts) != nil || facts.Profile != "shield-of-faith-isolated-v1" || facts.Hidden == nil || *facts.Hidden || pc.CurrentLocationID == nil || facts.Location != *pc.CurrentLocationID {
			continue
		}
		allowed := false
		for _, id := range facts.Actors {
			if id == pc.ID {
				allowed = true
			}
		}
		target.Space = combat.Space{DistanceFeet: facts.Distance, LineOfSight: facts.Sight, Cover: facts.Cover}
		if !allowed || target.Space.Validate() != nil || facts.Distance == nil || *facts.Distance > 60 || facts.Sight == nil || !*facts.Sight || facts.Cover == "total" {
			continue
		}
		target.Space.Revision = fmt.Sprintf("%x", sha256.Sum256(raw))
		result.HealingTargets = append(result.HealingTargets, target)
		result.Preview.Targets = append(result.Preview.Targets, ActionTarget{ID: target.ID, Label: target.Name + " (Shield of Faith; known creature)", DistanceFeet: facts.Distance, ReachFeet: 60})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	data, err := json.Marshal([]any{result.Preview.ExpectedRevision, result.HealingTargets})
	if err != nil {
		return nil, err
	}
	result.Preview.ExpectedRevision = fmt.Sprintf("faith-v1:%x", sha256.Sum256(data))
	result.Preview.Options = []string{"Shield of Faith; self or one explicitly entitled isolated known creature within 60 feet and clear path; Bonus Action; V/S and noncostly Holy Symbol focus; one level 1 slot; concentration up to 600 seconds. Equal +2 AC applies once; latest confirmed cast wins, including casts in the same second. Ending costs no action or refund."}
	return result, nil
}

// faithRevision fingerprints and, at confirmation, locks all participating
// casters, cast identities and recipient links in deterministic order. The
// campaign row lock serializes confirmations from independent engine instances.
func faithRevision(ctx context.Context, database statedb.DBTX, campaign uuid.UUID, lock bool) (string, error) {
	suffix := ""
	if lock {
		suffix = " FOR UPDATE"
	}
	parts := []json.RawMessage{}
	for _, query := range []string{
		`SELECT to_jsonb(p) FROM player_characters p WHERE campaign_id=$1 ORDER BY id`,
		`SELECT to_jsonb(c) FROM concentration_casts c WHERE campaign_id=$1 ORDER BY id`,
		`SELECT to_jsonb(r) FROM concentration_recipients r WHERE cast_id IN (SELECT id FROM concentration_casts WHERE campaign_id=$1) ORDER BY cast_id`,
	} {
		rows, err := database.Query(ctx, query+suffix, campaign)
		if err != nil {
			return "", err
		}
		for rows.Next() {
			var raw []byte
			if err := rows.Scan(&raw); err != nil {
				rows.Close()
				return "", err
			}
			parts = append(parts, raw)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return "", err
		}
	}
	data, err := json.Marshal(parts)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}
