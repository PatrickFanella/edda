package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	defaultChatGPTBaseURL = "https://api.openai.com/v1"
	defaultChatGPTModel   = "gpt-6.1-terra"
	chatGPTToolNamespace  = "edda"
	chatGPTReplayCacheMax = 256
)

// ChatGPTClient implements Provider with the Responses API, paid for by the
// signed-in user's ChatGPT plan (OpenAI's "Sign in with ChatGPT" plan usage).
//
// That route requires store:false and stream:true, rejects system-role items
// and several request fields, and wants function tools grouped in a
// namespace. Without stored responses, the reasoning items that come with a
// tool call must be sent back with the tool results; the client keeps them in
// a small in-memory cache keyed by call ID, because Edda's message history
// does not carry provider data.
type ChatGPTClient struct {
	baseURL         string
	model           string
	reasoningEffort string
	tokens          *ChatGPTTokenStore
	client          *http.Client

	replayMu    sync.Mutex
	replay      map[string][]json.RawMessage
	replayOrder []string
}

// NewChatGPTClient returns a ChatGPT-plan provider using tokens.
func NewChatGPTClient(tokens *ChatGPTTokenStore, model, reasoningEffort string) *ChatGPTClient {
	if model == "" {
		model = defaultChatGPTModel
	}
	return &ChatGPTClient{
		baseURL:         defaultChatGPTBaseURL,
		model:           model,
		reasoningEffort: strings.TrimSpace(reasoningEffort),
		tokens:          tokens,
		client:          &http.Client{Timeout: 5 * time.Minute},
		replay:          map[string][]json.RawMessage{},
	}
}

// Complete sends a streamed Responses request and returns the finished reply.
func (c *ChatGPTClient) Complete(ctx context.Context, messages []Message, tools []Tool) (*Response, error) {
	return c.run(ctx, messages, tools)
}

// Stream sends a Responses request and forwards text and tool calls as they
// arrive. A request error after the stream opens ends the stream early.
func (c *ChatGPTClient) Stream(ctx context.Context, messages []Message, tools []Tool) (<-chan StreamChunk, error) {
	body, err := c.buildRequest(messages, tools)
	if err != nil {
		return nil, err
	}
	resp, err := c.open(ctx, body)
	if err != nil {
		return nil, err
	}
	ch := make(chan StreamChunk, 32)
	go func() {
		defer close(ch)
		defer func() { _ = resp.Body.Close() }()
		send := func(chunk StreamChunk) {
			select {
			case ch <- chunk:
			case <-ctx.Done():
			}
		}
		_, _ = c.readStream(resp.Body, send)
		send(StreamChunk{Done: true})
	}()
	return ch, nil
}

func (c *ChatGPTClient) run(ctx context.Context, messages []Message, tools []Tool) (*Response, error) {
	body, err := c.buildRequest(messages, tools)
	if err != nil {
		return nil, err
	}
	resp, err := c.open(ctx, body)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	out, err := c.readStream(resp.Body, nil)
	if err != nil {
		if ctx.Err() != nil {
			return nil, &ErrTimeout{URL: c.responsesURL(), Err: ctx.Err()}
		}
		return nil, err
	}
	return out, nil
}

type chatGPTRequest struct {
	Model     string            `json:"model"`
	Input     []any             `json:"input"`
	Tools     []chatGPTToolSpec `json:"tools,omitempty"`
	Reasoning *chatGPTReasoning `json:"reasoning,omitempty"`
	Store     bool              `json:"store"`
	Stream    bool              `json:"stream"`
}

type chatGPTReasoning struct {
	Effort string `json:"effort"`
}

type chatGPTToolSpec struct {
	Type        string            `json:"type"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Parameters  map[string]any    `json:"parameters,omitempty"`
	Tools       []chatGPTToolSpec `json:"tools,omitempty"`
}

type chatGPTMessageItem struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatGPTFunctionCallItem struct {
	Type      string `json:"type"`
	CallID    string `json:"call_id"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Arguments string `json:"arguments"`
}

