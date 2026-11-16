package srd521

import (
	"fmt"
	"math"
	"strconv"
)

// EncounterBudgetSourceID identifies the budget table and spending procedure (PDF p. 202).
const EncounterBudgetSourceID = "srd-5.2.1:toolbox/encounter-budget"

// EncounterXPSourceID identifies Experience Points by Challenge Rating (PDF p. 256).
const EncounterXPSourceID = "srd-5.2.1:monsters/experience-points-by-challenge-rating"

// EncounterPartyGroup explicitly supplies characters of one level. Counts must be positive.
type EncounterPartyGroup struct {
	Level int   `json:"level"`
	Count int64 `json:"count"`
}

// EncounterMonsterGroup supplies selected creatures, never generated creatures.
// XP must be explicit, including CR 0's choice of 0 or 10. MonsterID is an
// optional provenance label; it does not grant playable support.
type EncounterMonsterGroup struct {
	MonsterID string `json:"monster_id,omitempty"`
	CR        string `json:"cr"`
	XP        *int64 `json:"xp"`
	Count     int64  `json:"count"`
}

// EncounterBudgets contains XP spending ceilings, not XP rewards.
type EncounterBudgets struct {
	Low      int64 `json:"low"`
	Moderate int64 `json:"moderate"`
	High     int64 `json:"high"`
}

// EncounterPartyDerivation records each table lookup and multiplication.
type EncounterPartyDerivation struct {
	EncounterPartyGroup
	PerCharacter EncounterBudgets `json:"per_character"`
	Subtotal     EncounterBudgets `json:"subtotal"`
	SourceID     string           `json:"source_id"`
}

// EncounterMonsterDerivation records each CR/XP check and multiplication.
type EncounterMonsterDerivation struct {
	MonsterID string `json:"monster_id,omitempty"`
	CR        string `json:"cr"`
	XP        int64  `json:"xp"`
	Count     int64  `json:"count"`
	Subtotal  int64  `json:"subtotal"`
	SourceID  string `json:"source_id"`
}

// EncounterBudgetResult is an estimate relative to spending ceilings. Difficulty
// is low, moderate, high, or above-high: the first ceiling that fits wins.
// Empty selections yield zero XP and low. Qualitative risks on p. 203 are not
// modeled. The result neither starts combat nor authorizes rewards or NPC rules.
type EncounterBudgetResult struct {
	Source      Source                       `json:"source"`
	SourcePages []int                        `json:"source_pages"`
	Party       []EncounterPartyDerivation   `json:"party"`
	Monsters    []EncounterMonsterDerivation `json:"monsters"`
	Budgets     EncounterBudgets             `json:"budgets"`
	TotalXP     int64                        `json:"total_xp"`
	Difficulty  string                       `json:"difficulty"`
}

// Transcribed from XP Budget per Character, PDF p. 202. Index zero is unused.
var encounterBudgetRows = [...]EncounterBudgets{
	{}, {50, 75, 100}, {100, 150, 200}, {150, 225, 400}, {250, 375, 500},
	{500, 750, 1100}, {600, 1000, 1400}, {750, 1300, 1700}, {1000, 1700, 2100},
	{1300, 2000, 2600}, {1600, 2300, 3100}, {1900, 2900, 4100}, {2200, 3700, 4700},
	{2600, 4200, 5400}, {2900, 4900, 6200}, {3300, 5400, 7800}, {3800, 6100, 9800},
	{4500, 7200, 11700}, {5000, 8700, 14200}, {5500, 10700, 17200}, {6400, 13200, 22000},
}

// Transcribed from Experience Points by Challenge Rating, PDF p. 256.
var encounterIntegerXP = [...]int64{0, 200, 450, 700, 1100, 1800, 2300, 2900, 3900, 5000, 5900, 7200, 8400, 10000, 11500, 13000, 15000, 18000, 20000, 22000, 25000, 33000, 41000, 50000, 62000, 75000, 90000, 105000, 120000, 135000, 155000}

func encounterValidXP(cr string, xp int64) bool {
	switch cr {
	case "0":
		return xp == 0 || xp == 10
	case "1/8":
		return xp == 25
	case "1/4":
		return xp == 50
	case "1/2":
		return xp == 100
	}
	n, err := strconv.Atoi(cr)
	return err == nil && n >= 1 && n <= 30 && strconv.Itoa(n) == cr && encounterIntegerXP[n] == xp
}

