package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	defaultOpenRouterBaseURL = "https://openrouter.ai"
	openRouterChatPath       = "/api/v1/chat/completions"
	defaultOpenRouterModel   = "openai/gpt-4o-mini"
)

// OpenRouterClient implements Provider via OpenRouter's OpenAI-compatible API.
type OpenRouterClient struct {
	baseURL         string
	apiKey          string
	model           string
	fallbackModels  []string
	reasoningEffort string
	client          *http.Client
}

// NewOpenRouterClient returns an OpenRouter-backed provider.
func NewOpenRouterClient(apiKey, model string) *OpenRouterClient {
	if model == "" {
		model = defaultOpenRouterModel
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

	return &OpenRouterClient{
		baseURL: strings.TrimRight(defaultOpenRouterBaseURL, "/"),
		apiKey:  apiKey,
		model:   model,
		client: &http.Client{
			Timeout:   5 * time.Minute,
			Transport: transport,
		},
	}
}

// WithFallbackModels sets the models OpenRouter tries, in order, when the
// primary model returns an error (outage, rate limit, moderation, context limit).
// OpenRouter does the fallback server-side within the same request.
func (c *OpenRouterClient) WithFallbackModels(models []string) *OpenRouterClient {
	c.fallbackModels = nil
	for _, m := range models {
		if m = strings.TrimSpace(m); m != "" && m != c.model {
			c.fallbackModels = append(c.fallbackModels, m)
		}
	}
	return c
}

// WithReasoningEffort sets OpenRouter's unified reasoning.effort for every
// request ("low" through "max"; empty leaves the model default). The same
// effort applies to fallback models.
func (c *OpenRouterClient) WithReasoningEffort(effort string) *OpenRouterClient {
	c.reasoningEffort = strings.TrimSpace(effort)
	return c
}

func (c *OpenRouterClient) Complete(ctx context.Context, messages []Message, tools []Tool) (*Response, error) {
	reqBody, err := c.buildRequest(messages, tools, false)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(ctx, reqBody)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var chatResp openRouterChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, &ErrMalformedResponse{
			URL: c.chatURL(),
			Err: fmt.Errorf("failed to decode openrouter response: %w", err),
		}
	}

	return c.toResponse(&chatResp)
}

func (c *OpenRouterClient) Stream(ctx context.Context, messages []Message, tools []Tool) (<-chan StreamChunk, error) {
	reqBody, err := c.buildRequest(messages, tools, true)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		_ = resp.Body.Close()
		return nil, &ErrMalformedResponse{
			URL: c.chatURL(),
			Err: fmt.Errorf("expected SSE stream with Content-Type %q, got %q", "text/event-stream", ct),
		}
	}

	ch := make(chan StreamChunk)
	go func() {
		defer close(ch)
		defer func() { _ = resp.Body.Close() }()
		c.parseSSEStream(ctx, resp.Body, ch)
	}()
	return ch, nil
}

