package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultOllamaBaseURL = "http://localhost:11434"
	defaultOllamaModel   = "llama3.2"
	defaultOllamaTimeout = 3 * time.Minute
	ollamaChatPath       = "/api/chat"

	// Broker-specific SSE terminal statuses (see llama-line skill).
	brokerStatusQueued           = "queued"
	brokerStatusOllamaUnavail    = "ollama_unavailable"
	brokerStatusDroppedByAdmin   = "dropped_by_admin"
	brokerHeaderResponse         = "X-Ollama-Broker"
	brokerHeaderCache            = "X-Llama-Line-Cache"
)

func ollamaLogger() *slog.Logger {
	return slog.Default().WithGroup("ollama")
}

// OllamaClient implements Provider via Ollama's HTTP API. It also speaks the
// llama-line broker dialect (Bearer auth + SSE-wrapped responses with
// pre-pended queue status events). The broker dialect is auto-detected from
// the response Content-Type / body framing, so a single client can talk to
// either a vanilla ollama server or a llama-line broker.
type OllamaClient struct {
	baseURL string
	model   string
	numCtx  int
	apiKey  string
	client  *http.Client
}

const defaultOllamaNumCtx = 16384

func boolPtr(b bool) *bool { return &b }

// NewOllamaClient returns an Ollama-backed provider using the default request timeout.
func NewOllamaClient(baseURL, model string) *OllamaClient {
	return NewOllamaClientWithTimeout(baseURL, model, defaultOllamaTimeout)
}

// NewOllamaClientWithTimeout returns an Ollama-backed provider using the
// supplied request timeout for non-streaming requests.
func NewOllamaClientWithTimeout(baseURL, model string, timeout time.Duration) *OllamaClient {
	if baseURL == "" {
		baseURL = defaultOllamaBaseURL
	}
	if model == "" {
		model = defaultOllamaModel
	}
	if timeout <= 0 {
		timeout = defaultOllamaTimeout
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &OllamaClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		numCtx:  defaultOllamaNumCtx,
		client: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}
}

// WithAPIKey attaches a Bearer token to all chat requests. Required when the
// configured endpoint is a llama-line broker; ignored by vanilla ollama. The
// receiver is returned to support chaining at construction time.
func (o *OllamaClient) WithAPIKey(key string) *OllamaClient {
	o.apiKey = strings.TrimSpace(key)
	return o
}

func (o *OllamaClient) Complete(ctx context.Context, messages []Message, tools []Tool) (*Response, error) {
	resp, err := o.callChat(ctx, messages, tools, false)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	chatResp, err := o.readNonStreamingResponse(resp)
	if err != nil {
		return nil, err
	}

	toolCalls, err := fromOllamaToolCalls(chatResp.Message.ToolCalls)
	if err != nil {
		return nil, &ErrMalformedResponse{
			URL: o.baseURL + ollamaChatPath,
			Err: err,
		}
	}

	if chatResp.Message.Content == "" && len(toolCalls) == 0 {
		ollamaLogger().Warn("ollama returned empty response",
			"model", o.model,
		)
	}

	return &Response{
		Content:      chatResp.Message.Content,
		ToolCalls:    toolCalls,
		FinishReason: chatResp.DoneReason,
		Usage: Usage{
			PromptTokens:     chatResp.PromptEvalCount,
			CompletionTokens: chatResp.EvalCount,
			TotalTokens:      chatResp.PromptEvalCount + chatResp.EvalCount,
		},
	}, nil
}

// readNonStreamingResponse reads a non-streaming chat response, transparently
// handling both raw ollama JSON bodies and llama-line broker SSE-wrapped
// bodies (with optional leading status events).
func (o *OllamaClient) readNonStreamingResponse(resp *http.Response) (*ollamaChatResponse, error) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &ErrMalformedResponse{
			URL: o.baseURL + ollamaChatPath,
			Err: fmt.Errorf("failed to read ollama chat response body: %w", err),
		}
	}

	if !isSSEResponse(resp, bodyBytes) {
		var chatResp ollamaChatResponse
		if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
			return nil, &ErrMalformedResponse{
				URL: o.baseURL + ollamaChatPath,
				Err: fmt.Errorf("failed to decode ollama chat response: %w", err),
			}
		}
		return &chatResp, nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(bodyBytes))
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		payload, kind, perr := parseSSELine(scanner.Bytes())
		if perr != nil {
			return nil, &ErrMalformedResponse{URL: o.baseURL + ollamaChatPath, Err: perr}
		}
		switch kind {
		case sseLineSkip:
			continue
		case sseLineBrokerStatus:
			ollamaLogger().Debug("broker status",
				"request_id", payload.RequestID,
				"status", payload.Status,
				"position", payload.Position,
				"wait_seconds", payload.WaitSeconds,
			)
			if terr := brokerTerminalError(o.baseURL+ollamaChatPath, payload); terr != nil {
				return nil, terr
			}
			continue
		case sseLinePayload:
			var chatResp ollamaChatResponse
			if err := json.Unmarshal(payload.Raw, &chatResp); err != nil {
				return nil, &ErrMalformedResponse{
					URL: o.baseURL + ollamaChatPath,
					Err: fmt.Errorf("failed to decode ollama chat response: %w", err),
				}
			}
			return &chatResp, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, &ErrMalformedResponse{
			URL: o.baseURL + ollamaChatPath,
			Err: fmt.Errorf("failed to read broker SSE stream: %w", err),
		}
	}
	return nil, &ErrMalformedResponse{
		URL: o.baseURL + ollamaChatPath,
		Err: errors.New("broker SSE stream ended without a response payload"),
	}
}