type chatGPTFunctionOutputItem struct {
	Type   string `json:"type"`
	CallID string `json:"call_id"`
	Output string `json:"output"`
}

func (c *ChatGPTClient) buildRequest(messages []Message, tools []Tool) ([]byte, error) {
	req := chatGPTRequest{Model: c.model, Store: false, Stream: true}
	if c.reasoningEffort != "" {
		req.Reasoning = &chatGPTReasoning{Effort: c.reasoningEffort}
	}
	if len(tools) > 0 {
		ns := chatGPTToolSpec{Type: "namespace", Name: chatGPTToolNamespace, Description: "Edda game-master tools."}
		for _, t := range tools {
			ns.Tools = append(ns.Tools, chatGPTToolSpec{Type: "function", Name: t.Name, Description: t.Description, Parameters: t.Parameters})
		}
		req.Tools = []chatGPTToolSpec{ns}
	}

	for _, m := range messages {
		switch m.Role {
		case RoleSystem:
			// This route rejects system-role items; developer messages carry
			// the same instructions and keep their place in the history.
			req.Input = append(req.Input, chatGPTMessageItem{Role: "developer", Content: m.Content})
		case RoleUser:
			req.Input = append(req.Input, chatGPTMessageItem{Role: "user", Content: m.Content})
		case RoleAssistant:
			if len(m.ToolCalls) == 0 {
				req.Input = append(req.Input, chatGPTMessageItem{Role: "assistant", Content: m.Content})
				continue
			}
			if items := c.replayItems(m.ToolCalls[0].ID); items != nil {
				for _, item := range items {
					req.Input = append(req.Input, item)
				}
				continue
			}
			// No saved output (another provider made these calls, or Edda
			// restarted): rebuild the calls without reasoning.
			if m.Content != "" {
				req.Input = append(req.Input, chatGPTMessageItem{Role: "assistant", Content: m.Content})
			}
			for _, tc := range m.ToolCalls {
				args, err := json.Marshal(tc.Arguments)
				if err != nil {
					return nil, fmt.Errorf("marshal tool call arguments: %w", err)
				}
				if tc.Arguments == nil {
					args = []byte("{}")
				}
				req.Input = append(req.Input, chatGPTFunctionCallItem{
					Type: "function_call", CallID: tc.ID, Name: tc.Name, Namespace: chatGPTToolNamespace, Arguments: string(args),
				})
			}
		case RoleTool:
			req.Input = append(req.Input, chatGPTFunctionOutputItem{Type: "function_call_output", CallID: m.ToolCallID, Output: m.Content})
		default:
			return nil, fmt.Errorf("unsupported message role %q", m.Role)
		}
	}
	return json.Marshal(req)
}

// open sends the request, refreshing the access token once if it is rejected.
func (c *ChatGPTClient) open(ctx context.Context, body []byte) (*http.Response, error) {
	token, err := c.tokens.AccessToken(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := c.post(ctx, body, token)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		_ = resp.Body.Close()
		if token, err = c.tokens.RefreshAfterReject(ctx, token); err != nil {
			return nil, err
		}
		if resp, err = c.post(ctx, body, token); err != nil {
			return nil, err
		}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer func() { _ = resp.Body.Close() }()
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
		return nil, classifyChatGPTError(c.responsesURL(), c.model, resp.StatusCode, data)
	}
	return resp, nil
}

func (c *ChatGPTClient) post(ctx context.Context, body []byte, token string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.responsesURL(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	resp, err := c.client.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, &ErrTimeout{URL: c.responsesURL(), Err: err}
		}
		return nil, &ErrConnection{URL: c.responsesURL(), Err: err}
	}
	return resp, nil
}

func (c *ChatGPTClient) responsesURL() string {
	return strings.TrimRight(c.baseURL, "/") + "/responses"
}

