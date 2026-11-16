package srd521

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestItemCatalogueSourceAndVariants(t *testing.T) {
	c, err := LoadItemCatalogue()
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Families) != 1 || len(c.Families[0].Variants) != 4 {
		t.Fatal("Healing must be the sole family with all four source variants")
	}
	// Join both source presentations to the existing source inventory, not names
	// inferred from runtime item records or from a third-party catalogue.
	data, err := os.ReadFile("../../../docs/investigations/issue-80-spell-item-inventory.json")
	if err != nil {
		t.Fatal(err)
	}
	var inventory struct {
		Entries []struct {
			ID   string `json:"id"`
			Page int    `json:"source_page"`
		} `json:"records"`
	}
	if err := json.Unmarshal(data, &inventory); err != nil {
		t.Fatal(err)
	}
	for _, r := range c.Families[0].Presentations {
		found := false
		for _, e := range inventory.Entries {
			if e.ID == r.ID && e.Page == r.Pages[0] {
				found = true
			}
		}
		if !found {
			t.Fatalf("source presentation absent from inventory: %s", r.ID)
		}
	}
	c.Families[0].Variants[0].Healing.Bonus = 999
	fresh, err := LoadItemCatalogue()
	if err != nil || fresh.Families[0].Variants[0].Healing.Bonus != 2 {
		t.Fatal("load must return independent data")
	}
}

func TestItemCatalogueRejectsInvalidFixtures(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*ItemCatalogue)
	}{
		{"schema", func(c *ItemCatalogue) { c.SchemaVersion = 2 }},
		{"rules version", func(c *ItemCatalogue) { c.RulesVersion = "legacy" }},
		{"source hash", func(c *ItemCatalogue) { c.Source.SHA256 = "wrong" }},
		{"source attribution", func(c *ItemCatalogue) { c.Source.Attribution = "" }},
		{"adaptation notice", func(c *ItemCatalogue) { c.Source.Changes = "" }},
		{"source license", func(c *ItemCatalogue) { c.Source.License = "MIT" }},
		{"source date", func(c *ItemCatalogue) { c.Source.RetrievedAt = "yesterday" }},
		{"source page", func(c *ItemCatalogue) { c.Families[0].Presentations[0].Pages = []int{0, 365} }},
		{"missing source", func(c *ItemCatalogue) { c.Families[0].Presentations = c.Families[0].Presentations[:1] }},
		{"duplicate source", func(c *ItemCatalogue) { c.Families[0].Presentations[1] = c.Families[0].Presentations[0] }},
		{"unknown family", func(c *ItemCatalogue) { c.Families[0].ID = "srd-5.2.1:content/items/unknown" }},
		{"duplicate family", func(c *ItemCatalogue) { c.Families = append(c.Families, c.Families[0]) }},
		{"missing variant", func(c *ItemCatalogue) { c.Families[0].Variants = c.Families[0].Variants[:3] }},
		{"missing variant id", func(c *ItemCatalogue) { c.Families[0].Variants[0].ID = "" }},
		{"duplicate variant id", func(c *ItemCatalogue) { c.Families[0].Variants[1].ID = c.Families[0].Variants[0].ID }},
		{"unknown variant", func(c *ItemCatalogue) { c.Families[0].Variants[0].ID += "-unknown" }},
		{"rarity", func(c *ItemCatalogue) { c.Families[0].Variants[0].Rarity = "legendary" }},
		{"missing attunement", func(c *ItemCatalogue) { c.Families[0].Variants[0].Attunement = nil }},
		{"unsupported attunement", func(c *ItemCatalogue) { c.Families[0].Variants[0].Attunement.Mode = "required" }},
		{"missing charges", func(c *ItemCatalogue) { c.Families[0].Variants[0].Charges = nil }},
		{"negative charges", func(c *ItemCatalogue) { c.Families[0].Variants[0].Charges.Maximum = -1 }},
		{"invented charges", func(c *ItemCatalogue) { c.Families[0].Variants[0].Charges.Maximum = 3 }},
		{"charge cost", func(c *ItemCatalogue) { c.Families[0].Variants[0].Charges.PerUse = -1 }},
		{"recharge", func(c *ItemCatalogue) { c.Families[0].Variants[0].Charges.Recharge = "dawn" }},
		{"equip grant", func(c *ItemCatalogue) { c.Families[0].Variants[0].Equip.Mode = "armor" }},
		{"unknown restriction", func(c *ItemCatalogue) { c.Families[0].Variants[0].Use.Restrictions[0] = "free-action" }},
		{"composite restriction", func(c *ItemCatalogue) { c.Families[0].Variants[0].Use.Restrictions[0] = "bonus-action|consume-one" }},
		{"missing restriction", func(c *ItemCatalogue) { c.Families[0].Variants[0].Use.Restrictions = nil }},
		{"missing mechanic", func(c *ItemCatalogue) { c.Families[0].Variants[0].RequiredMechanicKeys = nil }},
		{"unknown mechanic", func(c *ItemCatalogue) { c.Families[0].Variants[0].RequiredMechanicKeys[0] = "item.grant-ac" }},
		{"duplicate mechanic", func(c *ItemCatalogue) {
			c.Families[0].Variants[0].RequiredMechanicKeys[0] = c.Families[0].Variants[0].RequiredMechanicKeys[1]
		}},
		{"negative cost", func(c *ItemCatalogue) { c.Families[0].Variants[0].Cost.Amount = -1 }},
		{"cost currency", func(c *ItemCatalogue) { c.Families[0].Variants[0].Cost.Currency = "usd" }},
		{"cost page", func(c *ItemCatalogue) { c.Families[0].Variants[0].Cost.SourcePages = []int{236, 236} }},
		{"invented greater price", func(c *ItemCatalogue) { c.Families[0].Variants[1].Cost.Amount = 200 }},
		{"invented healing", func(c *ItemCatalogue) { c.Families[0].Variants[0].Healing.Bonus = 20 }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, err := LoadItemCatalogue()
			if err != nil {
				t.Fatal(err)
			}
			tc.mutate(&c)
			data, err := json.Marshal(c)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := DecodeItemCatalogue(data); err == nil {
				t.Fatal("invalid fixture accepted")
			}
		})
	}
}

func TestItemCatalogueRejectsUnsupportedJSON(t *testing.T) {
	for _, field := range []string{`"quantity":10,`, `"ac":99,`, `"actions":["attack"],`, `"resources":{"slots":99},`, `"playable":true,`, `"properties":{"healing":999},`, `"id":"hidden-duplicate",`, `"ID":"hidden-alias",`} {
		t.Run(field, func(t *testing.T) {
			data := bytes.Replace(itemFamiliesJSON, []byte(`"id": "srd-5.2.1:content/items/potion-of-healing/variant/basic"`), []byte(field+`"id": "srd-5.2.1:content/items/potion-of-healing/variant/basic"`), 1)
			if _, err := DecodeItemCatalogue(data); err == nil {
				t.Fatal("unsupported runtime grant/property accepted")
			}
		})
	}
	for _, data := range [][]byte{append(append([]byte{}, itemFamiliesJSON...), []byte(` {}`)...), []byte(`null`), []byte(`{"schema_version":1,"schema_version":1}`)} {
		if _, err := DecodeItemCatalogue(data); err == nil {
			t.Fatal("ambiguous JSON accepted")
		}
	}
}