func (o *OllamaClient) Stream(ctx context.Context, messages []Message, tools []Tool) (<-chan StreamChunk, error) {
	resp, err := o.callChat(ctx, messages, tools, true)
	if err != nil {
		return nil, err
	}

	ch := make(chan StreamChunk)
	go func() {
		defer close(ch)
		defer func() { _ = resp.Body.Close() }()

		// Peek a small prefix to detect framing without consuming bytes
		// destructively. We wrap the body in a buffered reader so the
		// scanner sees the full payload regardless of detection path.
		reader := bufio.NewReaderSize(resp.Body, 64*1024)
		sse := isSSEStreamResponse(resp, reader)

		scanner := bufio.NewScanner(reader)
		scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
		for scanner.Scan() {
			line := scanner.Bytes()
			var chunkBytes []byte

			if sse {
				payload, kind, perr := parseSSELine(line)
				if perr != nil {
					return
				}
				switch kind {
				case sseLineSkip:
					continue
				case sseLineBrokerStatus:
					if terr := brokerTerminalError(o.baseURL+ollamaChatPath, payload); terr != nil {
						// Best-effort terminal: surface via a final empty Done chunk;
						// the engine layer is responsible for error reporting on Stream
						// since the channel interface does not carry errors.
						ollamaLogger().Error("broker terminal status during stream",
							"error", terr,
						)
						return
					}
					continue
				case sseLinePayload:
					if bytes.Equal(bytes.TrimSpace(payload.Raw), []byte("[DONE]")) {
						return
					}
					chunkBytes = payload.Raw
				}
			} else {
				if len(bytes.TrimSpace(line)) == 0 {
					continue
				}
				chunkBytes = line
			}

			var chunkResp ollamaChatResponse
			if err := json.Unmarshal(chunkBytes, &chunkResp); err != nil {
				return
			}

			var toolDelta *ToolCall
			if len(chunkResp.Message.ToolCalls) > 0 {
				calls, err := fromOllamaToolCalls(chunkResp.Message.ToolCalls)
				if err == nil && len(calls) > 0 {
					toolDelta = &calls[0]
				}
			}

			select {
			case ch <- StreamChunk{
				ContentDelta:  chunkResp.Message.Content,
				ToolCallDelta: toolDelta,
				Done:          chunkResp.Done,
			}:
			case <-ctx.Done():
				return
			}
			if chunkResp.Done {
				return
			}
		}
	}()

	return ch, nil
}

func (o *OllamaClient) callChat(ctx context.Context, messages []Message, tools []Tool, stream bool) (*http.Response, error) {
	chatURL, err := o.chatURL()
	if err != nil {
		return nil, err
	}

	ollamaMessages, err := toOllamaMessages(messages)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(ollamaChatRequest{
		Model:    o.model,
		Messages: ollamaMessages,
		Tools:    toOllamaTools(tools),
		Stream:   stream,
		Options:  ollamaModelOptions{NumCtx: o.numCtx, Think: boolPtr(false)},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ollama chat request: %w", err)
	}

	started := time.Now()
	ollamaLogger().Debug("chat request starting",
		"model", o.model,
		"url", chatURL,
		"stream", stream,
		"messages", len(messages),
		"tools", len(tools),
		"body_len", len(body),
		"auth", o.apiKey != "",
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, chatURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create ollama chat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if o.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+o.apiKey)
		req.Header.Set("Accept", "text/event-stream")
	}

	resp, err := o.client.Do(req)
	if err != nil {
		formatted := formatConnectionError(chatURL, err)
		ollamaLogger().Error("chat request failed",
			"model", o.model,
			"url", chatURL,
			"stream", stream,
			"duration_ms", time.Since(started).Milliseconds(),
			"error", formatted,
		)
		return nil, formatted
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer func() { _ = resp.Body.Close() }()
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		classified := classifyHTTPError(chatURL, o.model, resp.StatusCode, strings.TrimSpace(string(respBody)))
		ollamaLogger().Error("chat request returned non-success status",
			"model", o.model,
			"url", chatURL,
			"stream", stream,
			"status_code", resp.StatusCode,
			"broker", resp.Header.Get(brokerHeaderResponse) != "",
			"duration_ms", time.Since(started).Milliseconds(),
			"error", classified,
		)
		return nil, classified
	}
	ollamaLogger().Debug("chat request completed",
		"model", o.model,
		"url", chatURL,
		"stream", stream,
		"status_code", resp.StatusCode,
		"cache", resp.Header.Get(brokerHeaderCache),
		"duration_ms", time.Since(started).Milliseconds(),
	)
	return resp, nil
}