func (c *OpenRouterClient) buildRequest(messages []Message, tools []Tool, stream bool) ([]byte, error) {
	openaiMessages := make([]openRouterMessage, 0, len(messages))
	for _, msg := range messages {
		om := openRouterMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
		for _, tc := range msg.ToolCalls {
			argsJSON, _ := json.Marshal(tc.Arguments)
			om.ToolCalls = append(om.ToolCalls, openRouterToolCall{
				ID:   tc.ID,
				Type: "function",
				Function: openRouterToolCallFunction{
					Name:      tc.Name,
					Arguments: string(argsJSON),
				},
			})
		}
		if msg.ToolCallID != "" {
			om.ToolCallID = msg.ToolCallID
		}
		openaiMessages = append(openaiMessages, om)
	}

	openaiTools := make([]openRouterToolDef, 0, len(tools))
	for _, t := range tools {
		openaiTools = append(openaiTools, openRouterToolDef{
			Type: "function",
			Function: openRouterToolFunctionDef{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		})
	}

	req := openRouterChatRequest{
		Model:    c.model,
		Messages: openaiMessages,
		Stream:   stream,
	}
	if len(openaiTools) > 0 {
		req.Tools = openaiTools
	}
	if len(c.fallbackModels) > 0 {
		req.Models = append([]string{c.model}, c.fallbackModels...)
	}
	if c.reasoningEffort != "" {
		req.Reasoning = &openRouterReasoning{Effort: c.reasoningEffort}
	}

	return json.Marshal(req)
}

func (c *OpenRouterClient) doRequest(ctx context.Context, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.chatURL(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create openrouter request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, formatConnectionError(c.chatURL(), err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer func() { _ = resp.Body.Close() }()
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, classifyOpenRouterHTTPError(c.chatURL(), c.model, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return resp, nil
}

func (c *OpenRouterClient) chatURL() string {
	return c.baseURL + openRouterChatPath
}

func (c *OpenRouterClient) toResponse(resp *openRouterChatResponse) (*Response, error) {
	if len(resp.Choices) == 0 {
		return nil, &ErrMalformedResponse{
			URL: c.chatURL(),
			Err: fmt.Errorf("openrouter response contained no choices"),
		}
	}

	if resp.Model != "" && len(c.fallbackModels) > 0 && resp.Model != c.model {
		slog.Default().WithGroup("openrouter").Warn("fallback model served request", "primary", c.model, "served_by", resp.Model)
	}

	choice := resp.Choices[0]
	msg := choice.Message

	toolCalls := make([]ToolCall, 0, len(msg.ToolCalls))
	for _, tc := range msg.ToolCalls {
		var args map[string]any
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			args = map[string]any{}
		}
		toolCalls = append(toolCalls, ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: args,
		})
	}

	return &Response{
		Content:      msg.Content,
		ToolCalls:    toolCalls,
		FinishReason: choice.FinishReason,
		Usage: Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

func (c *OpenRouterClient) parseSSEStream(ctx context.Context, body io.Reader, ch chan<- StreamChunk) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		if data == "[DONE]" {
			select {
			case ch <- StreamChunk{Done: true}:
			case <-ctx.Done():
			}
			return
		}

		var ev openRouterStreamChunk
		if err := json.Unmarshal([]byte(data), &ev); err != nil {
			continue
		}

		if len(ev.Choices) == 0 {
			continue
		}
		choice := ev.Choices[0]

		if choice.Delta.Content != "" {
			select {
			case ch <- StreamChunk{ContentDelta: choice.Delta.Content}:
			case <-ctx.Done():
				return
			}
		}

		for _, tc := range choice.Delta.ToolCalls {
			if tc.Function.Name != "" {
				select {
				case ch <- StreamChunk{ToolCallDelta: &ToolCall{
					ID:   tc.ID,
					Name: tc.Function.Name,
				}}:
				case <-ctx.Done():
					return
				}
			}
		}

		if choice.FinishReason != "" {
			select {
			case ch <- StreamChunk{Done: true}:
			case <-ctx.Done():
			}
			return
		}
	}
}

func classifyOpenRouterHTTPError(endpoint, model string, statusCode int, body string) error {
	baseErr := openRouterAPIError(body, statusCode)

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
	case http.StatusBadRequest:
		return &ErrMalformedResponse{
			URL: endpoint,
			Err: baseErr,
		}
	case http.StatusInternalServerError, 502, 503, 504:
		return &ErrTransient{
			URL:        endpoint,
			StatusCode: statusCode,
			Err:        baseErr,
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

func openRouterAPIError(body string, statusCode int) error {
	var apiErr struct {
		Error struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(body), &apiErr); err == nil && apiErr.Error.Message != "" {
		return fmt.Errorf("openrouter: %s (status %d)", apiErr.Error.Message, statusCode)
	}
	return fmt.Errorf("openrouter error (status %d): %s", statusCode, sanitizeErrorBody(body))
}

func sanitizeErrorBody(body string) string {
	const maxLen = 1024
	if len(body) > maxLen {
		return body[:maxLen] + "..."
	}
	return body
}

// --- OpenAI-compatible JSON types ---

type openRouterChatRequest struct {
	Model     string               `json:"model"`
	Models    []string             `json:"models,omitempty"`
	Messages  []openRouterMessage  `json:"messages"`
	Tools     []openRouterToolDef  `json:"tools,omitempty"`
	Reasoning *openRouterReasoning `json:"reasoning,omitempty"`
	Stream    bool                 `json:"stream,omitempty"`
}

type openRouterReasoning struct {
	Effort string `json:"effort"`
}

type openRouterMessage struct {
	Role       string               `json:"role"`
	Content    string               `json:"content"`
	ToolCalls  []openRouterToolCall `json:"tool_calls,omitempty"`
	ToolCallID string               `json:"tool_call_id,omitempty"`
}

type openRouterToolCall struct {
	ID       string                     `json:"id"`
	Type     string                     `json:"type"`
	Function openRouterToolCallFunction `json:"function"`
}

type openRouterToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openRouterToolFunction = openRouterToolCallFunction

type openRouterTool struct {
	Type     string                 `json:"type"`
	Function openRouterToolFunction `json:"function"`
}

type openRouterToolDef struct {
	Type     string                    `json:"type"`
	Function openRouterToolFunctionDef `json:"function"`
}

type openRouterToolFunctionDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type openRouterChatResponse struct {
	Model   string             `json:"model"`
	Choices []openRouterChoice `json:"choices"`
	Usage   openRouterUsage    `json:"usage"`
}

type openRouterChoice struct {
	Message      openRouterRespMessage `json:"message"`
	FinishReason string                `json:"finish_reason"`
}

type openRouterRespMessage struct {
	Role      string               `json:"role"`
	Content   string               `json:"content"`
	ToolCalls []openRouterToolCall `json:"tool_calls,omitempty"`
}

type openRouterStreamChunk struct {
	Choices []openRouterStreamChoice `json:"choices"`
}

type openRouterStreamChoice struct {
	Delta        openRouterStreamDelta `json:"delta"`
	FinishReason string                `json:"finish_reason"`
}

type openRouterStreamDelta struct {
	Content   string               `json:"content"`
	ToolCalls []openRouterToolCall `json:"tool_calls,omitempty"`
}

type openRouterUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
