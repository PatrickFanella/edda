package srd521

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"
)

//go:embed testdata/monsters.json
var monsterJSON []byte

// MonsterCatalogue contains source-backed data, not permission to run NPC rules.
type MonsterCatalogue struct {
	SchemaVersion int       `json:"schema_version"`
	Source        Source    `json:"source"`
	Monsters      []Monster `json:"monsters"`
}

// Monster is a mechanical transcription. SourcePages is an inclusive PDF range.
// Traits and effects must be retained even when their mechanics are unavailable.
type Monster struct {
	ID                string                    `json:"id"`
	Name              string                    `json:"name"`
	SourcePages       []int                     `json:"source_pages"`
	Sizes             []string                  `json:"sizes"`
	Type              string                    `json:"type"`
	Alignment         string                    `json:"alignment"`
	AC                int                       `json:"ac"`
	HP                MonsterHP                 `json:"hp"`
	Initiative        MonsterInitiative         `json:"initiative"`
	Speeds            map[string]int            `json:"speeds"`
	Abilities         map[string]MonsterAbility `json:"abilities"`
	Senses            MonsterSenses             `json:"senses"`
	Languages         []string                  `json:"languages"`
	Gear              []string                  `json:"gear"`
	CR                string                    `json:"cr"`
	XP                int                       `json:"xp"`
	ProficiencyBonus  int                       `json:"proficiency_bonus"`
	Traits            []MonsterMechanic         `json:"traits"`
	Actions           []MonsterAction           `json:"actions"`
	RequiredMechanics []string                  `json:"required_mechanics"`
}

// MonsterHP retains both the fixed HP and its source dice expression.
type MonsterHP struct {
	Average  int `json:"average"`
	Dice     int `json:"dice"`
	Sides    int `json:"sides"`
	Modifier int `json:"modifier"`
}

// MonsterInitiative retains the roll modifier and fixed initiative score.
type MonsterInitiative struct {
	Modifier int `json:"modifier"`
	Score    int `json:"score"`
}

// MonsterAbility records the score, ability modifier and saving throw modifier.
type MonsterAbility struct {
	Score    int `json:"score"`
	Modifier int `json:"modifier"`
	Save     int `json:"save"`
}

// MonsterSenses retains passive perception and any special sense ranges in feet.
type MonsterSenses struct {
	PassivePerception int            `json:"passive_perception"`
	Special           map[string]int `json:"special"`
}

// MonsterMechanic retains a trait or effect and its required implementation key.
// Text is source content, never an executable instruction or inferred capability.
type MonsterMechanic struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Text string `json:"text"`
}

// MonsterAction has a stable local ID and retains all effects. Attack is present
// only for ordinary weapon attacks; other mechanics remain keyed source text.
type MonsterAction struct {
	ID      string            `json:"id"`
	Key     string            `json:"key"`
	Name    string            `json:"name"`
	Text    string            `json:"text"`
	Attack  *MonsterAttack    `json:"attack,omitempty"`
	Effects []MonsterMechanic `json:"effects"`
}

// MonsterAttack describes the source weapon attack, without resolving it.
type MonsterAttack struct {
	Kind       string    `json:"kind"`
	Bonus      int       `json:"bonus"`
	Reach      int       `json:"reach"`
	Range      int       `json:"range"`
	LongRange  int       `json:"long_range"`
	Damage     MonsterHP `json:"damage"`
	DamageType string    `json:"damage_type"`
}

// decodeMonsterNumbers distinguishes an explicit zero from absent or null
// mechanical data. Alias destinations retain strict unknown-field decoding.
func decodeMonsterNumbers(data []byte, dst any, required ...string) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for _, key := range required {
		if value, ok := fields[key]; !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("explicit monster numeric field %s required", key)
		}
	}
	for _, key := range []string{"speeds", "special"} {
		if value, ok := fields[key]; ok {
			var ranges map[string]*int
			if err := json.Unmarshal(value, &ranges); err != nil {
				return err
			}
			for mode, feet := range ranges {
				if feet == nil {
					return fmt.Errorf("explicit %s range for %s required", key, mode)
				}
			}
		}
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	return d.Decode(dst)
}

// UnmarshalJSON requires all numeric block mechanics, including zero XP.
func (m *Monster) UnmarshalJSON(data []byte) error {
	type plain Monster
	return decodeMonsterNumbers(data, (*plain)(m), "ac", "xp", "proficiency_bonus")
}

// UnmarshalJSON requires an explicit fixed value, dice and modifier.
func (h *MonsterHP) UnmarshalJSON(data []byte) error {
	type plain MonsterHP
	return decodeMonsterNumbers(data, (*plain)(h), "average", "dice", "sides", "modifier")
}

// UnmarshalJSON requires explicit initiative modifier and score.
func (i *MonsterInitiative) UnmarshalJSON(data []byte) error {
	type plain MonsterInitiative
	return decodeMonsterNumbers(data, (*plain)(i), "modifier", "score")
}

