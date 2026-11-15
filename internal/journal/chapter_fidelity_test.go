package journal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/edda/pkg/api"
	"github.com/google/uuid"
)

const fidelityDir = "../../docs/playtest/issue-87"

type fidelityFixture struct {
	ID          string              `json:"id"`
	Source      api.ChapterResponse `json:"source"`
	ChapterFile string              `json:"chapter_file"`
	Names       []string            `json:"names"`
	Outcomes    []struct {
		TurnNumber  int    `json:"turn_number"`
		SourceText  string `json:"source_text"`
		ChapterText string `json:"chapter_text"`
	} `json:"outcomes"`
	Contradictions []string        `json:"contradictions"`
	HiddenLore     string          `json:"hidden_lore"`
	ToolSecret     string          `json:"tool_secret"`
	Pending        api.ChapterTurn `json:"pending"`
	Later          api.ChapterTurn `json:"later"`
}

func fidelityFixtures(t *testing.T) []fidelityFixture {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(fidelityDir, "campaigns.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []fidelityFixture
	if err := json.Unmarshal(b, &fixtures); err != nil {
		t.Fatal(err)
	}
	if len(fixtures) != 3 {
		t.Fatalf("want three curated campaigns, got %d", len(fixtures))
	}
	return fixtures
}

func (f fidelityFixture) chapter(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(fidelityDir, f.ChapterFile))
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(b))
}

// These literal anchors are a regression oracle for these fixed examples, not
// a semantic grader of arbitrary model prose. Human quality review is separate.
func (f fidelityFixture) checkChapter(content string) error {
	for _, name := range f.Names {
		if !strings.Contains(content, name) {
			return fmt.Errorf("missing name %q", name)
		}
	}
	for _, outcome := range f.Outcomes {
		if !strings.Contains(content, outcome.ChapterText) {
			return fmt.Errorf("missing outcome from turn %d", outcome.TurnNumber)
		}
	}
	for _, forbidden := range append(append([]string{}, f.Contradictions...), f.HiddenLore, f.ToolSecret, f.Pending.Narrative, f.Later.Narrative) {
		if strings.Contains(strings.ToLower(content), strings.ToLower(forbidden)) {
			return fmt.Errorf("forbidden claim %q", forbidden)
		}
	}
	return nil
}

func TestCuratedChapterFixtures(t *testing.T) {
	for _, f := range fidelityFixtures(t) {
		t.Run(f.ID, func(t *testing.T) {
			turns := f.Source.SourceTurns
			if len(turns) == 0 || f.Source.FromTurn != 1 || f.Source.ToTurn != len(turns) || f.Source.KnowledgeBoundary != chapterKnowledgeBoundary {
				t.Fatal("invalid curated source range/boundary")
			}
			for i, turn := range turns {
				if turn.TurnNumber != i+1 || turn.PlayerInput == "" || turn.Narrative == "" {
					t.Fatalf("invalid transcript turn %+v", turn)
				}
				for _, id := range []string{turn.LogID, turn.AttemptID} {
					if _, err := uuid.Parse(id); err != nil {
						t.Fatal(err)
					}
				}
			}
			source, hash, err := chapterSource(turns, f.Source.RulesVersion, f.Source.RulesMode)
			if err != nil {
				t.Fatal(err)
			}
			var decoded []api.ChapterTurn
			if err := json.Unmarshal(source, &decoded); err != nil || !reflect.DeepEqual(decoded, turns) {
				t.Fatalf("source did not retain exact original transcript: %v", err)
			}
			for _, hidden := range []string{f.HiddenLore, f.ToolSecret, f.Pending.Narrative, f.Later.Narrative} {
				if hidden == "" || strings.Contains(string(source), hidden) {
					t.Fatal("invalid isolation sentinel")
				}
			}
			for _, name := range f.Names {
				if !strings.Contains(string(source), name) {
					t.Fatalf("name absent from original transcript: %s", name)
				}
			}
			for _, outcome := range f.Outcomes {
				if outcome.TurnNumber < 1 || outcome.TurnNumber > len(turns) || outcome.SourceText == "" || outcome.ChapterText == "" || !strings.Contains(turns[outcome.TurnNumber-1].Narrative, outcome.SourceText) {
					t.Fatalf("outcome has no cited source: %+v", outcome)
				}
			}
			content := f.chapter(t)
			if err := f.checkChapter(content); err != nil {
				t.Fatal(err)
			}
			// Show that the oracle detects a removed name, erased outcome,
			// reversed result, and added private knowledge in each example.
			bad := map[string]string{
				"name":        strings.ReplaceAll(content, f.Names[0], "Someone"),
				"outcome":     strings.ReplaceAll(content, f.Outcomes[0].ChapterText, "Something happened"),
				"hidden lore": content + f.HiddenLore,
				"tool":        content + f.ToolSecret,
				"pending":     content + f.Pending.Narrative,
				"later":       content + f.Later.Narrative,
			}
			for _, contradiction := range f.Contradictions {
				bad[contradiction] = content + contradiction
			}
			for name, prose := range bad {
				t.Run("reject "+name, func(t *testing.T) {
					if err := f.checkChapter(prose); err == nil {
						t.Fatal("negative control accepted")
					}
				})
			}
			changed := append([]api.ChapterTurn(nil), turns...)
			changed[0].Narrative = "A different outcome."
			_, changedHash, err := chapterSource(changed, f.Source.RulesVersion, f.Source.RulesMode)
			if err != nil || hash == changedHash {
				t.Fatal("changed outcome did not change source identity")
			}
		})
	}
}