func (o *OllamaClient) chatURL() (string, error) {
	if _, err := url.ParseRequestURI(o.baseURL); err != nil {
		return "", fmt.Errorf("invalid ollama base url %q: %w", o.baseURL, err)
	}
	return o.baseURL + ollamaChatPath, nil
}

// --- SSE / broker helpers ---------------------------------------------------

// brokerStatusEvent matches both the heartbeat queue status event and the
// terminal broker error/dropped events emitted by llama-line.
type brokerStatusEvent struct {
	RequestID   string `json:"request_id"`
	Status      string `json:"status"`
	Position    int    `json:"position,omitempty"`
	WaitSeconds int    `json:"wait_seconds,omitempty"`
	Message     string `json:"message,omitempty"`

	// Raw retains the original JSON bytes so that, when this payload is
	// actually an ollama response (no broker `status` field), callers can
	// re-decode into the ollama-specific struct without re-marshaling.
	Raw json.RawMessage `json:"-"`
}

type sseLineKind int

const (
	sseLineSkip sseLineKind = iota
	sseLineBrokerStatus
	sseLinePayload
)

// parseSSELine classifies a single line from an SSE stream. Returns:
//   - sseLineSkip for comments, empty lines, or non-data lines
//   - sseLineBrokerStatus when the data payload is a broker status event
//   - sseLinePayload when the data payload should be parsed as an ollama body
//     (or a sentinel like "[DONE]")
func parseSSELine(line []byte) (brokerStatusEvent, sseLineKind, error) {
	trimmed := bytes.TrimRight(line, "\r")
	if len(bytes.TrimSpace(trimmed)) == 0 {
		return brokerStatusEvent{}, sseLineSkip, nil
	}
	if bytes.HasPrefix(trimmed, []byte(":")) {
		return brokerStatusEvent{}, sseLineSkip, nil
	}
	if !bytes.HasPrefix(trimmed, []byte("data:")) {
		// event:/id:/retry: lines or unknown framing — ignore.
		return brokerStatusEvent{}, sseLineSkip, nil
	}
	raw := bytes.TrimSpace(trimmed[len("data:"):])
	if len(raw) == 0 {
		return brokerStatusEvent{}, sseLineSkip, nil
	}

	// Stream-end sentinel: pass through as a payload line so the streaming
	// loop can decide whether to terminate.
	if bytes.Equal(raw, []byte("[DONE]")) {
		return brokerStatusEvent{Raw: append([]byte(nil), raw...)}, sseLinePayload, nil
	}

	// Broker status events always include both `request_id` AND `status`.
	// Inspect the JSON minimally to discriminate without forcing the ollama
	// response through a tolerant decode.
	var probe struct {
		RequestID string `json:"request_id"`
		Status    string `json:"status"`
	}
	if err := json.Unmarshal(raw, &probe); err == nil && probe.RequestID != "" && isBrokerStatus(probe.Status) {
		var evt brokerStatusEvent
		if err := json.Unmarshal(raw, &evt); err != nil {
			return brokerStatusEvent{}, sseLineSkip, fmt.Errorf("decode broker status event: %w", err)
		}
		evt.Raw = append([]byte(nil), raw...)
		return evt, sseLineBrokerStatus, nil
	}

	return brokerStatusEvent{Raw: append([]byte(nil), raw...)}, sseLinePayload, nil
}

func isBrokerStatus(s string) bool {
	switch s {
	case brokerStatusQueued, brokerStatusOllamaUnavail, brokerStatusDroppedByAdmin:
		return true
	}
	return false
}

