// Command claude-bridge lets Edda's claudecode provider use Claude through the
// unmodified Claude Code CLI, signed in with the owner's own Claude
// subscription. It is meant for one person's own games on their own machine:
// Anthropic intends subscription sign-in for ordinary use of Claude Code, and
// asks apps serving other people to use an API key instead.
//
// Each POST /v1/run starts `claude -p` with Claude Code's own tools turned
// off, a system prompt and a JSON schema, reads the prompt from stdin, and
// returns the structured result. Requests need the bearer token from
// -token-file. Model and effort are fixed by flags, not by the caller.
package main

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type runRequest struct {
	System     string          `json:"system"`
	Prompt     string          `json:"prompt"`
	JSONSchema json.RawMessage `json:"json_schema"`
	// Skills asks for Claude Code's Skill tool and the skills in
	// ~/.claude/skills of the bridge's HOME. Ignored unless -skills is set.
	Skills bool `json:"skills"`
}

// claudeResult is the part of `claude -p --output-format json` the bridge reads.
type claudeResult struct {
	Type             string          `json:"type"`
	Subtype          string          `json:"subtype"`
	IsError          bool            `json:"is_error"`
	Result           string          `json:"result"`
	StructuredOutput json.RawMessage `json:"structured_output"`
	StopReason       string          `json:"stop_reason"`
	APIErrorStatus   *int            `json:"api_error_status"`
	DurationMS       int             `json:"duration_ms"`
	Usage            struct {
		InputTokens              int `json:"input_tokens"`
		OutputTokens             int `json:"output_tokens"`
		CacheReadInputTokens     int `json:"cache_read_input_tokens"`
		CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	} `json:"usage"`
}

type runResponse struct {
	StructuredOutput json.RawMessage `json:"structured_output"`
	StopReason       string          `json:"stop_reason"`
	InputTokens      int             `json:"input_tokens"`
	OutputTokens     int             `json:"output_tokens"`
	DurationMS       int             `json:"duration_ms"`
}

type bridge struct {
	token   []byte
	claude  string
	model   string
	effort  string
	timeout time.Duration
	workDir string
	slots   chan struct{}
	skills  bool
}

func main() {
	listen := flag.String("listen", "127.0.0.1:18095", "address to listen on")
	tokenFile := flag.String("token-file", "", "file holding the bearer token callers must send (required)")
	claudeBin := flag.String("claude", "claude", "claude binary")
	model := flag.String("model", "claude-sonnet-5-5", "model passed to claude --model")
	effort := flag.String("effort", "high", "effort passed to claude --effort")
	concurrency := flag.Int("concurrency", 2, "claude processes allowed at once")
	timeout := flag.Duration("timeout", 5*time.Minute, "longest a single run may take")
	skills := flag.Bool("skills", false, "allow requests to use the Skill tool and the skills in $HOME/.claude/skills")
	flag.Parse()

	if *tokenFile == "" {
		fail(errors.New("-token-file is required"))
	}
	token, err := os.ReadFile(*tokenFile)
	if err != nil {
		fail(err)
	}
	token = bytes.TrimSpace(token)
	if len(token) < 32 {
		fail(errors.New("bearer token must be at least 32 characters"))
	}
	bin, err := exec.LookPath(*claudeBin)
	if err != nil {
		fail(fmt.Errorf("find claude: %w", err))
	}
	// Runs happen in an empty directory, so no CLAUDE.md or project
	// settings are picked up.
	workDir, err := os.MkdirTemp("", "claude-bridge-")
	if err != nil {
		fail(err)
	}
	defer func() { _ = os.RemoveAll(workDir) }()

	b := &bridge{token: token, claude: bin, model: *model, effort: *effort, timeout: *timeout,
		workDir: workDir, slots: make(chan struct{}, max(1, *concurrency)), skills: *skills}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "ok\n") })
	mux.HandleFunc("POST /v1/run", b.handleRun)

	srv := &http.Server{Addr: *listen, Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	slog.Info("claude-bridge listening", "addr", *listen, "model", *model, "effort", *effort)
	fail(srv.ListenAndServe())
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "claude-bridge:", err)
	os.Exit(1)
}

