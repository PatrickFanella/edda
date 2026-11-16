package combat

import (
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	"testing"
)

func TestBanditGeometry(t *testing.T) {
	_, version, err := srd521.BanditRecord()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, action, cover string
		distance            int
		sight, legal        bool
	}{
		{"melee", "scimitar", "none", 5, true, true},
		{"melee far", "scimitar", "none", 10, true, false},
		{"normal ranged", "light-crossbow", "none", 80, true, true},
		{"ranged half cover", "light-crossbow", "half", 30, true, true},
		{"ranged near", "light-crossbow", "none", 5, true, false},
		{"long range", "light-crossbow", "none", 85, true, false},
		{"total cover", "light-crossbow", "total", 30, true, false},
		{"hidden", "scimitar", "none", 5, false, false},
		{"unknown cover", "scimitar", "", 5, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := srd521.BanditProfile{Version: version, ActionID: tc.action}
			s := Space{DistanceFeet: &tc.distance, LineOfSight: &tc.sight, Cover: tc.cover}
			if err := BanditAttack(&p, &s); (err == nil) != tc.legal {
				t.Fatalf("legal=%t err=%v", tc.legal, err)
			}
		})
	}
	if BanditAttack(nil, nil) == nil {
		t.Fatal("missing geometry accepted")
	}
	p := srd521.BanditProfile{Version: version, ActionID: "scimitar"}
	if BanditAttack(&p, &Space{}) == nil {
		t.Fatal("unknown distance accepted")
	}
}
