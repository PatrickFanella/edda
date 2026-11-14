package llm

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func repoSkills(t *testing.T) *SkillLibrary {
	t.Helper()
	lib, err := LoadSkillLibrary(filepath.Join("..", "..", "skills"))
	if err != nil {
		t.Fatal(err)
	}
	return lib
}

func TestSkillLibraryLoadsRepoSkills(t *testing.T) {
	lib := repoSkills(t)
	if _, ok := lib.byName["run-rpg-sessions"]; !ok || len(lib.skills) != 8 {
		t.Fatalf("skills = %v", lib.names())
	}
	body, err := lib.load("run-rpg-sessions")
	if err != nil || strings.Contains(body, "license: MIT") || !strings.Contains(body, "references/base-workflow.md") {
		t.Fatalf("load: err=%v body=%.200q", err, body)
	}
	file, err := lib.readFile("run-rpg-sessions", "references/base-workflow.md")
	if err != nil || len(file) < 100 {
		t.Fatalf("readFile: err=%v len=%d", err, len(file))
	}
	if !strings.Contains(lib.instructions(), "- run-rpg-sessions: ") {
		t.Errorf("instructions do not list the skills: %s", lib.instructions())
	}
}

func TestSkillLibraryRefusesPathsOutsideSkill(t *testing.T) {
	lib := repoSkills(t)
	for _, p := range []string{"../edit-fiction/SKILL.md", "../../go.mod", "/etc/passwd", "references", "SKILL.md.bak"} {
		if out, err := lib.readFile("run-rpg-sessions", p); err == nil {
			t.Errorf("readFile(%q) = %.60q, want an error", p, out)
		}
	}
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, "s"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "s", "SKILL.md"), []byte("---\nname: s\ndescription: d\n---\nbody"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "secret.md"), []byte("secret"), 0o644)
	_ = os.Symlink(filepath.Join(dir, "secret.md"), filepath.Join(dir, "s", "link.md"))
	lib2, err := LoadSkillLibrary(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lib2.readFile("s", "link.md"); err == nil {
		t.Error("readFile followed a symlink out of the skill folder")
	}
}

// scriptedProvider returns its responses in order and records each request.
type scriptedProvider struct {
	responses []*Response
	requests  [][]Message
	tools     [][]Tool
}

func (s *scriptedProvider) Complete(_ context.Context, messages []Message, tools []Tool) (*Response, error) {
	s.requests = append(s.requests, append([]Message{}, messages...))
	s.tools = append(s.tools, tools)
	r := s.responses[0]
	s.responses = s.responses[1:]
	return r, nil
}

func (s *scriptedProvider) Stream(context.Context, []Message, []Tool) (<-chan StreamChunk, error) {
	return nil, nil
}

func TestSkillProviderAnswersSkillCallsAndKeepsThemForTheFollowUp(t *testing.T) {
	inner := &scriptedProvider{responses: []*Response{
		{ToolCalls: []ToolCall{{ID: "c1", Name: skillToolLoad, Arguments: map[string]any{"name": "run-rpg-sessions"}}}},
		{ToolCalls: []ToolCall{
			{ID: "c2", Name: skillToolRead, Arguments: map[string]any{"skill": "run-rpg-sessions", "path": "references/base-workflow.md"}},
			{ID: "c3", Name: "move_player", Arguments: map[string]any{"to": "inn"}},
		}},
		{Content: "You reach the dock.", ToolCalls: []ToolCall{{ID: "c4", Name: "move_player", Arguments: map[string]any{"to": "dock"}}}},
		{Content: "The ferryman nods."},
	}}
	p := NewSkillProvider(inner, repoSkills(t))
	game := []Tool{{Name: "move_player"}}
	history := []Message{{Role: RoleSystem, Content: "GM"}, {Role: RoleUser, Content: "To the dock."}}

	resp, err := p.Complete(context.Background(), history, game)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.ToolCalls) != 1 || resp.ToolCalls[0].ID != "c4" {
		t.Fatalf("engine should see only the final game call, got %+v", resp.ToolCalls)
	}
	if len(inner.tools[0]) != 3 || inner.requests[0][1].Role != RoleSystem || !strings.Contains(inner.requests[0][1].Content, skillToolLoad) {
		t.Fatalf("first request should add the skill tools and instructions after the system prompt")
	}
	third := inner.requests[2]
	var results []string
	for _, m := range third {
		if m.Role == RoleTool {
			results = append(results, m.Content)
		}
	}
	if len(results) != 3 || !strings.HasPrefix(results[0], "Skill run-rpg-sessions") ||
		!strings.Contains(results[1], "references/base-workflow.md") || !strings.Contains(results[2], "not run") {
		t.Fatalf("skill results = %q", results)
	}

	follow := append(append([]Message{}, history...),
		Message{Role: RoleAssistant, Content: resp.Content, ToolCalls: resp.ToolCalls},
		Message{Role: RoleTool, ToolCallID: "c4", Content: `{"ok":true}`})
	if _, err := p.Complete(context.Background(), follow, game); err != nil {
		t.Fatal(err)
	}
	last := inner.requests[3]
	var ids []string
	for _, m := range last {
		for _, tc := range m.ToolCalls {
			ids = append(ids, tc.ID)
		}
	}
	if strings.Join(ids, ",") != "c1,c2,c3,c4" {
		t.Fatalf("follow-up should replay the skill exchange before c4, got calls %v", ids)
	}
}

func TestSkillProviderLastRoundHasNoSkillTools(t *testing.T) {
	load := &Response{ToolCalls: []ToolCall{{ID: "x", Name: skillToolLoad, Arguments: map[string]any{"name": "edit-fiction"}}}}
	inner := &scriptedProvider{responses: []*Response{load, load, load, {Content: "done"}}}
	p := NewSkillProvider(inner, repoSkills(t))
	resp, err := p.Complete(context.Background(), []Message{{Role: RoleUser, Content: "hi"}}, nil)
	if err != nil || resp.Content != "done" {
		t.Fatalf("resp=%+v err=%v", resp, err)
	}
	if n := len(inner.tools[3]); n != 0 {
		t.Fatalf("last round offered %d tools, want none", n)
	}
}