type chatGPTEvent struct {
	Type     string          `json:"type"`
	Delta    string          `json:"delta"`
	Item     json.RawMessage `json:"item"`
	Response *struct {
		Status            string `json:"status"`
		IncompleteDetails *struct {
			Reason string `json:"reason"`
		} `json:"incomplete_details"`
		Error *chatGPTAPIError `json:"error"`
		Usage *struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
			TotalTokens  int `json:"total_tokens"`
		} `json:"usage"`
	} `json:"response"`
	// Top-level fields of an "error" event.
	Code    string `json:"code"`
	Message string `json:"message"`
	Param   string `json:"param"`
}

type chatGPTAPIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Param   string `json:"param"`
}

type chatGPTOutputItem struct {
	Type      string `json:"type"`
	CallID    string `json:"call_id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
	Content   []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// readStream reads Responses server-sent events until response.completed.
// send, when set, receives text deltas and finished tool calls.
func (c *ChatGPTClient) readStream(body io.Reader, send func(StreamChunk)) (*Response, error) {
	endpoint := c.responsesURL()
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64<<10), 8<<20)

	out := &Response{}
	var items []json.RawMessage
	var streamedText strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}
		var ev chatGPTEvent
		if err := json.Unmarshal([]byte(data), &ev); err != nil {
			return nil, &ErrMalformedResponse{URL: endpoint, Err: fmt.Errorf("decode stream event: %w", err)}
		}

		switch ev.Type {
		case "response.output_text.delta":
			streamedText.WriteString(ev.Delta)
			if send != nil && ev.Delta != "" {
				send(StreamChunk{ContentDelta: ev.Delta})
			}
		case "response.output_item.done":
			if len(ev.Item) == 0 {
				continue
			}
			items = append(items, append(json.RawMessage(nil), ev.Item...))
			var item chatGPTOutputItem
			if err := json.Unmarshal(ev.Item, &item); err != nil {
				return nil, &ErrMalformedResponse{URL: endpoint, Err: fmt.Errorf("decode output item: %w", err)}
			}
			switch item.Type {
			case "message":
				for _, part := range item.Content {
					if part.Type == "output_text" {
						out.Content += part.Text
					}
				}
			case "function_call":
				args := map[string]any{}
				if strings.TrimSpace(item.Arguments) != "" {
					if err := json.Unmarshal([]byte(item.Arguments), &args); err != nil {
						return nil, &ErrMalformedResponse{URL: endpoint, Err: fmt.Errorf("decode arguments for %s: %w", item.Name, err)}
					}
				}
				tc := ToolCall{ID: item.CallID, Name: item.Name, Arguments: args}
				out.ToolCalls = append(out.ToolCalls, tc)
				if send != nil {
					send(StreamChunk{ToolCallDelta: &tc})
				}
			}
		case "response.completed":
			if ev.Response != nil && ev.Response.Usage != nil {
				out.Usage = Usage{
					PromptTokens:     ev.Response.Usage.InputTokens,
					CompletionTokens: ev.Response.Usage.OutputTokens,
					TotalTokens:      ev.Response.Usage.TotalTokens,
				}
			}
			if out.Content == "" {
				out.Content = streamedText.String()
			}
			out.FinishReason = "stop"
			if len(out.ToolCalls) > 0 {
				out.FinishReason = "tool_calls"
				c.saveReplay(out.ToolCalls, items)
			}
			return out, nil
		case "response.failed":
			apiErr := &chatGPTAPIError{Code: "response_failed"}
			if ev.Response != nil && ev.Response.Error != nil {
				apiErr = ev.Response.Error
			}
			return nil, classifyChatGPTCode(endpoint, c.model, 0, apiErr)
		case "response.incomplete":
			reason := "unknown"
			if ev.Response != nil && ev.Response.IncompleteDetails != nil {
				reason = ev.Response.IncompleteDetails.Reason
			}
			return nil, &ErrMalformedResponse{URL: endpoint, Err: fmt.Errorf("response incomplete: %s", reason)}
		case "error":
			return nil, classifyChatGPTCode(endpoint, c.model, 0, &chatGPTAPIError{Code: ev.Code, Message: ev.Message, Param: ev.Param})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, &ErrConnection{URL: endpoint, Err: fmt.Errorf("read stream: %w", err)}
	}
	return nil, &ErrConnection{URL: endpoint, Err: errors.New("stream ended before response.completed")}
}

// saveReplay keeps a response's output items so the next request can send
// them back with the tool results.
func (c *ChatGPTClient) saveReplay(calls []ToolCall, items []json.RawMessage) {
	key := calls[0].ID
	if key == "" {
		return
	}
	c.replayMu.Lock()
	defer c.replayMu.Unlock()
	if _, ok := c.replay[key]; !ok {
		c.replayOrder = append(c.replayOrder, key)
	}
	c.replay[key] = items
	for len(c.replayOrder) > chatGPTReplayCacheMax {
		delete(c.replay, c.replayOrder[0])
		c.replayOrder = c.replayOrder[1:]
	}
}

func (c *ChatGPTClient) replayItems(firstCallID string) []json.RawMessage {
	if firstCallID == "" {
		return nil
	}
	c.replayMu.Lock()
	defer c.replayMu.Unlock()
	return c.replay[firstCallID]
}

func classifyChatGPTError(endpoint, model string, status int, body []byte) error {
	var parsed struct {
		Error  *chatGPTAPIError `json:"error"`
		Detail string           `json:"detail"`
	}
	_ = json.Unmarshal(body, &parsed)
	apiErr := parsed.Error
	if apiErr == nil {
		msg := parsed.Detail
		if msg == "" {
			msg = sanitizeErrorBody(string(body))
		}
		apiErr = &chatGPTAPIError{Message: msg}
	}
	return classifyChatGPTCode(endpoint, model, status, apiErr)
}

func classifyChatGPTCode(endpoint, model string, status int, apiErr *chatGPTAPIError) error {
	msg := apiErr.Message
	if apiErr.Code != "" {
		msg = apiErr.Code + ": " + msg
	}
	if apiErr.Param != "" {
		msg += " (param " + apiErr.Param + ")"
	}
	base := fmt.Errorf("chatgpt: %s", strings.TrimSpace(msg))

	switch apiErr.Code {
	case "subscription_sharing_usage_limit_exceeded":
		return &ErrRateLimit{URL: endpoint, StatusCode: http.StatusTooManyRequests,
			Err: fmt.Errorf("%w; ChatGPT plan limit reached, see https://chatgpt.com/settings/usage", base)}
	case "subscription_sharing_usage_unavailable", "subscription_sharing_user_unavailable":
		return &ErrTransient{URL: endpoint, StatusCode: http.StatusServiceUnavailable, Err: base}
	case "subscription_sharing_user_not_eligible", "subscription_sharing_invalid_user",
		"subscription_sharing_route_not_supported", "chatpass_v2_scope_not_authorized", "chatpass_v2_invalid_authorization_context":
		return &ErrAuth{URL: endpoint, StatusCode: status, Err: base}
	case "model_not_found":
		return &ErrModelNotFound{URL: endpoint, Model: model, StatusCode: status, Err: base}
	case "subscription_sharing_unsupported_capability":
		return &ErrMalformedResponse{URL: endpoint, Err: base}
	}

	switch {
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return &ErrAuth{URL: endpoint, StatusCode: status, Err: base}
	case status == http.StatusTooManyRequests:
		return &ErrRateLimit{URL: endpoint, StatusCode: status, Err: base}
	case status == http.StatusNotFound:
		return &ErrModelNotFound{URL: endpoint, Model: model, StatusCode: status, Err: base}
	case status == http.StatusBadRequest:
		return &ErrMalformedResponse{URL: endpoint, Err: base}
	case status >= 500 || status == 0:
		return &ErrTransient{URL: endpoint, StatusCode: status, Err: base}
	default:
		return &ErrConnection{URL: endpoint, Err: base}
	}
}