// brokerTerminalError converts a terminal broker status event into a typed
// error. Returns nil for non-terminal (queued) events.
func brokerTerminalError(endpoint string, evt brokerStatusEvent) error {
	switch evt.Status {
	case brokerStatusOllamaUnavail:
		msg := evt.Message
		if msg == "" {
			msg = "ollama unavailable"
		}
		return &ErrConnection{
			URL: endpoint,
			Err: fmt.Errorf("llama-line broker reported ollama_unavailable (request_id=%s): %s", evt.RequestID, msg),
		}
	case brokerStatusDroppedByAdmin:
		return &ErrConnection{
			URL: endpoint,
			Err: fmt.Errorf("llama-line broker dropped request by admin (request_id=%s)", evt.RequestID),
		}
	}
	return nil
}

// isSSEResponse returns true when the response should be parsed as SSE rather
// than as a single JSON document. Detection is conservative: prefer the
// Content-Type header, then fall back to inspecting the body prefix.
func isSSEResponse(resp *http.Response, body []byte) bool {
	if resp != nil {
		if ct := resp.Header.Get("Content-Type"); strings.Contains(strings.ToLower(ct), "text/event-stream") {
			return true
		}
		if resp.Header.Get(brokerHeaderResponse) != "" || resp.Header.Get(brokerHeaderCache) != "" {
			return true
		}
	}
	trimmed := bytes.TrimLeft(body, " \t\r\n")
	return bytes.HasPrefix(trimmed, []byte("data:")) || bytes.HasPrefix(trimmed, []byte(":"))
}

// isSSEStreamResponse mirrors isSSEResponse for streaming responses where we
// cannot pre-buffer the entire body. Falls back to peeking the bufio reader.
func isSSEStreamResponse(resp *http.Response, r *bufio.Reader) bool {
	if resp != nil {
		if ct := resp.Header.Get("Content-Type"); strings.Contains(strings.ToLower(ct), "text/event-stream") {
			return true
		}
		if resp.Header.Get(brokerHeaderResponse) != "" {
			return true
		}
	}
	peek, _ := r.Peek(8)
	trimmed := bytes.TrimLeft(peek, " \t\r\n")
	return bytes.HasPrefix(trimmed, []byte("data:")) || bytes.HasPrefix(trimmed, []byte(":"))
}

// ---------------------------------------------------------------------------

func formatConnectionError(endpoint string, err error) error {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return &ErrTimeout{URL: endpoint, Err: err}
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		if urlErr.Timeout() {
			return &ErrTimeout{URL: endpoint, Err: urlErr}
		}
		return &ErrConnection{URL: endpoint, Err: urlErr.Err}
	}
	return &ErrConnection{URL: endpoint, Err: err}
}

func classifyHTTPError(endpoint, model string, statusCode int, body string) error {
	baseErr := fmt.Errorf("ollama chat request failed with status %d: %s", statusCode, body)

	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return &ErrAuth{
			URL:        endpoint,
			StatusCode: statusCode,
			Err:        baseErr,
		}
	case http.StatusTooManyRequests:
		return &ErrRateLimit{
			URL:        endpoint,
			StatusCode: statusCode,
			Err:        baseErr,
		}
	case http.StatusServiceUnavailable:
		// llama-line returns 503 when the queue is at capacity. Treat as a
		// retryable rate-limit so upstream backoff/throttling kicks in.
		return &ErrRateLimit{
			URL:        endpoint,
			StatusCode: statusCode,
			Err:        fmt.Errorf("llama-line broker queue full: %w", baseErr),
		}
	case http.StatusGatewayTimeout:
		return &ErrTimeout{
			URL: endpoint,
			Err: fmt.Errorf("llama-line broker timeout: %w", baseErr),
		}
	case http.StatusBadGateway:
		return &ErrConnection{
			URL: endpoint,
			Err: fmt.Errorf("llama-line broker bad gateway (ollama unreachable): %w", baseErr),
		}
	case http.StatusNotFound:
		if isModelNotFoundMessage(body) {
			return &ErrModelNotFound{
				URL:        endpoint,
				Model:      model,
				StatusCode: statusCode,
				Err:        baseErr,
			}
		}
		return &ErrConnection{
			URL: endpoint,
			Err: baseErr,
		}
	default:
		return &ErrConnection{
			URL: endpoint,
			Err: baseErr,
		}
	}
}

func isModelNotFoundMessage(body string) bool {
	text := strings.ToLower(body)
	if !strings.Contains(text, "model") {
		return false
	}
	return strings.Contains(text, "not found") ||
		strings.Contains(text, "does not exist") ||
		strings.Contains(text, "unknown model")
}