func (b *bridge) handleRun(w http.ResponseWriter, r *http.Request) {
	auth := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if subtle.ConstantTimeCompare([]byte(auth), b.token) != 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req runRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 8<<20)).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Prompt) == "" || len(req.JSONSchema) == 0 || !json.Valid(req.JSONSchema) {
		http.Error(w, "prompt and a valid json_schema are required", http.StatusBadRequest)
		return
	}

	select {
	case b.slots <- struct{}{}:
		defer func() { <-b.slots }()
	case <-r.Context().Done():
		return
	}

	res, status, err := b.run(r.Context(), req)
	if err != nil {
		slog.Warn("claude run failed", "status", status, "error", err)
		http.Error(w, err.Error(), status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(runResponse{
		StructuredOutput: res.StructuredOutput,
		StopReason:       res.StopReason,
		InputTokens:      res.Usage.InputTokens + res.Usage.CacheReadInputTokens + res.Usage.CacheCreationInputTokens,
		OutputTokens:     res.Usage.OutputTokens,
		DurationMS:       res.DurationMS,
	})
}

func (b *bridge) run(ctx context.Context, req runRequest) (*claudeResult, int, error) {
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	sysFile := filepath.Join(b.workDir, fmt.Sprintf("system-%d.txt", time.Now().UnixNano()))
	if err := os.WriteFile(sysFile, []byte(req.System), 0o600); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer func() { _ = os.Remove(sysFile) }()

	args := []string{
		"-p",
		"--output-format", "json",
		"--json-schema", string(req.JSONSchema),
		"--system-prompt-file", sysFile,
		"--model", b.model,
		"--effort", b.effort,
		"--strict-mcp-config",
		"--no-session-persistence",
	}
	if req.Skills && b.skills {
		// The Skill tool only reads skill files. Skills live in the user
		// scope, so user settings load too; keep HOME free of a settings
		// file.
		args = append(args, "--tools", "Skill", "--allowedTools", "Skill", "--setting-sources", "user")
	} else {
		args = append(args, "--tools", "", "--setting-sources", "")
	}
	cmd := exec.CommandContext(ctx, b.claude, args...)
	cmd.Dir = b.workDir
	cmd.Stdin = strings.NewReader(req.Prompt)
	cmd.Env = claudeEnv()
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	runErr := cmd.Run()

	var res claudeResult
	if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
		if ctx.Err() != nil {
			return nil, http.StatusGatewayTimeout, fmt.Errorf("claude run timed out after %s", b.timeout)
		}
		return nil, http.StatusBadGateway, fmt.Errorf("claude exited without a result (%v): %s", runErr, tail(stderr.String()))
	}
	if res.IsError || res.Subtype != "success" {
		status := http.StatusBadGateway
		if res.APIErrorStatus != nil && *res.APIErrorStatus == http.StatusTooManyRequests {
			status = http.StatusTooManyRequests
		}
		return nil, status, fmt.Errorf("claude run failed (%s): %s", res.Subtype, tail(res.Result))
	}
	if len(res.StructuredOutput) == 0 || string(res.StructuredOutput) == "null" {
		return nil, http.StatusBadGateway, fmt.Errorf("claude returned no structured output (stop_reason %q)", res.StopReason)
	}
	return &res, http.StatusOK, nil
}

// claudeEnv passes only what the CLI needs: the sign-in (from HOME, or
// CLAUDE_CODE_OAUTH_TOKEN for a `claude setup-token` token) and no API key,
// so a stray ANTHROPIC_API_KEY never switches the run to API billing.
func claudeEnv() []string {
	var env []string
	for _, kv := range os.Environ() {
		k, _, _ := strings.Cut(kv, "=")
		switch k {
		case "HOME", "PATH", "USER", "LANG", "TMPDIR", "XDG_CONFIG_HOME", "CLAUDE_CODE_OAUTH_TOKEN", "CLAUDE_CONFIG_DIR":
			env = append(env, kv)
		}
	}
	return append(env, "TERM=dumb", "DISABLE_AUTOUPDATER=1")
}

func tail(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 500 {
		s = "…" + s[len(s)-500:]
	}
	return s
}