// UnmarshalJSON requires explicit score, modifier and saving throw.
func (a *MonsterAbility) UnmarshalJSON(data []byte) error {
	type plain MonsterAbility
	return decodeMonsterNumbers(data, (*plain)(a), "score", "modifier", "save")
}

// UnmarshalJSON requires an explicit passive perception value.
func (s *MonsterSenses) UnmarshalJSON(data []byte) error {
	type plain MonsterSenses
	return decodeMonsterNumbers(data, (*plain)(s), "passive_perception")
}

// UnmarshalJSON requires bonus and all range fields, including explicit zeros.
func (a *MonsterAttack) UnmarshalJSON(data []byte) error {
	type plain MonsterAttack
	return decodeMonsterNumbers(data, (*plain)(a), "bonus", "reach", "range", "long_range")
}

// LoadMonsters returns a fresh validated catalogue. Unknown JSON fields are
// rejected so future traits cannot silently disappear during decoding.
func LoadMonsters() (MonsterCatalogue, error) {
	return DecodeMonsters(bytes.NewReader(monsterJSON))
}

// DecodeMonsters strictly decodes one catalogue and validates its source and data.
func DecodeMonsters(r io.Reader) (MonsterCatalogue, error) {
	var c MonsterCatalogue
	data, err := io.ReadAll(r)
	if err != nil {
		return c, err
	}
	if err := uniqueItemJSON(json.NewDecoder(bytes.NewReader(data))); err != nil {
		return c, err
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(&c); err != nil {
		return MonsterCatalogue{}, fmt.Errorf("decode monsters: %w", err)
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return MonsterCatalogue{}, fmt.Errorf("monster catalogue must contain one JSON object")
	}
	if err := c.Validate(); err != nil {
		return MonsterCatalogue{}, err
	}
	return c, nil
}

// Validate checks source provenance and structural integrity. It deliberately
// accepts unknown mechanic keys as catalogue data, while Available rejects them.
func (c MonsterCatalogue) Validate() error {
	if c.SchemaVersion != 1 {
		return fmt.Errorf("unsupported monster schema %d", c.SchemaVersion)
	}
	if err := validateSource(c.Source); err != nil {
		return err
	}
	if len(c.Monsters) == 0 {
		return fmt.Errorf("monster catalogue is empty")
	}
	seen := map[string]bool{}
	for _, m := range c.Monsters {
		if seen[m.ID] {
			return fmt.Errorf("duplicate monster ID %q", m.ID)
		}
		seen[m.ID] = true
		if err := m.Validate(); err != nil {
			return fmt.Errorf("%s: %w", m.ID, err)
		}
	}
	return nil
}

// Validate checks a single block without claiming that its rules are implemented.
func (m Monster) Validate() error {
	if !strings.HasPrefix(m.ID, "srd-5.2.1:content/monsters/") || !contentID.MatchString(m.ID) || strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("invalid monster ID or name")
	}
	if len(m.SourcePages) != 2 || m.SourcePages[0] < 1 || m.SourcePages[0] > m.SourcePages[1] || m.SourcePages[1] > 364 {
		return fmt.Errorf("invalid source pages")
	}
	if !uniqueMonsterStrings(m.Sizes) || len(m.Sizes) == 0 {
		return fmt.Errorf("sizes required")
	}
	for _, size := range m.Sizes {
		if !slices.Contains([]string{"tiny", "small", "medium", "large", "huge", "gargantuan"}, size) {
			return fmt.Errorf("invalid size")
		}
	}
	if !slices.Contains(strings.Fields("aberration beast celestial construct dragon elemental fey fiend giant humanoid monstrosity ooze plant undead"), m.Type) || strings.TrimSpace(m.Alignment) == "" {
		return fmt.Errorf("invalid type or alignment")
	}
	if m.AC <= 0 || !validMonsterDice(m.HP) || m.Initiative.Score != 10+m.Initiative.Modifier {
		return fmt.Errorf("invalid AC, HP or initiative")
	}
	if len(m.Speeds) == 0 {
		return fmt.Errorf("speeds required")
	}
	for kind, feet := range m.Speeds {
		if !slices.Contains(strings.Fields("walk burrow climb fly swim"), kind) || feet < 0 {
			return fmt.Errorf("invalid speed")
		}
	}
	if len(m.Abilities) != 6 {
		return fmt.Errorf("six abilities required")
	}
	for _, key := range strings.Fields("str dex con int wis cha") {
		a, ok := m.Abilities[key]
		if !ok || a.Score < 1 || a.Score > 30 || a.Modifier != a.Score/2-5 || a.Save < -5 || a.Save > 30 {
			return fmt.Errorf("invalid ability %s", key)
		}
	}
	if m.Senses.PassivePerception < 1 || m.Senses.Special == nil {
		return fmt.Errorf("invalid senses")
	}
	for sense, feet := range m.Senses.Special {
		if !familyName.MatchString(sense) || feet <= 0 {
			return fmt.Errorf("invalid special sense")
		}
	}
	if !uniqueMonsterStrings(m.Languages) || !uniqueMonsterStrings(m.Gear) || m.Traits == nil || len(m.Actions) == 0 {
		return fmt.Errorf("explicit languages, gear, traits and actions required")
	}
	if !slices.Contains(strings.Fields("0 1/8 1/4 1/2 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30"), m.CR) || m.XP < 0 || m.ProficiencyBonus < 2 || m.ProficiencyBonus > 9 {
		return fmt.Errorf("invalid CR, XP or proficiency")
	}
	if !uniqueMonsterStrings(m.RequiredMechanics) || !slices.Contains(m.RequiredMechanics, "monster/stat-block") {
		return fmt.Errorf("required mechanics missing or duplicated")
	}
	for _, key := range m.RequiredMechanics {
		if !validMonsterKey(key) {
			return fmt.Errorf("invalid required mechanic %q", key)
		}
	}
	check := func(key, name, text string) error {
		if !validMonsterKey(key) || !slices.Contains(m.RequiredMechanics, key) || strings.TrimSpace(name) == "" || strings.TrimSpace(text) == "" {
			return fmt.Errorf("missing mechanic dependency or source text: %q", key)
		}
		return nil
	}
	for _, trait := range m.Traits {
		if err := check(trait.Key, trait.Name, trait.Text); err != nil {
			return err
		}
	}
	actions := map[string]bool{}
	actionNames := map[string]bool{}
	for _, a := range m.Actions {
		if !familyName.MatchString(a.ID) || actions[a.ID] || actionNames[strings.ToLower(a.Name)] {
			return fmt.Errorf("invalid or duplicate action %q", a.ID)
		}
		actions[a.ID] = true
		actionNames[strings.ToLower(a.Name)] = true
		if err := check(a.Key, a.Name, a.Text); err != nil {
			return err
		}
		if a.Effects == nil {
			return fmt.Errorf("explicit action effects required")
		}
		for _, effect := range a.Effects {
			if err := check(effect.Key, effect.Name, effect.Text); err != nil {
				return err
			}
		}
		if (a.Key == "action/melee-attack" || a.Key == "action/ranged-attack") && a.Attack == nil {
			return fmt.Errorf("weapon action requires typed attack")
		}
		if a.Attack != nil {
			attack := a.Attack
			if a.Key != "action/"+attack.Kind+"-attack" || !slices.Contains(m.RequiredMechanics, "effect/"+attack.DamageType+"-damage") {
				return fmt.Errorf("attack requires matching action and damage mechanic keys")
			}
			hasDamage := false
			for _, effect := range a.Effects {
				if effect.Key == "effect/"+attack.DamageType+"-damage" {
					hasDamage = true
				}
			}
			if !hasDamage {
				return fmt.Errorf("attack damage effect must be retained")
			}
			if !validMonsterDice(attack.Damage) || !slices.Contains(strings.Fields("acid bludgeoning cold fire force lightning necrotic piercing poison psychic radiant slashing thunder"), attack.DamageType) {
				return fmt.Errorf("invalid attack damage")
			}
			if (attack.Kind != "melee" && attack.Kind != "ranged") || (attack.Kind == "melee" && (attack.Reach <= 0 || attack.Range != 0 || attack.LongRange != 0)) || (attack.Kind == "ranged" && (attack.Reach != 0 || attack.Range <= 0 || attack.LongRange < attack.Range)) {
				return fmt.Errorf("invalid attack range")
			}
		}
	}
	return nil
}

func validMonsterDice(d MonsterHP) bool {
	return d.Dice > 0 && slices.Contains([]int{4, 6, 8, 10, 12, 20}, d.Sides) && d.Average > 0 && d.Average == d.Dice*(d.Sides+1)/2+d.Modifier
}

func uniqueMonsterStrings(values []string) bool {
	if values == nil {
		return false
	}
	seen := map[string]bool{}
	for _, v := range values {
		if strings.TrimSpace(v) == "" || seen[v] {
			return false
		}
		seen[v] = true
	}
	return true
}

func validMonsterKey(key string) bool {
	parts := strings.Split(key, "/")
	return len(parts) == 2 && familyName.MatchString(parts[0]) && familyName.MatchString(parts[1])
}

// Available reports whole-block runtime support. No monster mechanics are yet
// qualified; Bandit execution belongs to #148. Catalogue presence, names and NPC
// properties cannot grant support. Future keys must be added with runtime tests.
func (m Monster) Available() bool {
	return m.availableWith(nil)
}

func (m Monster) availableWith(implemented map[string]bool) bool {
	if m.Validate() != nil {
		return false
	}
	for _, key := range m.RequiredMechanics {
		if !implemented[key] {
			return false
		}
	}
	return true
}
