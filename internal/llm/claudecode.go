package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

// ClaudeCodeClient implements Provider through cmd/claude-bridge, which runs
// the Claude Code CLI on the owner's Claude subscription. The CLI takes one
// prompt and returns one structured reply, so the client writes the
// conversation and the tool list into the prompt and asks for
// {"text", "tool_calls"} back. Claude Code's own tools stay off; Edda runs
// the returned tool calls itself, as with every other provider.
type ClaudeCodeClient struct {
	endpoint string
	token    string
	client   *http.Client
	calls    atomic.Uint64
}

// NewClaudeCodeClient returns a provider that calls the bridge at endpoint
// (for example http://10.0.0.50:18095) with the bearer token.
func NewClaudeCodeClient(endpoint, token string) *ClaudeCodeClient {
	return &ClaudeCodeClient{
		endpoint: strings.TrimRight(endpoint, "/"),
		token:    token,
		client:   &http.Client{Timeout: 6 * time.Minute},
	}
}

var claudeCodeReplySchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "text": {"type": "string", "description": "What the assistant says to the player. Empty when the turn is only tool calls."},
    "tool_calls": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": {"type": "string"},
          "arguments": {"type": "object"}
        },
        "required": ["name", "arguments"]
      }
    }
  },
  "required": ["text", "tool_calls"]
}`)

const claudeCodeFormatInstructions = `You are the assistant in the conversation below, and you write its next turn.
The conversation arrives as tagged text: <user>, <assistant>, <tool_call> and
<tool_result>. Reply only with the JSON the response schema describes: put what
you say in "text", and put each tool you call in "tool_calls" with its name and
arguments that match the tool's parameters. Call only the tools listed in
<tools>. When you call tools, their results come back in the next request.`

// Complete implements Provider.
func (c *ClaudeCodeClient) Complete(ctx context.Context, messages []Message, tools []Tool) (*Response, error) {
	system, prompt, err := buildClaudeCodePrompt(messages, tools)
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal(map[string]any{"system": system, "prompt": prompt, "json_schema": claudeCodeReplySchema})
	if err != nil {
		return nil, err
	}
	url := c.endpoint + "/v1/run"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, &ErrTimeout{URL: url, Err: err}
		}
		return nil, &ErrConnection{URL: url, Err: err}
	}
	defer func() { _ = resp.Body.Close() }()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))

	if resp.StatusCode != http.StatusOK {
		base := fmt.Errorf("claude-bridge: %s", sanitizeErrorBody(strings.TrimSpace(string(data))))
		switch {
		case resp.StatusCode == http.StatusUnauthorized:
			return nil, &ErrAuth{URL: url, StatusCode: resp.StatusCode, Err: base}
		case resp.StatusCode == http.StatusTooManyRequests:
			return nil, &ErrRateLimit{URL: url, StatusCode: resp.StatusCode, Err: base}
		case resp.StatusCode == http.StatusBadRequest:
			return nil, &ErrMalformedResponse{URL: url, Err: base}
		default:
			return nil, &ErrTransient{URL: url, StatusCode: resp.StatusCode, Err: base}
		}
	}

	var out struct {
		StructuredOutput struct {
			Text      string `json:"text"`
			ToolCalls []struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments"`
			} `json:"tool_calls"`
		} `json:"structured_output"`
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, &ErrMalformedResponse{URL: url, Err: fmt.Errorf("decode bridge reply: %w", err)}
	}

	known := map[string]bool{}
	for _, t := range tools {
		known[t.Name] = true
	}
	r := &Response{
		Content:      out.StructuredOutput.Text,
		FinishReason: "stop",
		Usage: Usage{
			PromptTokens:     out.InputTokens,
			CompletionTokens: out.OutputTokens,
			TotalTokens:      out.InputTokens + out.OutputTokens,
		},
	}
	for _, tc := range out.StructuredOutput.ToolCalls {
		if !known[tc.Name] {
			return nil, &ErrMalformedResponse{URL: url, Err: fmt.Errorf("model called unknown tool %q", tc.Name)}
		}
		if tc.Arguments == nil {
			tc.Arguments = map[string]any{}
		}
		r.ToolCalls = append(r.ToolCalls, ToolCall{
			ID:        fmt.Sprintf("cc_%d", c.calls.Add(1)),
			Name:      tc.Name,
			Arguments: tc.Arguments,
		})
	}
	if len(r.ToolCalls) > 0 {
		r.FinishReason = "tool_calls"
	}
	return r, nil
}

// Stream implements Provider. The CLI returns the reply in one piece, so the
// stream carries it as a single chunk.
func (c *ClaudeCodeClient) Stream(ctx context.Context, messages []Message, tools []Tool) (<-chan StreamChunk, error) {
	resp, err := c.Complete(ctx, messages, tools)
	if err != nil {
		return nil, err
	}
	ch := make(chan StreamChunk, 2+len(resp.ToolCalls))
	if resp.Content != "" {
		ch <- StreamChunk{ContentDelta: resp.Content}
	}
	for i := range resp.ToolCalls {
		ch <- StreamChunk{ToolCallDelta: &resp.ToolCalls[i]}
	}
	ch <- StreamChunk{Done: true}
	close(ch)
	return ch, nil
}

// buildClaudeCodePrompt returns the system prompt (Edda's system messages
// plus the reply format) and the prompt (tool list and tagged conversation).
func buildClaudeCodePrompt(messages []Message, tools []Tool) (string, string, error) {
	var system []string
	var conv strings.Builder
	for _, m := range messages {
		switch m.Role {
		case RoleSystem:
			if m.Content != "" {
				system = append(system, m.Content)
			}
		case RoleUser:
			fmt.Fprintf(&conv, "<user>\n%s\n</user>\n", m.Content)
		case RoleAssistant:
			conv.WriteString("<assistant>\n")
			if m.Content != "" {
				conv.WriteString(m.Content + "\n")
			}
			for _, tc := range m.ToolCalls {
				args, err := json.Marshal(tc.Arguments)
				if err != nil {
					return "", "", fmt.Errorf("marshal tool call arguments: %w", err)
				}
				fmt.Fprintf(&conv, "<tool_call id=%q name=%q>%s</tool_call>\n", tc.ID, tc.Name, args)
			}
			conv.WriteString("</assistant>\n")
		case RoleTool:
			fmt.Fprintf(&conv, "<tool_result id=%q>\n%s\n</tool_result>\n", m.ToolCallID, m.Content)
		default:
			return "", "", fmt.Errorf("unsupported message role %q", m.Role)
		}
	}

	var prompt strings.Builder
	if len(tools) > 0 {
		specs := make([]map[string]any, 0, len(tools))
		for _, t := range tools {
			specs = append(specs, map[string]any{"name": t.Name, "description": t.Description, "parameters": t.Parameters})
		}
		data, err := json.Marshal(specs)
		if err != nil {
			return "", "", fmt.Errorf("marshal tools: %w", err)
		}
		fmt.Fprintf(&prompt, "<tools>\n%s\n</tools>\n\n", data)
	} else {
		prompt.WriteString("<tools>[]</tools>\nNo tools are available for this turn; leave tool_calls empty.\n\n")
	}
	prompt.WriteString("<conversation>\n")
	prompt.WriteString(conv.String())
	prompt.WriteString("</conversation>\n\nWrite the assistant's next turn.")

	system = append(system, claudeCodeFormatInstructions)
	return strings.Join(system, "\n\n"), prompt.String(), nil
}
