package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type stubDescribeSceneStore struct {
	lastLocationID uuid.UUID
	lastDesc       string
	lastMood       *string
	lastTimeOfDay  *string
	err            error
}

func (s *stubDescribeSceneStore) UpdateScene(_ context.Context, locationID uuid.UUID, description string, mood, timeOfDay *string) error {
	if s.err != nil {
		return s.err
	}
	s.lastLocationID = locationID
	s.lastDesc = description
	s.lastMood = mood
	s.lastTimeOfDay = timeOfDay
	return nil
}

func TestRegisterDescribeScene(t *testing.T) {
	reg := NewRegistry()
	if err := RegisterDescribeScene(reg, &stubDescribeSceneStore{}); err != nil {
		t.Fatalf("register describe_scene: %v", err)
	}

	registered := reg.List()
	if len(registered) != 1 {
		t.Fatalf("registered tool count = %d, want 1", len(registered))
	}
	if registered[0].Name != describeSceneToolName {
		t.Fatalf("tool name = %q, want %q", registered[0].Name, describeSceneToolName)
	}
	required, ok := registered[0].Parameters["required"].([]string)
	if !ok {
		t.Fatalf("required schema has unexpected type %T", registered[0].Parameters["required"])
	}
	if len(required) != 1 || required[0] != "description" {
		t.Fatalf("required schema = %#v, want [description]", required)
	}
}

func TestDescribeSceneHandleSuccess(t *testing.T) {
	store := &stubDescribeSceneStore{}
	h := NewDescribeSceneHandler(store)
	locationID := uuid.New()
	ctx := WithCurrentLocationID(context.Background(), locationID)

	got, err := h.Handle(ctx, map[string]any{
		"description": "The forest glade is lit by silver moonlight.",
		"mood":        "mysterious",
		"time_of_day": "night",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if store.lastLocationID != locationID {
		t.Fatalf("store location_id = %s, want %s", store.lastLocationID, locationID)
	}
	if store.lastDesc != "The forest glade is lit by silver moonlight." {
		t.Fatalf("store description = %q, want expected description", store.lastDesc)
	}
	if store.lastMood == nil || *store.lastMood != "mysterious" {
		t.Fatalf("store mood = %v, want mysterious", store.lastMood)
	}
	if store.lastTimeOfDay == nil || *store.lastTimeOfDay != "night" {
		t.Fatalf("store time_of_day = %v, want night", store.lastTimeOfDay)
	}
	if got.Success != true {
		t.Fatalf("result success = %v, want true", got.Success)
	}
	if got.Data["location_id"] != locationID.String() {
		t.Fatalf("result location_id = %v, want %s", got.Data["location_id"], locationID)
	}
}

func TestDescribeSceneHandleMissingRequiredDescription(t *testing.T) {
	store := &stubDescribeSceneStore{}
	h := NewDescribeSceneHandler(store)
	ctx := WithCurrentLocationID(context.Background(), uuid.New())

	_, err := h.Handle(ctx, map[string]any{
		"mood": "tense",
	})
	if err == nil {
		t.Fatal("expected error for missing description")
	}
	if !strings.Contains(err.Error(), "description is required") {
		t.Fatalf("error = %v, want description-required message", err)
	}
}

var _ DescribeSceneStore = (*stubDescribeSceneStore)(nil)
