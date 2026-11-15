package export

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/edda/pkg/api"
)

func TestSelectedChapters(t *testing.T) {
	early := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	later := early.Add(time.Hour)
	chapters := []api.ChapterResponse{
		{ID: "restart", FromTurn: 1, ToTurn: 2, CreatedAt: later, Revisions: []api.ChapterRevisionResponse{{ID: "restart-r1", Status: "completed"}}},
		{ID: "old", FromTurn: 9, ToTurn: 10, CreatedAt: early, Revisions: []api.ChapterRevisionResponse{{ID: "r1", Status: "completed", Content: "Original prose.\n\nKept exactly."}, {ID: "r2", Status: "completed", Content: "Unselected newer prose."}, {ID: "failed", Status: "failed"}, {ID: "pending", Status: "generating"}}},
	}
	before := append([]api.ChapterRevisionResponse(nil), chapters[1].Revisions...)
	for _, tc := range []struct {
		name string
		ids  map[string]bool
		want []string
		ok   bool
	}{
		{"source order after restart", map[string]bool{"restart-r1": true, "r1": true}, []string{"old:r1", "restart:restart-r1"}, true},
		{"specific older revision", map[string]bool{"r1": true}, []string{"old:r1"}, true},
		{"unknown", map[string]bool{"foreign": true}, nil, false},
		{"failed", map[string]bool{"failed": true}, nil, false},
		{"generating", map[string]bool{"pending": true}, nil, false},
		{"partial selection rejected", map[string]bool{"r1": true, "foreign": true}, []string{"old:r1"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := selectedChapters(chapters, tc.ids)
			var ids []string
			for _, c := range got {
				for _, r := range c.Revisions {
					ids = append(ids, c.ID+":"+r.ID)
				}
			}
			if ok != tc.ok || !reflect.DeepEqual(ids, tc.want) {
				t.Fatalf("got %v, %v; want %v, %v", ids, ok, tc.want, tc.ok)
			}
		})
	}
	if !reflect.DeepEqual(chapters[1].Revisions, before) {
		t.Fatal("selection mutated reader revisions")
	}
	selected, _ := selectedChapters(chapters, map[string]bool{"r1": true})
	md := chapterMarkdown(ChapterExport{CampaignID: "fixture", Chapters: selected, Attribution: "Required attribution", Changes: "Change notice"})
	if !strings.Contains(md, before[0].Content) || strings.Contains(md, before[1].Content) || !strings.Contains(md, "Required attribution") || !strings.Contains(md, "Change notice") {
		t.Fatal("markdown changed selected prose or dropped notices")
	}
}
