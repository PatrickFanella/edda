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

//go:embed testdata/item-families.json
var itemFamiliesJSON []byte

// ItemCatalogue is descriptive source data, never inventory or executable rules.
// SchemaVersion versions this contract independently of the source RulesVersion.
type ItemCatalogue struct {
	SchemaVersion int          `json:"schema_version"`
	RulesVersion  string       `json:"rules_version"`
	Source        Source       `json:"source"`
	Families      []ItemFamily `json:"families"`
}

// ItemReference retains each source presentation, including mundane/magic aliases.
type ItemReference struct {
	ID    string `json:"id"`
	Pages []int  `json:"pages"`
}

// ItemFamily nests every source-listed variant under one canonical identity.
// Presentations reference planning inventory IDs, not owned item IDs.
type ItemFamily struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Kind          string          `json:"kind"`
	Presentations []ItemReference `json:"presentations"`
	Variants      []ItemVariant   `json:"variants"`
}

// ItemVariant contains no runtime quantity, current charges, AC or action grants.
// Restrictions and required mechanic keys describe requirements, not permissions.
type ItemVariant struct {
	ID                   string           `json:"id"`
	Name                 string           `json:"name"`
	Rarity               string           `json:"rarity"`
	Attunement           *ItemAttunement  `json:"attunement"`
	Charges              *ItemCharges     `json:"charges"`
	Equip                *ItemRestriction `json:"equip"`
	Use                  *ItemRestriction `json:"use"`
	Cost                 *ItemCost        `json:"cost"`
	Healing              *ItemHealing     `json:"healing"`
	RequiredMechanicKeys []string         `json:"required_mechanic_keys"`
}

// ItemAttunement records whether attunement and additional qualifications apply.
type ItemAttunement struct {
	Mode         string   `json:"mode"`
	Restrictions []string `json:"restrictions"`
}

// ItemCharges describes capacity and recharge, never a character resource balance.
// A consumable quantity is not a charge pool; it uses mode none.
type ItemCharges struct {
	Mode     string `json:"mode"`
	Maximum  int    `json:"maximum"`
	PerUse   int    `json:"per_use"`
	Recharge string `json:"recharge"`
}

// ItemRestriction names equip/use modes and explicit source requirement keys.
type ItemRestriction struct {
	Mode         string   `json:"mode"`
	Restrictions []string `json:"restrictions"`
}

// ItemCost distinguishes a listed purchase price from an unspecified price.
// Unspecified does not mean free. No rarity-to-price inference is performed.
type ItemCost struct {
	Kind        string `json:"kind"`
	Amount      int    `json:"amount"`
	Currency    string `json:"currency"`
	SourcePages []int  `json:"source_pages"`
}

// ItemHealing records source dice only; it is not passed to a resolver.
type ItemHealing struct {
	Dice  int `json:"dice"`
	Sides int `json:"sides"`
	Bonus int `json:"bonus"`
}

// LoadItemCatalogue returns a fresh validated copy of the single contract fixture.
// It has no state, inventory, action registry or resolver side effects.
func LoadItemCatalogue() (ItemCatalogue, error) { return DecodeItemCatalogue(itemFamiliesJSON) }

// DecodeItemCatalogue rejects unknown fields, duplicate JSON keys and trailing data.
// Validation is structural/source validation only, never gameplay qualification.
func DecodeItemCatalogue(data []byte) (ItemCatalogue, error) {
	if err := uniqueItemJSON(json.NewDecoder(bytes.NewReader(data))); err != nil {
		return ItemCatalogue{}, err
	}
	var c ItemCatalogue
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(&c); err != nil {
		return ItemCatalogue{}, fmt.Errorf("decode item catalogue: %w", err)
	}
	if err := c.Validate(); err != nil {
		return ItemCatalogue{}, err
	}
	return c, nil
}

// uniqueItemJSON prevents encoding/json's last-key-wins behavior hiding properties.
func uniqueItemJSON(d *json.Decoder) error {
	return uniqueCatalogueJSON(d, false)
}

