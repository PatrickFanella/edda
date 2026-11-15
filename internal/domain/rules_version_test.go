package domain

import (
	"errors"
	"testing"
)

func TestRulesVersionCompatibility(t *testing.T) {
	for _, tt := range []struct {
		version       RulesVersion
		normalized    RulesVersion
		valid, legacy bool
	}{
		{"", RulesVersionLegacy, true, true},
		{RulesVersionLegacy, RulesVersionLegacy, true, true},
		{RulesVersionSRD521, RulesVersionSRD521, true, false},
		{"unknown", "", false, false},
	} {
		t.Run(string(tt.version), func(t *testing.T) {
			got, err := NormalizeRulesVersion(tt.version)
			if (err == nil) != tt.valid || got != tt.normalized {
				t.Fatalf("normalize = %q %v", got, err)
			}
			err = RequireLegacyRules(tt.version)
			if tt.legacy && err != nil {
				t.Fatal(err)
			}
			if !tt.legacy && !errors.Is(err, ErrLegacyRulesMutation) {
				t.Fatalf("guard = %v", err)
			}
		})
	}
}
