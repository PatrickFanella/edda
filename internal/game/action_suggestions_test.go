package game

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"strings"
	"testing"
)

func TestSuggestionsKnownCluesAndUsableSkills(t *testing.T) {
	c, pc := previewFixture(t)
	clue := KnownClue{ID: uuid.New(), Text: "A blue seal marks the gate."}
	got, err := SuggestCharacterActions(context.Background(), c, pc, []KnownClue{clue}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Candidates) != 4 || !strings.HasPrefix(got.Version, "suggestions-v1:") {
		t.Fatalf("suggestions=%+v", got)
	}
	if !strings.Contains(got.Candidates[0].Text, clue.Text) || !strings.Contains(got.Candidates[0].Reason, "uncertain") {
		t.Fatalf("clue=%+v", got.Candidates[0])
	}
	before, _ := json.Marshal(got)
	again, err := SuggestCharacterActions(context.Background(), c, pc, []KnownClue{clue}, nil)
	after, _ := json.Marshal(again)
	if err != nil || string(before) != string(after) {
		t.Fatal("snapshot not deterministic")
	}
	pc.HP--
	changed, err := SuggestCharacterActions(context.Background(), c, pc, []KnownClue{clue}, nil)
	if err != nil || changed.Version == got.Version {
		t.Fatal("actor update did not change version")
	}
	for _, candidate := range got.Candidates[1:] {
		if candidate.Intent == nil || candidate.Intent.Kind != "ability-check" || len(candidate.SourceIDs) == 0 {
			t.Fatalf("unsupported capability=%+v", candidate)
		}
	}
}
func TestSuggestionsUnavailableAndCancellation(t *testing.T) {
	for _, state := range []string{"unconscious", "invalid-build", "paused"} {
		t.Run(state, func(t *testing.T) {
			c, pc := previewFixture(t)
			switch state {
			case "unconscious":
				pc.HP = 0
			case "invalid-build":
				pc.Stats = []byte(`{}`)
			case "paused":
				c.Status = "paused"
			}
			got, err := SuggestCharacterActions(context.Background(), c, pc, nil, nil)
			if err != nil || len(got.Candidates) != 0 {
				t.Fatalf("unavailable=%+v %v", got, err)
			}
		})
	}
	c, pc := previewFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := SuggestCharacterActions(ctx, c, pc, nil, nil); err == nil {
		t.Fatal("cancellation ignored")
	}
}
