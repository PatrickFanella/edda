package srd521

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func banditCatalogue(t *testing.T) MonsterCatalogue {
	t.Helper()
	c, err := LoadMonsters()
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestMonsterRequiredNumbersCannotDisappear(t *testing.T) {
	for _, path := range []string{
		"ac", "xp", "proficiency_bonus", "hp.average", "hp.dice", "hp.sides", "hp.modifier",
		"initiative.modifier", "initiative.score", "abilities.dex.score", "abilities.dex.modifier", "abilities.dex.save",
		"senses.passive_perception", "actions.0.attack.bonus", "actions.0.attack.reach", "actions.0.attack.range",
		"actions.0.attack.long_range", "actions.0.attack.damage.modifier",
	} {
		for _, mode := range []string{"omitted", "null"} {
			t.Run(path+"/"+mode, func(t *testing.T) {
				var root map[string]any
				if err := json.Unmarshal(monsterJSON, &root); err != nil {
					t.Fatal(err)
				}
				var value any = root["monsters"].([]any)[0]
				parts := strings.Split(path, ".")
				for _, part := range parts[:len(parts)-1] {
					if part == "0" {
						value = value.([]any)[0]
					} else {
						value = value.(map[string]any)[part]
					}
				}
				object := value.(map[string]any)
				key := parts[len(parts)-1]
				if mode == "omitted" {
					delete(object, key)
				} else {
					object[key] = nil
				}
				data, err := json.Marshal(root)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := DecodeMonsters(bytes.NewReader(data)); err == nil {
					t.Fatal("incomplete numeric mechanics accepted")
				}
			})
		}
	}
	for name, data := range map[string][]byte{
		"null speed":   bytes.Replace(monsterJSON, []byte(`"walk": 30`), []byte(`"walk": null`), 1),
		"duplicate XP": bytes.Replace(monsterJSON, []byte(`"xp": 25`), []byte(`"xp": 99, "xp": 25`), 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeMonsters(bytes.NewReader(data)); err == nil {
				t.Fatal("ambiguous numeric mechanics accepted")
			}
		})
	}
	// Existing Bandit STR/INT/WIS/CHA saves and melee range zeros are explicit
	// valid data; presence checks must not reject those source-backed zeros.
	if _, err := LoadMonsters(); err != nil {
		t.Fatal(err)
	}
}

func TestBanditSourceFixture(t *testing.T) {
	c := banditCatalogue(t)
	if len(c.Monsters) != 1 {
		t.Fatal("Bandit must be the sole initial record")
	}
	m := c.Monsters[0]
	if m.ID != "srd-5.2.1:content/monsters/bandit" || m.Name != "Bandit" || !reflect.DeepEqual(m.SourcePages, []int{261, 261}) {
		t.Fatal("wrong source identity")
	}
	if !reflect.DeepEqual(m.Sizes, []string{"medium", "small"}) || m.Type != "humanoid" || m.Alignment != "Neutral" || m.AC != 12 || m.HP != (MonsterHP{11, 2, 8, 2}) || m.Initiative != (MonsterInitiative{1, 11}) || m.Speeds["walk"] != 30 {
		t.Fatal("source combat highlights differ")
	}
	wantAbilities := map[string]MonsterAbility{"str": {11, 0, 0}, "dex": {12, 1, 1}, "con": {12, 1, 1}, "int": {10, 0, 0}, "wis": {10, 0, 0}, "cha": {10, 0, 0}}
	if !reflect.DeepEqual(m.Abilities, wantAbilities) || m.Senses.PassivePerception != 10 || len(m.Senses.Special) != 0 || !reflect.DeepEqual(m.Languages, []string{"Common", "Thieves’ Cant"}) || !reflect.DeepEqual(m.Gear, []string{"Leather Armor", "Light Crossbow", "Scimitar"}) || m.CR != "1/8" || m.XP != 25 || m.ProficiencyBonus != 2 {
		t.Fatal("source details differ")
	}
	if len(m.Traits) != 0 || len(m.Actions) != 2 {
		t.Fatal("traits/actions differ")
	}
	scimitar, crossbow := m.Actions[0], m.Actions[1]
	if scimitar.ID != "scimitar" || *scimitar.Attack != (MonsterAttack{Kind: "melee", Bonus: 3, Reach: 5, Damage: MonsterHP{4, 1, 6, 1}, DamageType: "slashing"}) {
		t.Fatal("scimitar differs")
	}
	if crossbow.ID != "light-crossbow" || *crossbow.Attack != (MonsterAttack{Kind: "ranged", Bonus: 3, Range: 80, LongRange: 320, Damage: MonsterHP{5, 1, 8, 1}, DamageType: "piercing"}) {
		t.Fatal("crossbow differs")
	}
	if c.Source.Attribution != Attribution || c.Source.SHA256 != artifactSHA256 || c.Source.Changes == "" {
		t.Fatal("provenance lost")
	}
	if m.Available() {
		t.Fatal("catalogue presence cannot grant NPC behavior")
	}
	// Every load owns its maps/slices; no caller can mutate the embedded catalogue.
	m.Abilities["str"] = MonsterAbility{}
	if banditCatalogue(t).Monsters[0].Abilities["str"].Score != 11 {
		t.Fatal("shared mutable catalogue")
	}
}

