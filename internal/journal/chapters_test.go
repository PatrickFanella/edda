package journal

import (
	"encoding/json"
	"testing"

	"git.subcult.tv/subculture-collective/edda/pkg/api"
)

func TestChapterSourceIdentity(t *testing.T) {
	base := []api.ChapterTurn{{LogID: "log-a", AttemptID: "receipt-a", TurnNumber: 1, PlayerInput: "Try the gate", Narrative: "The gate remained shut."}}
	source, hash, err := chapterSource(base, "srd-5.2.1", "strict")
	if err != nil {
		t.Fatal(err)
	}
	again, replayHash, err := chapterSource(base, "srd-5.2.1", "strict")
	if err != nil || string(source) != string(again) || hash != replayHash || len(hash) != 64 {
		t.Fatalf("unstable source: %s %s %v", hash, replayHash, err)
	}
	var decoded []api.ChapterTurn
	if err := json.Unmarshal(source, &decoded); err != nil || len(decoded) != 1 || decoded[0] != base[0] {
		t.Fatalf("snapshot changed: %s %v", source, err)
	}
	for _, tc := range []struct {
		name, version, mode string
		change              func(*api.ChapterTurn)
	}{
		{name: "rules version", version: "legacy", mode: "strict"},
		{name: "rules mode", version: "srd-5.2.1", mode: "narrative"},
		{name: "event identity", version: "srd-5.2.1", mode: "strict", change: func(t *api.ChapterTurn) { t.LogID = "log-b" }},
		{name: "receipt identity", version: "srd-5.2.1", mode: "strict", change: func(t *api.ChapterTurn) { t.AttemptID = "receipt-b" }},
		{name: "played outcome", version: "srd-5.2.1", mode: "strict", change: func(t *api.ChapterTurn) { t.Narrative = "The gate opened." }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			turns := append([]api.ChapterTurn(nil), base...)
			if tc.change != nil {
				tc.change(&turns[0])
			}
			_, changed, err := chapterSource(turns, tc.version, tc.mode)
			if err != nil || changed == hash {
				t.Fatalf("identity did not change: %s %v", changed, err)
			}
		})
	}
}
