package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestUserValidate(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr bool
	}{
		{"valid", User{Name: "Player"}, false},
		{"empty name", User{Name: ""}, true},
		{"whitespace name", User{Name: "   "}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.user.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCampaignValidate(t *testing.T) {
	valid := Campaign{Name: "Test", Status: CampaignStatusActive, CreatedBy: uuid.New()}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid campaign: %v", err)
	}

	tests := []struct {
		name string
		mod  func(*Campaign)
	}{
		{"empty name", func(c *Campaign) { c.Name = "" }},
		{"bad status", func(c *Campaign) { c.Status = "invalid" }},
		{"nil creator", func(c *Campaign) { c.CreatedBy = uuid.Nil }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := valid
			tt.mod(&c)
			if err := c.Validate(); err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestLocationValidate(t *testing.T) {
	valid := Location{Name: "Tavern", CampaignID: uuid.New()}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid location: %v", err)
	}
	invalid := Location{Name: "", CampaignID: uuid.New()}
	if err := invalid.Validate(); err == nil {
		t.Error("expected error for empty name")
	}
	invalid2 := Location{Name: "X", CampaignID: uuid.Nil}
	if err := invalid2.Validate(); err == nil {
		t.Error("expected error for nil campaign")
	}
}

func TestNPCValidate(t *testing.T) {
	valid := NPC{Name: "Grim", CampaignID: uuid.New(), Disposition: 0}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid npc: %v", err)
	}

	tests := []struct {
		name string
		mod  func(*NPC)
	}{
		{"empty name", func(n *NPC) { n.Name = "" }},
		{"nil campaign", func(n *NPC) { n.CampaignID = uuid.Nil }},
		{"disposition too low", func(n *NPC) { n.Disposition = -101 }},
		{"disposition too high", func(n *NPC) { n.Disposition = 101 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := valid
			tt.mod(&n)
			if err := n.Validate(); err == nil {
				t.Error("expected error")
			}
		})
	}

	// Boundary values should pass
	edge := valid
	edge.Disposition = -100
	if err := edge.Validate(); err != nil {
		t.Errorf("-100 should be valid: %v", err)
	}
	edge.Disposition = 100
	if err := edge.Validate(); err != nil {
		t.Errorf("100 should be valid: %v", err)
	}
}

func TestFactionValidate(t *testing.T) {
	valid := Faction{Name: "Guild", CampaignID: uuid.New()}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid faction: %v", err)
	}
	if err := (&Faction{Name: "", CampaignID: uuid.New()}).Validate(); err == nil {
		t.Error("expected error for empty name")
	}
}

func TestPlayerCharacterValidate(t *testing.T) {
	valid := PlayerCharacter{
		Name:       "Hero",
		CampaignID: uuid.New(),
		UserID:     uuid.New(),
		Level:      1,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid pc: %v", err)
	}

	tests := []struct {
		name string
		mod  func(*PlayerCharacter)
	}{
		{"empty name", func(pc *PlayerCharacter) { pc.Name = "" }},
		{"nil campaign", func(pc *PlayerCharacter) { pc.CampaignID = uuid.Nil }},
		{"nil user", func(pc *PlayerCharacter) { pc.UserID = uuid.Nil }},
		{"level zero", func(pc *PlayerCharacter) { pc.Level = 0 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := valid
			tt.mod(&pc)
			if err := pc.Validate(); err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestItemValidate(t *testing.T) {
	valid := Item{Name: "Sword", CampaignID: uuid.New(), ItemType: ItemTypeWeapon, Quantity: 1}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid item: %v", err)
	}

	tests := []struct {
		name string
		mod  func(*Item)
	}{
		{"empty name", func(i *Item) { i.Name = "" }},
		{"nil campaign", func(i *Item) { i.CampaignID = uuid.Nil }},
		{"negative quantity", func(i *Item) { i.Quantity = -1 }},
		{"invalid type", func(i *Item) { i.ItemType = "invalid" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := valid
			tt.mod(&i)
			if err := i.Validate(); err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestQuestValidate(t *testing.T) {
	valid := Quest{Title: "Find the key", CampaignID: uuid.New()}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid quest: %v", err)
	}
	if err := (&Quest{Title: "", CampaignID: uuid.New()}).Validate(); err == nil {
		t.Error("expected error for empty title")
	}
}

func TestWorldFactValidate(t *testing.T) {
	valid := WorldFact{Fact: "Dragons exist", CampaignID: uuid.New()}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid fact: %v", err)
	}
	if err := (&WorldFact{Fact: "", CampaignID: uuid.New()}).Validate(); err == nil {
		t.Error("expected error for empty fact")
	}
}

func TestSessionLogValidate(t *testing.T) {
	valid := SessionLog{CampaignID: uuid.New(), TurnNumber: 1, PlayerInput: "look around"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid session log: %v", err)
	}

	tests := []struct {
		name string
		mod  func(*SessionLog)
	}{
		{"nil campaign", func(sl *SessionLog) { sl.CampaignID = uuid.Nil }},
		{"zero turn", func(sl *SessionLog) { sl.TurnNumber = 0 }},
		{"empty input", func(sl *SessionLog) { sl.PlayerInput = "" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sl := valid
			tt.mod(&sl)
			if err := sl.Validate(); err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestMemoryValidate(t *testing.T) {
	valid := Memory{CampaignID: uuid.New(), Content: "The hero arrived at the tavern"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid memory: %v", err)
	}
	if err := (&Memory{CampaignID: uuid.Nil, Content: "x"}).Validate(); err == nil {
		t.Error("expected error for nil campaign")
	}
	if err := (&Memory{CampaignID: uuid.New(), Content: ""}).Validate(); err == nil {
		t.Error("expected error for empty content")
	}
}

func TestEntityRelationshipValidate(t *testing.T) {
	valid := EntityRelationship{
		CampaignID:       uuid.New(),
		SourceEntityID:   uuid.New(),
		TargetEntityID:   uuid.New(),
		RelationshipType: "allied",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid relationship: %v", err)
	}

	tests := []struct {
		name string
		mod  func(*EntityRelationship)
	}{
		{"nil campaign", func(er *EntityRelationship) { er.CampaignID = uuid.Nil }},
		{"nil source", func(er *EntityRelationship) { er.SourceEntityID = uuid.Nil }},
		{"nil target", func(er *EntityRelationship) { er.TargetEntityID = uuid.Nil }},
		{"empty type", func(er *EntityRelationship) { er.RelationshipType = "" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			er := valid
			tt.mod(&er)
			if err := er.Validate(); err == nil {
				t.Error("expected error")
			}
		})
	}
}