func TestMonsterMalformedFixtures(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*MonsterCatalogue)
	}{
		{"schema", func(c *MonsterCatalogue) { c.SchemaVersion = 2 }},
		{"pin", func(c *MonsterCatalogue) { c.Source.SHA256 = "bad" }},
		{"version", func(c *MonsterCatalogue) { c.Source.Version = "5.1" }},
		{"attribution", func(c *MonsterCatalogue) { c.Source.Attribution = "" }},
		{"adaptation", func(c *MonsterCatalogue) { c.Source.Changes = " " }},
		{"date", func(c *MonsterCatalogue) { c.Source.RetrievedAt = "yesterday" }},
		{"ID", func(c *MonsterCatalogue) { c.Monsters[0].ID = "srd-5.2.1:content/monsters/Bad ID" }},
		{"wrong family", func(c *MonsterCatalogue) { c.Monsters[0].ID = "srd-5.2.1:content/spells/bandit" }},
		{"duplicate blocks", func(c *MonsterCatalogue) { c.Monsters = append(c.Monsters, c.Monsters[0]) }},
		{"pages", func(c *MonsterCatalogue) { c.Monsters[0].SourcePages = []int{261, 365} }},
		{"size", func(c *MonsterCatalogue) { c.Monsters[0].Sizes = []string{"enormous"} }},
		{"type", func(c *MonsterCatalogue) { c.Monsters[0].Type = "person" }},
		{"AC", func(c *MonsterCatalogue) { c.Monsters[0].AC = 0 }},
		{"HP", func(c *MonsterCatalogue) { c.Monsters[0].HP.Average = 12 }},
		{"initiative", func(c *MonsterCatalogue) { c.Monsters[0].Initiative.Score = 0 }},
		{"speed", func(c *MonsterCatalogue) { c.Monsters[0].Speeds["walk"] = -1 }},
		{"ability missing", func(c *MonsterCatalogue) { delete(c.Monsters[0].Abilities, "str") }},
		{"ability modifier", func(c *MonsterCatalogue) { c.Monsters[0].Abilities["str"] = MonsterAbility{11, 2, 0} }},
		{"save", func(c *MonsterCatalogue) { c.Monsters[0].Abilities["str"] = MonsterAbility{11, 0, -100} }},
		{"senses", func(c *MonsterCatalogue) { c.Monsters[0].Senses.PassivePerception = -1 }},
		{"languages", func(c *MonsterCatalogue) { c.Monsters[0].Languages = nil }},
		{"CR", func(c *MonsterCatalogue) { c.Monsters[0].CR = "1/9" }},
		{"XP", func(c *MonsterCatalogue) { c.Monsters[0].XP = -1 }},
		{"duplicate actions", func(c *MonsterCatalogue) { m := &c.Monsters[0]; m.Actions = append(m.Actions, m.Actions[0]) }},
		{"missing attack", func(c *MonsterCatalogue) { c.Monsters[0].Actions[0].Attack = nil }},
		{"missing damage effect", func(c *MonsterCatalogue) { c.Monsters[0].Actions[0].Effects = []MonsterMechanic{} }},
		{"duplicate action name", func(c *MonsterCatalogue) { c.Monsters[0].Actions[1].Name = c.Monsters[0].Actions[0].Name }},
		{"action ID", func(c *MonsterCatalogue) { c.Monsters[0].Actions[0].ID = "BAD" }},
		{"damage", func(c *MonsterCatalogue) { c.Monsters[0].Actions[0].Attack.Damage.Sides = 7 }},
		{"range", func(c *MonsterCatalogue) { c.Monsters[0].Actions[1].Attack.LongRange = 1 }},
		{"no traits declaration", func(c *MonsterCatalogue) { c.Monsters[0].Traits = nil }},
		{"missing required keys", func(c *MonsterCatalogue) { c.Monsters[0].RequiredMechanics = nil }},
		{"duplicate keys", func(c *MonsterCatalogue) {
			m := &c.Monsters[0]
			m.RequiredMechanics = append(m.RequiredMechanics, m.RequiredMechanics[0])
		}},
		{"unlisted effect", func(c *MonsterCatalogue) { c.Monsters[0].Actions[0].Effects[0].Key = "effect/new-effect" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := banditCatalogue(t)
			tt.mutate(&c)
			if c.Validate() == nil {
				t.Fatal("accepted malformed catalogue")
			}
			data, err := json.Marshal(c)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := DecodeMonsters(bytes.NewReader(data))
			if err == nil || len(decoded.Monsters) != 0 {
				t.Fatal("decoder must return no catalogue on invalid data")
			}
		})
	}
}