// EncounterMonsterFromRecord copies a validated typed monster's CR/XP into a
// calculation input. Available is deliberately not consulted: catalogue data
// can inform a budget even when its combat mechanics remain unavailable.
func EncounterMonsterFromRecord(m Monster, count int64) (EncounterMonsterGroup, error) {
	if err := m.Validate(); err != nil {
		return EncounterMonsterGroup{}, fmt.Errorf("encounter monster: %w", err)
	}
	xp := int64(m.XP)
	if count <= 0 || !encounterValidXP(m.CR, xp) {
		return EncounterMonsterGroup{}, fmt.Errorf("invalid encounter monster count or CR/XP")
	}
	return EncounterMonsterGroup{MonsterID: m.ID, CR: m.CR, XP: &xp, Count: count}, nil
}

func encounterSource() Source {
	return Source{
		Version: Version, Language: "en", LandingURL: "https://www.dndbeyond.com/srd",
		ArtifactURL: "https://media.dndbeyond.com/compendium-images/srd/5.2/SRD_CC_v5.2.1.pdf",
		SHA256:      artifactSHA256, Bytes: artifactBytes, Pages: 364, RetrievedAt: "2026-11-15",
		License: "CC-BY-4.0", LicenseURL: "https://creativecommons.org/licenses/by/4.0/legalcode",
		Attribution: Attribution,
		Changes:     "Adapted PDF pp. 202–203 and 256 into a read-only budget calculator. Mixed-level parties sum individual level budgets. Difficulty uses the smallest spending ceiling that contains total XP; above-high is a local out-of-budget label. Qualitative risks are not modeled; no combat, rewards, generation or loot is authorized.",
	}
}

// CalculateEncounterBudget sums per-level party budgets and selected monster XP.
// It accepts levels 1–20 and canonical CR strings 0, 1/8, 1/4, 1/2, 1–30.
// Missing/null XP, mismatched CR/XP, nonpositive counts, empty parties and
// int64 overflow fail with a zero result. An empty monster selection is valid.
// No caller data is retained or mutated, and no external effects are performed.
func CalculateEncounterBudget(party []EncounterPartyGroup, monsters []EncounterMonsterGroup) (EncounterBudgetResult, error) {
	fail := func(message string) (EncounterBudgetResult, error) {
		return EncounterBudgetResult{}, fmt.Errorf("encounter budget: %s", message)
	}
	if len(party) == 0 {
		return fail("explicit nonempty party required")
	}
	r := EncounterBudgetResult{Source: encounterSource(), SourcePages: []int{202, 203, 256}, Party: make([]EncounterPartyDerivation, 0, len(party)), Monsters: make([]EncounterMonsterDerivation, 0, len(monsters))}
	for _, g := range party {
		if g.Level < 1 || g.Level > 20 || g.Count <= 0 {
			return fail("unsupported level or nonpositive party count")
		}
		row := encounterBudgetRows[g.Level]
		// The high ceiling dominates every multiplication and accumulated budget.
		if g.Count > (math.MaxInt64-r.Budgets.High)/row.High {
			return fail("party budget overflow")
		}
		sub := EncounterBudgets{row.Low * g.Count, row.Moderate * g.Count, row.High * g.Count}
		r.Budgets.Low += sub.Low
		r.Budgets.Moderate += sub.Moderate
		r.Budgets.High += sub.High
		r.Party = append(r.Party, EncounterPartyDerivation{g, row, sub, EncounterBudgetSourceID})
	}
	for _, g := range monsters {
		if g.XP == nil || g.Count <= 0 {
			return fail("explicit XP and positive monster count required")
		}
		xp := *g.XP
		if !encounterValidXP(g.CR, xp) {
			return fail("unsupported or mismatched monster CR/XP")
		}
		if xp > 0 && g.Count > (math.MaxInt64-r.TotalXP)/xp {
			return fail("monster XP overflow")
		}
		sub := xp * g.Count
		r.TotalXP += sub
		r.Monsters = append(r.Monsters, EncounterMonsterDerivation{g.MonsterID, g.CR, xp, g.Count, sub, EncounterXPSourceID})
	}
	switch {
	case r.TotalXP <= r.Budgets.Low:
		r.Difficulty = "low"
	case r.TotalXP <= r.Budgets.Moderate:
		r.Difficulty = "moderate"
	case r.TotalXP <= r.Budgets.High:
		r.Difficulty = "high"
	default:
		r.Difficulty = "above-high"
	}
	return r, nil
}
