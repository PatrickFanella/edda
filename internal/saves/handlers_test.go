package saves

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSaveAndTimeGoldenShapes(t *testing.T) {
	t.Parallel()

	campaignID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	when := time.Date(2026, 6, 19, 8, 15, 0, 0, time.UTC)
	got := map[string]any{
		"save_point":    savePointToJSON(SavePoint{ID: uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"), CampaignID: campaignID, Name: "Manual save", TurnNumber: 12, IsAuto: false, CreatedAt: when}),
		"campaign_time": campaignTimeJSON{Day: 3, Hour: 8, Minute: 15},
	}
	want, err := os.ReadFile(filepath.Join("testdata", "save_shapes.golden.json"))
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	gotJSON, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	gotJSON = append(gotJSON, '\n')
	if string(gotJSON) != string(want) {
		t.Fatalf("golden mismatch\nwant:\n%s\ngot:\n%s", string(want), string(gotJSON))
	}
}
