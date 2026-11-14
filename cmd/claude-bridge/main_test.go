package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fakeClaude writes a script that records its arguments and stdin, then
// prints reply.
func fakeClaude(t *testing.T, reply string) (bin, argsFile, stdinFile string) {
	t.Helper()
	dir := t.TempDir()
	argsFile, stdinFile = filepath.Join(dir, "args"), filepath.Join(dir, "stdin")
	bin = filepath.Join(dir, "claude")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > " + argsFile + "\ncat > " + stdinFile + "\ncat <<'JSON'\n" + reply + "\nJSON\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return bin, argsFile, stdinFile
}

func newTestBridge(t *testing.T, bin string) *bridge {
	return &bridge{token: []byte(strings.Repeat("t", 32)), claude: bin, model: "claude-sonnet-5-5", effort: "high",
		timeout: 10 * time.Second, workDir: t.TempDir(), slots: make(chan struct{}, 1)}
}

func post(b *bridge, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/v1/run", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	b.handleRun(rec, req)
	return rec
}

const okBody = `{"system":"GM","prompt":"<user>hi</user>","json_schema":{"type":"object"}}`

func TestBridgeRunsClaudeWithToolsOff(t *testing.T) {
	bin, argsFile, stdinFile := fakeClaude(t, `{"type":"result","subtype":"success","is_error":false,"structured_output":{"text":"hello","tool_calls":[]},"usage":{"input_tokens":3,"cache_read_input_tokens":7,"output_tokens":2}}`)
	b := newTestBridge(t, bin)
	rec := post(b, strings.Repeat("t", 32), okBody)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d %s", rec.Code, rec.Body.String())
	}
	var out runResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	if !strings.Contains(string(out.StructuredOutput), "hello") || out.InputTokens != 10 {
		t.Errorf("out = %+v", out)
	}
	args, _ := os.ReadFile(argsFile)
	for _, want := range []string{"-p", "--tools\n\n", "--strict-mcp-config", "--model\nclaude-sonnet-5-5", "--effort\nhigh", "--no-session-persistence"} {
		if !strings.Contains(string(args), want) {
			t.Errorf("claude args missing %q:\n%s", want, args)
		}
	}
	if stdin, _ := os.ReadFile(stdinFile); string(stdin) != "<user>hi</user>" {
		t.Errorf("stdin = %q", stdin)
	}
}

func TestBridgeRejectsWrongToken(t *testing.T) {
	bin, _, _ := fakeClaude(t, `{}`)
	if rec := post(newTestBridge(t, bin), "wrong", okBody); rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestBridgeMapsUsageLimitTo429(t *testing.T) {
	bin, _, _ := fakeClaude(t, `{"type":"result","subtype":"error_during_execution","is_error":true,"api_error_status":429,"result":"usage limit reached"}`)
	if rec := post(newTestBridge(t, bin), strings.Repeat("t", 32), okBody); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", rec.Code)
	}
}

func TestClaudeEnvDropsAPIKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "sk-should-not-pass")
	t.Setenv("CLAUDE_CODE_OAUTH_TOKEN", "oauth")
	env := strings.Join(claudeEnv(), "\n")
	if strings.Contains(env, "ANTHROPIC_API_KEY") || !strings.Contains(env, "CLAUDE_CODE_OAUTH_TOKEN=oauth") {
		t.Errorf("env = %s", env)
	}
}

func TestBridgeSkillsOnlyWhenAllowedAndAsked(t *testing.T) {
	reply := `{"type":"result","subtype":"success","structured_output":{"text":"ok","tool_calls":[]},"usage":{}}`
	skillsBody := `{"system":"GM","prompt":"hi","json_schema":{"type":"object"},"skills":true}`
	for _, tc := range []struct {
		name          string
		bridgeAllows  bool
		body          string
		wantSkillTool bool
	}{
		{"allowed and asked", true, skillsBody, true},
		{"asked but not allowed", false, skillsBody, false},
		{"allowed but not asked", true, okBody, false},
	} {
		bin, argsFile, _ := fakeClaude(t, reply)
		b := newTestBridge(t, bin)
		b.skills = tc.bridgeAllows
		if rec := post(b, strings.Repeat("t", 32), tc.body); rec.Code != http.StatusOK {
			t.Fatalf("%s: status = %d", tc.name, rec.Code)
		}
		args, _ := os.ReadFile(argsFile)
		got := strings.Contains(string(args), "--tools\nSkill\n") && strings.Contains(string(args), "--setting-sources\nuser\n")
		if got != tc.wantSkillTool {
			t.Errorf("%s: Skill tool enabled = %v, want %v\n%s", tc.name, got, tc.wantSkillTool, args)
		}
		if !tc.wantSkillTool && !strings.Contains(string(args), "--tools\n\n") {
			t.Errorf("%s: tools not turned off:\n%s", tc.name, args)
		}
	}
}