// uniqueCatalogueJSON checks member names before typed decoding can discard them.
// Existing catalogues retain their null policy; multiclass rejects null values.
func uniqueCatalogueJSON(d *json.Decoder, rejectNull bool) error {
	var walk func() error
	walk = func() error {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		if tok == nil && rejectNull {
			return fmt.Errorf("null catalogue values are not allowed")
		}
		delim, ok := tok.(json.Delim)
		if !ok {
			return nil
		}
		switch delim {
		case '{':
			seen := map[string]bool{}
			for d.More() {
				key, err := d.Token()
				if err != nil {
					return err
				}
				name, ok := key.(string)
				if !ok || seen[name] || name != strings.ToLower(name) {
					return fmt.Errorf("duplicate or invalid catalogue property %v", key)
				}
				seen[name] = true
				if err := walk(); err != nil {
					return err
				}
			}
		case '[':
			for d.More() {
				if err := walk(); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unexpected delimiter %v", delim)
		}
		_, err = d.Token()
		return err
	}
	if err := walk(); err != nil {
		return fmt.Errorf("catalogue JSON: %w", err)
	}
	if _, err := d.Token(); err != io.EOF {
		return fmt.Errorf("catalogue must contain one JSON object")
	}
	return nil
}

func itemPages(p []int) bool { return len(p) == 2 && p[0] > 0 && p[0] <= p[1] && p[1] <= 364 }
func itemKeys(keys []string, allowed string) bool {
	if keys == nil {
		return false
	}
	seen := map[string]bool{}
	for _, k := range keys {
		if seen[k] || !slices.Contains(strings.Split(allowed, "|"), k) || k == "" {
			return false
		}
		seen[k] = true
	}
	return true
}

// Validate fails closed on unreviewed families, omitted variants and unsupported
// properties. The initial reviewed source manifest is deliberately only Healing;
// subsequent data PRs must extend it with independent source/variant evidence.
func (c ItemCatalogue) Validate() error {
	if c.SchemaVersion != 1 || c.RulesVersion != "srd-5.2.1" {
		return fmt.Errorf("unsupported item contract version")
	}
	if err := validateSource(c.Source); err != nil {
		return err
	}
	if len(c.Families) != 1 {
		return fmt.Errorf("initial item contract requires exactly the Healing family")
	}
	f := c.Families[0]
	if f.ID != "srd-5.2.1:content/items/potion-of-healing" || f.Name != "Potion of Healing" || f.Kind != "magic" {
		return fmt.Errorf("unreviewed item family")
	}
	expected := map[string]int{"srd-5.2.1:content/adventuring-gear/potion-of-healing": 99, "srd-5.2.1:content/magic-items/potions-of-healing": 236}
	for _, r := range f.Presentations {
		page, ok := expected[r.ID]
		if !ok || !itemPages(r.Pages) || r.Pages[0] != page || r.Pages[1] != page {
			return fmt.Errorf("invalid or duplicate item source presentation")
		}
		delete(expected, r.ID)
	}
	if len(expected) != 0 {
		return fmt.Errorf("missing item source presentation")
	}
	variants := map[string]string{"basic": "common", "greater": "uncommon", "superior": "rare", "supreme": "very-rare"}
	for _, v := range f.Variants {
		variant := strings.TrimPrefix(v.ID, f.ID+"/variant/")
		rarity, ok := variants[variant]
		if !ok || v.ID != f.ID+"/variant/"+variant || v.Rarity != rarity || strings.TrimSpace(v.Name) == "" {
			return fmt.Errorf("invalid, duplicate or unreviewed variant %q", v.ID)
		}
		delete(variants, variant)
		name := "Potion of Healing"
		if variant != "basic" {
			name += " (" + variant + ")"
		}
		if v.Name != name {
			return fmt.Errorf("variant name disagrees with source")
		}
		if v.Attunement == nil || v.Attunement.Mode != "none" || !itemKeys(v.Attunement.Restrictions, "") {
			return fmt.Errorf("unsupported Healing attunement")
		}
		if v.Charges == nil || v.Charges.Mode != "none" || v.Charges.Maximum != 0 || v.Charges.PerUse != 0 || v.Charges.Recharge != "none" {
			return fmt.Errorf("invalid Healing charges/recharge")
		}
		if v.Equip == nil || v.Equip.Mode != "none" || !itemKeys(v.Equip.Restrictions, "") {
			return fmt.Errorf("unsupported Healing equip restrictions")
		}
		if v.Use == nil || v.Use.Mode != "drink-or-administer" || !itemKeys(v.Use.Restrictions, "bonus-action|consume-one|administer-within-5-feet") || len(v.Use.Restrictions) != 3 {
			return fmt.Errorf("invalid Healing use restrictions")
		}
		if !itemKeys(v.RequiredMechanicKeys, "item.consume|item.bonus-action|item.administer-within-5-feet|healing.roll|healing.cap-at-max-hp") || len(v.RequiredMechanicKeys) != 5 {
			return fmt.Errorf("missing or unknown Healing mechanic key")
		}
		if v.Healing == nil || v.Healing.Sides != 4 {
			return fmt.Errorf("invalid Healing dice")
		}
		values := map[string][2]int{"basic": {2, 2}, "greater": {4, 4}, "superior": {8, 8}, "supreme": {10, 20}}
		if v.Healing.Dice != values[variant][0] || v.Healing.Bonus != values[variant][1] {
			return fmt.Errorf("Healing values disagree with source")
		}
		if v.Cost == nil {
			return fmt.Errorf("explicit cost required")
		}
		cost := v.Cost
		if variant == "basic" {
			if cost.Kind != "listed" || cost.Amount != 50 || cost.Currency != "gp" || !itemPages(cost.SourcePages) || cost.SourcePages[0] != 99 || cost.SourcePages[1] != 99 {
				return fmt.Errorf("invalid listed Healing cost")
			}
		} else if cost.Kind != "unspecified" || cost.Amount != 0 || cost.Currency != "none" || !itemPages(cost.SourcePages) || cost.SourcePages[0] != 236 || cost.SourcePages[1] != 236 {
			return fmt.Errorf("unlisted variant cost must remain unspecified")
		}
	}
	if len(variants) != 0 {
		return fmt.Errorf("missing source-listed Healing variants")
	}
	return nil
}
