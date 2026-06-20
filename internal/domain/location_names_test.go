package domain

import "testing"

func TestCanonicalLocationName(t *testing.T) {
	tests := map[string]string{
		"Lower Tiers Corridor":     "lower tier corridor",
		"Lower Tier Corridor":      "lower tier corridor",
		"  The Core Reactor  ":     "core reactor",
		"core-reactor":             "core reactor",
		"The   Forgotten   Vaults": "forgotten vault",
		"Forgotten Vault":          "forgotten vault",
		"Abyssal Tiers":            "abyssal tier",
	}

	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			if got := CanonicalLocationName(input); got != want {
				t.Fatalf("CanonicalLocationName(%q) = %q, want %q", input, got, want)
			}
		})
	}
}

func TestSameCanonicalLocationName(t *testing.T) {
	if !SameCanonicalLocationName("Lower Tiers Corridor", "Lower Tier Corridor") {
		t.Fatal("expected obvious singular/plural variant to match")
	}
	if SameCanonicalLocationName("Core Archive", "Core Reactor") {
		t.Fatal("expected distinct core locations not to match")
	}
}