func TestMonsterUnknownMechanicsFailWholeBlock(t *testing.T) {
	for _, kind := range []string{"trait", "action", "effect", "required"} {
		t.Run(kind, func(t *testing.T) {
			m := banditCatalogue(t).Monsters[0]
			implemented := map[string]bool{}
			for _, key := range m.RequiredMechanics {
				implemented[key] = true
			}
			if !m.availableWith(implemented) {
				t.Fatal("test capability set must cover original fixture")
			}
			unknown := MonsterMechanic{Key: kind + "/future", Name: "Future mechanic", Text: "Retained unsupported source behavior."}
			m.RequiredMechanics = append(m.RequiredMechanics, unknown.Key)
			switch kind {
			case "trait":
				m.Traits = append(m.Traits, unknown)
			case "action":
				m.Actions = append(m.Actions, MonsterAction{ID: "future", Key: unknown.Key, Name: unknown.Name, Text: unknown.Text, Effects: []MonsterMechanic{}})
			case "effect":
				m.Actions[0].Effects = append(m.Actions[0].Effects, unknown)
			}
			if err := m.Validate(); err != nil {
				t.Fatal("unknown mechanics must remain valid catalogue data:", err)
			}
			before, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			if m.availableWith(implemented) || m.Available() {
				t.Fatal("unsupported mechanic must block entire record")
			}
			after, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, after) {
				t.Fatal("availability dropped unsupported content")
			}
		})
	}
}

func TestMonsterStrictDecoding(t *testing.T) {
	for _, data := range []string{string(monsterJSON) + " {}", strings.Replace(string(monsterJSON), `"traits": []`, `"unmodelled_trait": true, "traits": []`, 1), `{`} {
		if _, err := DecodeMonsters(strings.NewReader(data)); err == nil {
			t.Fatal("accepted unknown field, extra object or malformed JSON")
		}
	}
}