type ollamaChatRequest struct {
	Model    string             `json:"model"`
	Messages []ollamaMessage    `json:"messages"`
	Tools    []ollamaTool       `json:"tools,omitempty"`
	Stream   bool               `json:"stream"`
	Options  ollamaModelOptions `json:"options"`
}

type ollamaModelOptions struct {
	NumCtx int   `json:"num_ctx,omitempty"`
	Think  *bool `json:"think,omitempty"`
}

type ollamaMessage struct {
	Role       Role             `json:"role"`
	Content    string           `json:"content,omitempty"`
	ToolCalls  []ollamaToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type ollamaTool struct {
	Type     string             `json:"type"`
	Function ollamaToolFunction `json:"function"`
}

type ollamaToolFunction struct {
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Parameters  map[string]any      `json:"parameters,omitempty"`
	Arguments   ollamaToolArguments `json:"arguments,omitempty"`
}

// ollamaToolArguments handles Ollama returning arguments as either a JSON
// string (e.g. Llama) or a JSON object (e.g. Gemma4). It also marshals back
// as a raw JSON object so Ollama accepts it in retry messages.
type ollamaToolArguments string

func (a *ollamaToolArguments) UnmarshalJSON(data []byte) error {
	// If it's a JSON string, unwrap it.
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*a = ollamaToolArguments(s)
		return nil
	}
	// Otherwise keep the raw JSON (object/array) as-is.
	*a = ollamaToolArguments(data)
	return nil
}

func (a ollamaToolArguments) MarshalJSON() ([]byte, error) {
	s := string(a)
	if s == "" {
		return []byte(`""`), nil
	}
	// If it's already valid JSON (object/array), emit it raw so Ollama
	// sees an object rather than a quoted string.
	if json.Valid([]byte(s)) && (s[0] == '{' || s[0] == '[') {
		return []byte(s), nil
	}
	return json.Marshal(s)
}

type ollamaToolCall struct {
	Function ollamaToolFunction `json:"function"`
}

type ollamaChatResponse struct {
	Message struct {
		Content   string           `json:"content"`
		ToolCalls []ollamaToolCall `json:"tool_calls,omitempty"`
	} `json:"message"`
	Done            bool   `json:"done"`
	DoneReason      string `json:"done_reason"`
	PromptEvalCount int    `json:"prompt_eval_count"`
	EvalCount       int    `json:"eval_count"`
}

func toOllamaMessages(messages []Message) ([]ollamaMessage, error) {
	if len(messages) == 0 {
		return make([]ollamaMessage, 0), nil
	}
	out := make([]ollamaMessage, 0, len(messages))
	for _, msg := range messages {
		toolCalls, err := toOllamaToolCalls(msg.ToolCalls)
		if err != nil {
			return nil, err
		}
		out = append(out, ollamaMessage{
			Role:       msg.Role,
			Content:    msg.Content,
			ToolCalls:  toolCalls,
			ToolCallID: msg.ToolCallID,
		})
	}
	return out, nil
}

func toOllamaTools(tools []Tool) []ollamaTool {
	if len(tools) == 0 {
		return nil
	}
	out := make([]ollamaTool, 0, len(tools))
	for _, t := range tools {
		out = append(out, ollamaTool{
			Type: "function",
			Function: ollamaToolFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		})
	}
	return out
}

func toOllamaToolCalls(calls []ToolCall) ([]ollamaToolCall, error) {
	if len(calls) == 0 {
		return nil, nil
	}
	out := make([]ollamaToolCall, 0, len(calls))
	for _, c := range calls {
		args, err := json.Marshal(c.Arguments)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tool arguments for %q: %w", c.Name, err)
		}
		out = append(out, ollamaToolCall{
			Function: ollamaToolFunction{
				Name:      c.Name,
				Arguments: ollamaToolArguments(args),
			},
		})
	}
	return out, nil
}

func fromOllamaToolCalls(calls []ollamaToolCall) ([]ToolCall, error) {
	if len(calls) == 0 {
		return nil, nil
	}

	out := make([]ToolCall, 0, len(calls))
	for _, c := range calls {
		parsed := map[string]any{}
		if c.Function.Arguments != "" {
			if err := json.Unmarshal([]byte(string(c.Function.Arguments)), &parsed); err != nil {
				return nil, fmt.Errorf("failed to decode ollama tool arguments for %q: %w", c.Function.Name, err)
			}
		}
		out = append(out, ToolCall{Name: c.Function.Name, Arguments: parsed})
	}
	return out, nil
}
