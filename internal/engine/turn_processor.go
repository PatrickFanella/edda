package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/llm"
	"git.subcult.tv/subculture-collective/edda/internal/tools"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
)

var (
	// ErrEmptyTurnResponse means the model produced no usable player-facing
	// narrative and no tool-backed state changes for a turn. Callers should treat
	// this as a provider/output failure, not as a successful empty turn.
	ErrEmptyTurnResponse = errors.New("llm returned empty turn response")
	// ErrUnresolvedDurableClaims means the model made durable world-state claims
	// that could not be backed by successful tool calls after one repair attempt.
	ErrUnresolvedDurableClaims = errors.New("llm response contained unresolved durable state claims")
)

// IsLLMOutputError reports whether err represents malformed or unusable model
// output rather than an application/server failure.
func IsLLMOutputError(err error) bool {
	return errors.Is(err, ErrEmptyTurnResponse) || errors.Is(err, ErrUnresolvedDurableClaims)
}

// TurnProcessor handles the tool-call portion of the turn pipeline with
// built-in error recovery. When a tool call fails validation or execution
// it sends the error back to the LLM and retries once. If the retry also
// fails the tool call is skipped; narrative text and all successful tool
// calls from the same response are still returned.
// StatusCallback is an optional function called when the turn processor
// transitions between processing stages. It is safe to set to nil.
type StatusCallback func(api.StatusPayload)

// TurnProcessor handles the tool-call portion of the turn pipeline with
// built-in error recovery and optional status callbacks.
type TurnProcessor struct {
	logger           *slog.Logger
	provider         llm.Provider
	postTurnProvider llm.Provider
	registry         *tools.Registry
	validator        *tools.Validator
	StatusCallback   StatusCallback
}

// SetPostTurnProvider routes durable repair and post-turn extraction calls to a
// separate provider. Passing nil restores fallback to the main turn provider.
func (tp *TurnProcessor) SetPostTurnProvider(provider llm.Provider) {
	tp.postTurnProvider = provider
}

func (tp *TurnProcessor) postTurnLLM() llm.Provider {
	if tp.postTurnProvider != nil {
		return tp.postTurnProvider
	}
	return tp.provider
}

// TurnProcessorOptions configures per-call processor behavior.
type TurnProcessorOptions struct {
	CandidateCallback func(string) error
	StatusCallback    StatusCallback
}

// NewTurnProcessor creates a TurnProcessor backed by the given LLM provider,
// tool registry, and validator.
func NewTurnProcessor(
	provider llm.Provider,
	registry *tools.Registry,
	validator *tools.Validator,
	logger *slog.Logger,
) *TurnProcessor {
	if logger == nil {
		logger = slog.Default()
	}
	return &TurnProcessor{
		logger:    logger,
		provider:  provider,
		registry:  registry,
		validator: validator,
	}
}

// ProcessWithRecovery sends messages to the LLM provider, then executes
// every tool call in the response. For each tool call that fails validation
// or execution it:
//  1. Sends the error back to the LLM (together with the original context)
//     and requests exactly one retry.
//  2. If the retry also fails, skips the tool call and logs the failure at
//     ERROR level with full context.
//
// Only tool calls whose names are present in availableTools are executed;
// tool calls for tools not in that set are treated as validation failures and
// follow the same retry-then-skip path. This prevents hallucinated tool calls
// from being dispatched even if they happen to exist in the registry.
//
// Narrative text from the initial response is always returned regardless of
// tool call outcomes. Successful tool calls – including successful retries –
// are collected in the returned slice.
//
// The function only returns a non-nil error when the initial LLM call itself
// fails; individual tool call failures are handled via retry-then-skip.
func (tp *TurnProcessor) ProcessWithRecovery(
	ctx context.Context,
	messages []llm.Message,
	availableTools []llm.Tool,
) (narrative string, applied []AppliedToolCall, err error) {
	return tp.ProcessWithRecoveryWithOptions(ctx, messages, availableTools, TurnProcessorOptions{StatusCallback: tp.StatusCallback})
}

// ProcessWithRecoveryWithOptions runs ProcessWithRecovery with request-local options.
func (tp *TurnProcessor) ProcessWithRecoveryWithOptions(
	ctx context.Context,
	messages []llm.Message,
	availableTools []llm.Tool,
	opts TurnProcessorOptions,
) (narrative string, applied []AppliedToolCall, err error) {
	local := *tp
	if opts.StatusCallback != nil {
		local.StatusCallback = opts.StatusCallback
	}
	tp = &local
	started := time.Now()
	tp.logger.Info("turn processor started", "messages", len(messages), "tools", len(availableTools), "tool_names", advertisedToolNames(availableTools))
	tp.emitStatus(api.StatusPayload{Stage: "thinking", Description: "Generating response..."})
	resp, err := tp.completeInitialWithRetry(ctx, messages, availableTools)
	if err != nil {
		tp.logger.Error("turn processor initial llm call failed", "duration_ms", time.Since(started).Milliseconds(), "error", err)
		return "", nil, fmt.Errorf("initial LLM call failed: %w", err)
	}

	narrative = resp.Content
	tp.logger.Info("turn processor initial llm response", "tool_calls", len(resp.ToolCalls), "tool_call_names", toolCallNames(resp.ToolCalls), "narrative_len", len(narrative), "finish_reason", resp.FinishReason)
	if len(resp.ToolCalls) == 0 && strings.TrimSpace(narrative) == "" {
		tp.logger.Warn("turn processor received empty initial response; requesting regeneration", "duration_ms", time.Since(started).Milliseconds(), "finish_reason", resp.FinishReason)
		regenResp, regenErr := tp.regenerateEmptyInitialResponse(ctx, messages, availableTools)
		if regenErr != nil {
			tp.logger.Error("turn processor empty-response regeneration failed", "duration_ms", time.Since(started).Milliseconds(), "error", regenErr)
			return "", nil, ErrEmptyTurnResponse
		}
		resp = regenResp
		narrative = resp.Content
		tp.logger.Info("turn processor regenerated initial llm response", "tool_calls", len(resp.ToolCalls), "tool_call_names", toolCallNames(resp.ToolCalls), "narrative_len", len(narrative), "finish_reason", resp.FinishReason)
	}

	if opts.CandidateCallback != nil {
		if err := opts.CandidateCallback(narrative); err != nil {
			return "", nil, err
		}
	}
	allowed := make(map[string]struct{}, len(availableTools))
	for _, t := range availableTools {
		allowed[t.Name] = struct{}{}
	}
	if len(resp.ToolCalls) == 0 {
		if strings.TrimSpace(narrative) == "" {
			tp.logger.Error("turn processor received empty response without tool calls", "duration_ms", time.Since(started).Milliseconds(), "finish_reason", resp.FinishReason)
			return "", nil, ErrEmptyTurnResponse
		}
		narrative, applied, err = tp.finalizeResponseState(ctx, messages, availableTools, narrative, applied, nil, nil)
		if err != nil {
			return "", nil, err
		}
		tp.logger.Warn("turn processor completed without tool calls", "duration_ms", time.Since(started).Milliseconds(), "advertised_tools", len(availableTools), "advertised_tool_names", advertisedToolNames(availableTools))
		return narrative, applied, nil
	}

	tp.emitStatus(api.StatusPayload{Stage: "tools", Description: "Executing tool calls..."})
	unresolvedDurableIssues := []DurableClaimIssue{}
	unresolvedDurableRequirements := []durableRequirement{}
	assistantContent := resp.Content
	for _, tc := range resp.ToolCalls {
		if err := ctx.Err(); err != nil {
			return "", applied, err
		}
		tp.emitStatus(api.StatusPayload{Stage: "tool_execution", Tool: tc.Name, Description: fmt.Sprintf("Executing %s...", tc.Name)})
		if err := ctx.Err(); err != nil {
			return "", applied, err
		}
		tp.logger.Debug("attempting tool call", "tool", tc.Name, "tool_call_id", tc.ID)
		result, execErr := tp.attemptToolCall(ctx, tc, allowed)
		if execErr == nil {
			if atc, encErr := buildAppliedToolCall(tc, result); encErr != nil {
				tp.logger.Error("failed to encode applied tool call; skipping",
					"tool", tc.Name,
					"tool_call_id", tc.ID,
					"error", encErr.Error(),
				)
			} else {
				applied = append(applied, atc)
				tp.logger.Debug("tool call applied", "tool", tc.Name, "tool_call_id", tc.ID)
			}
			continue
		}

		if errors.Is(execErr, ErrTurnPersistence) {
			return "", applied, execErr
		}
		tp.logger.Warn("tool call failed; requesting retry", "tool", tc.Name, "tool_call_id", tc.ID, "error", execErr.Error())
		retryTC, retryLLMErr := tp.requestRetry(ctx, tc, execErr, messages, assistantContent, availableTools)
		if retryLLMErr != nil {
			tp.logger.Error("tool call failed and retry LLM call also failed; skipping",
				"tool", tc.Name,
				"tool_call_id", tc.ID,
				"initial_error", execErr.Error(),
				"retry_llm_error", retryLLMErr.Error(),
			)
			unresolvedDurableIssues = appendFailedDurableIssue(unresolvedDurableIssues, tc)
			unresolvedDurableRequirements = appendFailedDurableRequirement(unresolvedDurableRequirements, tc, applied)
			continue
		}

		tp.logger.Debug("retry tool call received", "tool", retryTC.Name, "tool_call_id", retryTC.ID)
		retryResult, retryExecErr := tp.attemptToolCall(ctx, retryTC, allowed)
		if retryExecErr != nil {
			if errors.Is(retryExecErr, ErrTurnPersistence) {
				return "", applied, retryExecErr
			}
			tp.logger.Error("tool call failed after retry; skipping",
				"tool", tc.Name,
				"tool_call_id", tc.ID,
				"retry_tool_call_id", retryTC.ID,
				"initial_error", execErr.Error(),
				"retry_error", retryExecErr.Error(),
				"retry_arguments", retryTC.Arguments,
			)
			unresolvedDurableIssues = appendFailedDurableIssue(unresolvedDurableIssues, retryTC)
			unresolvedDurableIssues = appendFailedDurableIssue(unresolvedDurableIssues, tc)
			if _, ok := durableIssueForFailedToolCall(retryTC); ok {
				unresolvedDurableRequirements = appendFailedDurableRequirement(unresolvedDurableRequirements, retryTC, applied)
			} else {
				unresolvedDurableRequirements = appendFailedDurableRequirement(unresolvedDurableRequirements, tc, applied)
			}
			continue
		}

		if atc, encErr := buildAppliedToolCall(retryTC, retryResult); encErr != nil {
			tp.logger.Error("failed to encode applied tool call after retry; skipping",
				"tool", retryTC.Name,
				"tool_call_id", retryTC.ID,
				"error", encErr.Error(),
			)
		} else {
			applied = append(applied, atc)
			if issue, ok := durableIssueForFailedToolCall(tc); ok && !appliedToolSatisfiesDurableIssue(atc, issue.Kind) {
				unresolvedDurableIssues = appendUniqueDurableIssue(unresolvedDurableIssues, issue)
				unresolvedDurableRequirements = appendDurableRequirement(unresolvedDurableRequirements, issue, applied)
			}
			tp.logger.Debug("tool call applied after retry", "tool", retryTC.Name, "tool_call_id", retryTC.ID)
		}
	}

	// If the model returned tool calls but no narrative text, send the tool
	// results back so it can generate the narrative based on outcomes.
	if narrative == "" && len(applied) > 0 {
		tp.logger.Debug("no narrative after tool calls; issuing continuation call", "applied_tool_calls", len(applied))
		narrative, err = tp.requestContinuation(ctx, messages, resp, applied, availableTools)
		if err != nil {
			tp.logger.Error("continuation call failed", "error", err)
			// Non-fatal: return what we have (empty narrative + applied tools).
		}
	}

	narrative, applied, err = tp.finalizeResponseState(ctx, messages, availableTools, narrative, applied, unresolvedDurableIssues, unresolvedDurableRequirements)
	if err != nil {
		return "", nil, err
	}
	if strings.TrimSpace(narrative) == "" {
		if fallback := fallbackNarrativeForAppliedTools(applied); fallback != "" {
			tp.logger.Warn("using fallback narrative for applied tool calls", "applied_tool_calls", len(applied), "applied_tool_names", appliedToolNames(applied))
			narrative = fallback
		} else {
			tp.logger.Error("turn processor completed with empty narrative", "duration_ms", time.Since(started).Milliseconds(), "applied_tool_calls", len(applied), "applied_tool_names", appliedToolNames(applied))
			return "", nil, ErrEmptyTurnResponse
		}
	}

	tp.logger.Info("turn processor completed", "duration_ms", time.Since(started).Milliseconds(), "applied_tool_calls", len(applied), "applied_tool_names", appliedToolNames(applied), "narrative_len", len(narrative))
	return narrative, applied, nil
}

func fallbackNarrativeForAppliedTools(applied []AppliedToolCall) string {
	if len(applied) == 0 {
		return ""
	}
	last := applied[len(applied)-1]
	switch last.Tool {
	case "move_player":
		if name := stringFieldFromJSON(last.Result, "name"); name != "" {
			return fmt.Sprintf("You move to %s.", name)
		}
		return "You move to the connected location."
	case "create_location":
		if boolFieldFromJSON(last.Result, "move_player_here") {
			if name := stringFieldFromJSON(last.Result, "name"); name != "" {
				return fmt.Sprintf("You enter %s.", name)
			}
			return "You enter the newly revealed location."
		}
		return "A new location is revealed."
	case "update_player_hp":
		return "Your health changes."
	case "update_player_status":
		if status := stringFieldFromJSON(last.Result, "status"); status != "" {
			return fmt.Sprintf("Your status is now %s.", status)
		}
		return "Your status changes."
	case "add_item", "create_item":
		if name := stringFieldFromJSON(last.Result, "name"); name != "" {
			return fmt.Sprintf("You gain %s.", name)
		}
		return "Your inventory is updated."
	case "remove_item", "modify_item", "update_item":
		return "Your inventory is updated."
	case "create_quest", "update_quest", "complete_objective":
		return "Your quest journal is updated."
	case "establish_fact", "revise_fact":
		return "You record the new information."
	case "initiate_combat":
		return "Combat begins."
	case "combat_round", "apply_damage", "apply_condition":
		return "The combat state changes."
	case "resolve_combat":
		return "Combat is resolved."
	default:
		return "The action takes effect."
	}
}

func stringFieldFromJSON(data json.RawMessage, key string) string {
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		return ""
	}
	value, _ := obj[key].(string)
	return value
}

func boolFieldFromJSON(data json.RawMessage, key string) bool {
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		return false
	}
	value, _ := obj[key].(bool)
	return value
}

func (tp *TurnProcessor) regenerateEmptyInitialResponse(ctx context.Context, messages []llm.Message, availableTools []llm.Tool) (*llm.Response, error) {
	regenMessages := make([]llm.Message, len(messages), len(messages)+1)
	copy(regenMessages, messages)
	regenMessages = append(regenMessages, llm.Message{
		Role:    llm.RoleUser,
		Content: "Your previous response was empty. Regenerate the turn response now. Resolve the player's action with either player-facing narrative text, valid tool calls, or both. Do not return an empty response.",
	})
	resp, err := tp.provider.Complete(ctx, regenMessages, availableTools)
	if err != nil {
		return nil, fmt.Errorf("empty-response regeneration LLM call: %w", err)
	}
	if resp == nil || (strings.TrimSpace(resp.Content) == "" && len(resp.ToolCalls) == 0) {
		return nil, ErrEmptyTurnResponse
	}
	return resp, nil
}

func (tp *TurnProcessor) completeInitialWithRetry(ctx context.Context, messages []llm.Message, availableTools []llm.Tool) (*llm.Response, error) {
	const maxAttempts = 2
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		resp, err := tp.provider.Complete(ctx, messages, availableTools)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if attempt == maxAttempts || !isRetryableInitialLLMError(ctx, err) {
			return nil, err
		}
		tp.emitStatus(api.StatusPayload{Stage: "retrying", Description: "Retrying initial response after transient provider failure..."})
		if err := waitForRetryDelay(ctx, err); err != nil {
			return nil, err
		}
	}
	return nil, lastErr
}

const maxInitialRetryAfter = 2 * time.Second

func waitForRetryDelay(ctx context.Context, err error) error {
	if delay := retryDelayForInitialError(err); delay > 0 {
		select {
		case <-time.After(delay):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func retryDelayForInitialError(err error) time.Duration {
	var rl *llm.ErrRateLimit
	if errors.As(err, &rl) && rl.HasRetryAfter && rl.RetryAfter > 0 {
		if rl.RetryAfter > maxInitialRetryAfter {
			return maxInitialRetryAfter
		}
		return rl.RetryAfter
	}
	return 25 * time.Millisecond
}

func isRetryableInitialLLMError(ctx context.Context, err error) bool {
	var (
		transient *llm.ErrTransient
		timeout   *llm.ErrTimeout
		rateLimit *llm.ErrRateLimit
		conn      *llm.ErrConnection
	)
	switch {
	case errors.As(err, &transient):
		return true
	case errors.As(err, &timeout):
		return ctx.Err() == nil
	case errors.As(err, &rateLimit):
		return true
	case errors.As(err, &conn):
		msg := strings.ToLower(err.Error())
		for _, fragment := range []string{"transport", "timeout", "timed out", "reset", "refused", "temporar", "connection refused", "connection reset", "bad gateway", "ollama_unavailable", "ollama unreachable", "unreachable", "502", "503", "504"} {
			if strings.Contains(msg, fragment) {
				return true
			}
		}
	}
	return false
}

func appendFailedDurableIssue(issues []DurableClaimIssue, tc llm.ToolCall) []DurableClaimIssue {
	issue, ok := durableIssueForFailedToolCall(tc)
	if !ok {
		return issues
	}
	return appendUniqueDurableIssue(issues, issue)
}

type durableRequirement struct {
	Kind          DurableClaimKind
	Message       string
	RequiredCount int
}

func appendFailedDurableRequirement(requirements []durableRequirement, tc llm.ToolCall, applied []AppliedToolCall) []durableRequirement {
	issue, ok := durableIssueForFailedToolCall(tc)
	if !ok {
		return requirements
	}
	return appendDurableRequirement(requirements, issue, applied)
}

func appendDurableRequirement(requirements []durableRequirement, issue DurableClaimIssue, applied []AppliedToolCall) []durableRequirement {
	requiredCount := countAppliedSatisfyingDurableKind(applied, issue.Kind) + 1
	for _, existing := range requirements {
		if existing.Kind == issue.Kind && existing.RequiredCount >= requiredCount {
			requiredCount = existing.RequiredCount + 1
		}
	}
	return append(requirements, durableRequirement{Kind: issue.Kind, Message: issue.Message, RequiredCount: requiredCount})
}

func countAppliedSatisfyingDurableKind(applied []AppliedToolCall, kind DurableClaimKind) int {
	count := 0
	for _, call := range applied {
		if appliedToolSatisfiesDurableIssue(call, kind) {
			count++
		}
	}
	return count
}

func appendDurableIssues(issues []DurableClaimIssue, extra ...DurableClaimIssue) []DurableClaimIssue {
	for _, issue := range extra {
		issues = appendUniqueDurableIssue(issues, issue)
	}
	return issues
}

func appendUniqueDurableIssue(issues []DurableClaimIssue, issue DurableClaimIssue) []DurableClaimIssue {
	for _, existing := range issues {
		if existing.Kind == issue.Kind && existing.Message == issue.Message {
			return issues
		}
	}
	return append(issues, issue)
}

func durableIssueForFailedToolCall(tc llm.ToolCall) (DurableClaimIssue, bool) {
	switch tc.Name {
	case "move_player":
		return DurableClaimIssue{Kind: DurableClaimMovement, Message: "movement tool failed before durable state was updated"}, true
	case "create_location":
		if boolArg(tc.Arguments, "move_player_here") {
			return DurableClaimIssue{Kind: DurableClaimMovement, Message: "create_location move_player_here failed before player location was updated"}, true
		}
	case "create_quest", "update_quest", "complete_objective":
		return DurableClaimIssue{Kind: DurableClaimQuest, Message: "quest tool failed before durable state was updated"}, true
	case "establish_fact", "revise_fact":
		return DurableClaimIssue{Kind: DurableClaimFact, Message: "fact tool failed before durable state was updated"}, true
	}
	return DurableClaimIssue{}, false
}

func boolArg(args map[string]any, key string) bool {
	v, ok := args[key]
	if !ok {
		return false
	}
	b, _ := v.(bool)
	return b
}

func appliedToolSatisfiesDurableIssue(call AppliedToolCall, kind DurableClaimKind) bool {
	switch kind {
	case DurableClaimMovement:
		if call.Tool == "move_player" {
			return true
		}
		return call.Tool == "create_location" && hasMovePlayerHere(call.Result)
	case DurableClaimQuest:
		return call.Tool == "create_quest" || call.Tool == "update_quest" || call.Tool == "complete_objective"
	case DurableClaimFact:
		return call.Tool == "establish_fact" || call.Tool == "revise_fact"
	default:
		return false
	}
}

func (tp *TurnProcessor) repairDurableClaims(ctx context.Context, messages []llm.Message, availableTools []llm.Tool, narrative string, applied []AppliedToolCall, issues []DurableClaimIssue, requirements []durableRequirement) (string, []AppliedToolCall, error) {
	repairTools := durableRepairTools(availableTools, issues)
	if len(repairTools) == 0 {
		return "", nil, fmt.Errorf("%w: no repair tools available for issues %v", ErrUnresolvedDurableClaims, issues)
	}

	repairMessages := make([]llm.Message, len(messages), len(messages)+2)
	copy(repairMessages, messages)
	repairMessages = append(repairMessages, llm.Message{Role: llm.RoleAssistant, Content: narrative})
	repairMessages = append(repairMessages, llm.Message{Role: llm.RoleUser, Content: durableRepairPrompt(narrative, applied, issues)})

	resp, err := tp.postTurnLLM().Complete(ctx, repairMessages, repairTools)
	if err != nil {
		tp.logger.Error("durable-claim repair call failed", "error", err)
		return "", nil, fmt.Errorf("%w: repair call failed: %v", ErrUnresolvedDurableClaims, err)
	}

	allowed := toolNameSet(repairTools)
	repairApplied := make([]AppliedToolCall, 0, len(resp.ToolCalls))
	for _, tc := range resp.ToolCalls {
		if result, execErr := tp.attemptToolCall(ctx, tc, allowed); execErr == nil {
			if atc, encErr := buildAppliedToolCall(tc, result); encErr == nil {
				applied = append(applied, atc)
				repairApplied = append(repairApplied, atc)
			}
		} else {
			if errors.Is(execErr, ErrTurnPersistence) {
				return "", applied, execErr
			}
			tp.logger.Warn("durable-claim repair tool failed", "tool", tc.Name, "error", execErr)
		}
	}

	repairedNarrative := resp.Content
	if repairedNarrative == "" && len(repairApplied) > 0 {
		continued, contErr := tp.requestContinuation(ctx, repairMessages, resp, repairApplied, repairTools)
		if contErr != nil {
			tp.logger.Error("durable-claim repair continuation failed", "error", contErr)
		} else {
			repairedNarrative = continued
		}
	}
	if repairedNarrative == "" {
		return "", nil, fmt.Errorf("%w: repair produced empty narrative", ErrUnresolvedDurableClaims)
	}
	if !durableRequirementsSatisfied(applied, requirements) {
		return "", nil, fmt.Errorf("%w: repair did not satisfy required durable tool calls", ErrUnresolvedDurableClaims)
	}
	if len(AuditDurableClaims(repairedNarrative, applied, advertisedToolNames(repairTools))) > 0 {
		return "", nil, fmt.Errorf("%w: repair narrative still contains unbacked claims", ErrUnresolvedDurableClaims)
	}
	return repairedNarrative, applied, nil
}

func durableRequirementsSatisfied(applied []AppliedToolCall, requirements []durableRequirement) bool {
	for _, requirement := range requirements {
		if countAppliedSatisfyingDurableKind(applied, requirement.Kind) < requirement.RequiredCount {
			return false
		}
	}
	return true
}

func durableRepairPrompt(narrative string, applied []AppliedToolCall, issues []DurableClaimIssue) string {
	var issueLines strings.Builder
	for _, issue := range issues {
		fmt.Fprintf(&issueLines, "- %s: %s\n", issue.Kind, issue.Message)
	}
	return fmt.Sprintf("Durable state audit found unbacked claims in the previous narrative:\n%s\nAlready applied tools: %v\n\nCall only the missing durable state tools now, then provide a corrected narrative. Do not invent a fallback, provisional, or pretend resolution.\n\nPrevious narrative: %s", issueLines.String(), appliedToolNames(applied), narrative)
}

func durableRepairTools(availableTools []llm.Tool, issues []DurableClaimIssue) []llm.Tool {
	wanted := map[string]struct{}{}
	for _, issue := range issues {
		switch issue.Kind {
		case DurableClaimMovement:
			wanted["move_player"] = struct{}{}
			wanted["create_location"] = struct{}{}
		case DurableClaimQuest:
			wanted["create_quest"] = struct{}{}
			wanted["update_quest"] = struct{}{}
			wanted["complete_objective"] = struct{}{}
		case DurableClaimFact:
			wanted["establish_fact"] = struct{}{}
			wanted["revise_fact"] = struct{}{}
		case DurableClaimInventoryCreated:
			wanted["add_item"] = struct{}{}
			wanted["create_item"] = struct{}{}
		case DurableClaimInventoryRemoved:
			wanted["remove_item"] = struct{}{}
		case DurableClaimInventoryUpdated:
			wanted["modify_item"] = struct{}{}
			wanted["update_item"] = struct{}{}
		case DurableClaimCombatStarted:
			wanted["initiate_combat"] = struct{}{}
		case DurableClaimCombatResolved:
			wanted["resolve_combat"] = struct{}{}
		}
	}
	filtered := make([]llm.Tool, 0, len(availableTools))
	for _, tool := range availableTools {
		if _, ok := wanted[tool.Name]; ok {
			filtered = append(filtered, tool)
		}
	}
	return filtered
}

func (tp *TurnProcessor) extractDurableLoreFacts(ctx context.Context, messages []llm.Message, availableTools []llm.Tool, narrative string, applied []AppliedToolCall) []AppliedToolCall {
	// Kept for backward compatibility with existing callers; delegates to the
	// unified extractDurableState.
	return tp.extractDurableState(ctx, messages, availableTools, narrative, applied)
}

func durableLoreExtractionTools(availableTools []llm.Tool) []llm.Tool {
	for _, tool := range availableTools {
		if tool.Name == "establish_fact" {
			return []llm.Tool{tool}
		}
	}
	return nil
}

func (tp *TurnProcessor) finalizeResponseState(ctx context.Context, messages []llm.Message, availableTools []llm.Tool, narrative string, applied []AppliedToolCall, unresolvedIssues []DurableClaimIssue, unresolvedRequirements []durableRequirement) (string, []AppliedToolCall, error) {
	if issues := appendDurableIssues(AuditDurableClaims(narrative, applied, advertisedToolNames(availableTools)), unresolvedIssues...); len(issues) > 0 {
		var err error
		narrative, applied, err = tp.repairDurableClaims(ctx, messages, availableTools, narrative, applied, issues, unresolvedRequirements)
		if err != nil {
			if errors.Is(err, ErrTurnPersistence) {
				return "", applied, err
			}
			fallback, fallbackErr := tp.rewriteUnsafeDurableClaimsAsProvisional(ctx, messages, narrative, applied, issues, err)
			if fallbackErr != nil {
				return "", nil, err
			}
			return fallback, applied, nil
		}
	}
	applied = tp.extractDurableState(ctx, messages, availableTools, narrative, applied)
	if err := turnPersistenceFailure(ctx); err != nil {
		return "", applied, err
	}
	return narrative, applied, nil
}

func (tp *TurnProcessor) rewriteUnsafeDurableClaimsAsProvisional(ctx context.Context, messages []llm.Message, narrative string, applied []AppliedToolCall, issues []DurableClaimIssue, repairErr error) (string, error) {
	rewriteMessages := make([]llm.Message, len(messages), len(messages)+2)
	copy(rewriteMessages, messages)
	rewriteMessages = append(rewriteMessages, llm.Message{Role: llm.RoleAssistant, Content: narrative})
	rewriteMessages = append(rewriteMessages, llm.Message{Role: llm.RoleUser, Content: provisionalDurableRewritePrompt(narrative, applied, issues, repairErr)})

	resp, err := tp.postTurnLLM().Complete(ctx, rewriteMessages, nil)
	if err != nil {
		tp.logger.Warn("durable-claim provisional rewrite failed", "error", err)
		return "", err
	}
	rewritten := strings.TrimSpace(resp.Content)
	if rewritten == "" {
		return "", ErrEmptyTurnResponse
	}
	if remaining := AuditDurableClaims(rewritten, applied, nil); len(remaining) > 0 {
		tp.logger.Warn("durable-claim provisional rewrite still has unbacked claims", "issues", remaining)
		return "", ErrUnresolvedDurableClaims
	}
	tp.logger.Info("durable-claim repair fell back to provisional narrative", "issues", len(issues), "repair_error", repairErr)
	return rewritten, nil
}

func provisionalDurableRewritePrompt(narrative string, applied []AppliedToolCall, issues []DurableClaimIssue, repairErr error) string {
	var issueLines strings.Builder
	for _, issue := range issues {
		fmt.Fprintf(&issueLines, "- %s: %s\n", issue.Kind, issue.Message)
	}
	return fmt.Sprintf("The previous narrative made durable state claims that could not be saved by tools. Rewrite it as player-facing narrative that preserves the moment but makes those state changes provisional, blocked, attempted, noticed, or inconclusive. Do NOT claim movement, quest updates, facts learned, inventory changes, HP/status changes, or combat resolution unless already backed by applied tools. Return ONLY the corrected narrative; do not call tools and do not explain the correction.\n\nUnbacked durable claims:\n%s\nAlready applied tools: %v\nRepair error: %v\n\nPrevious narrative: %s", issueLines.String(), appliedToolNames(applied), repairErr, narrative)
}

// extractDurableState runs two sequential post-turn extraction passes:
//  1. Lore extraction (establish_fact only) — single-purpose prompt.
//  2. Quest extraction (create_quest / update_quest / complete_objective only) — separate prompt.
//
// Splitting them avoids the model ignoring the second task, which happened
// with weaker models when both were combined in one prompt.
func (tp *TurnProcessor) extractDurableState(ctx context.Context, messages []llm.Message, availableTools []llm.Tool, narrative string, applied []AppliedToolCall) []AppliedToolCall {
	if strings.TrimSpace(narrative) == "" {
		return applied
	}

	hasFact := countAppliedSatisfyingDurableKind(applied, DurableClaimFact) > 0
	hasQuest := countAppliedSatisfyingDurableKind(applied, DurableClaimQuest) > 0

	// Pass 1: lore facts (single-purpose — establish_fact only).
	if !hasFact && shouldRunLoreExtraction(narrative) {
		factTools := durableLoreExtractionTools(availableTools)
		if len(factTools) > 0 {
			extractionMessages := make([]llm.Message, len(messages), len(messages)+2)
			copy(extractionMessages, messages)
			extractionMessages = append(extractionMessages, llm.Message{Role: llm.RoleAssistant, Content: narrative})
			extractionMessages = append(extractionMessages, llm.Message{Role: llm.RoleUser, Content: durableLoreExtractionPromptV2(narrative)})

			applied = tp.runPostTurnExtractionTools(ctx, extractionMessages, factTools, applied, "lore extraction", 1, func(name string) bool {
				return name == "establish_fact"
			})
		}
	}

	if turnPersistenceFailure(ctx) != nil {
		return applied
	}
	// Pass 2: quest goals (single-purpose — create_quest / update_quest / complete_objective only).
	if !hasQuest {
		questTools := questOnlyExtractionTools(availableTools)
		if len(questTools) > 0 {
			extractionMessages := make([]llm.Message, len(messages), len(messages)+2)
			copy(extractionMessages, messages)
			extractionMessages = append(extractionMessages, llm.Message{Role: llm.RoleAssistant, Content: narrative})
			extractionMessages = append(extractionMessages, llm.Message{Role: llm.RoleUser, Content: questOnlyExtractionPrompt(narrative)})

			applied = tp.runPostTurnExtractionTools(ctx, extractionMessages, questTools, applied, "quest extraction", 0, func(name string) bool {
				return name == "create_quest" || name == "update_quest" || name == "complete_objective"
			})
		}
	}

	return applied
}

func questOnlyExtractionTools(availableTools []llm.Tool) []llm.Tool {
	wanted := map[string]struct{}{}
	for _, tool := range availableTools {
		switch tool.Name {
		case "create_quest", "update_quest", "complete_objective":
			wanted[tool.Name] = struct{}{}
		}
	}
	filtered := make([]llm.Tool, 0, len(wanted))
	for _, tool := range availableTools {
		if _, ok := wanted[tool.Name]; ok {
			filtered = append(filtered, tool)
		}
	}
	return filtered
}

func durableLoreExtractionPromptV2(narrative string) string {
	return fmt.Sprintf("Extract durable canonical facts from the narrative below. Call establish_fact for at most ONE important NEW fact the player explicitly learned, discovered, confirmed, was told, or had revealed. Use categories: lore, history, hazard, faction, location, relic, mechanism, or magic. Skip facts already listed in the system message's World Facts section. Skip vague atmosphere, restatements, sensory detail, mood, symbolism, and one-off descriptions. If no important new durable lore was explicitly revealed, call no tools and return no prose.\n\nNarrative: %s", narrative)
}

func shouldRunLoreExtraction(narrative string) bool {
	lower := strings.ToLower(narrative)
	for _, cue := range []string{
		"you learn", "you now know", "you confirm", "you discover", "you find out",
		"reveals", "revealed", "tells you", "says", "explains", "confirms",
		"is called", "called it", "known as", "symbolizes", "signifies", "means that",
	} {
		if strings.Contains(lower, cue) {
			return true
		}
	}
	return false
}

func questOnlyExtractionPrompt(narrative string) string {
	return fmt.Sprintf("Extract quest progress from the narrative below. Use ONLY quest_id and objective_id values explicitly listed in the system message's Active Quests section. Never invent IDs. If the narrative completes a listed objective, call complete_objective with that exact quest_id and objective_id. If the narrative advances a listed active quest without completing an objective, call update_quest with that exact quest_id. If the narrative establishes a concrete new multi-step goal that is NOT covered by any active quest, call create_quest with a short title, 1-sentence description, quest_type \"short_term\", and 1-3 ordered objectives. Do NOT create quests for vague atmosphere, one-off actions, or mere lore. If no listed quest/objective matches and no clearly new quest-shaped goal exists, call no tools and return no prose.\n\nNarrative: %s", narrative)
}

func (tp *TurnProcessor) runPostTurnExtractionTools(ctx context.Context, messages []llm.Message, extractionTools []llm.Tool, applied []AppliedToolCall, label string, maxAccepted int, acceptTool func(string) bool) []AppliedToolCall {
	resp, err := tp.postTurnLLM().Complete(ctx, messages, extractionTools)
	if err != nil {
		tp.logger.Warn(label+" failed", "error", err)
		return applied
	}
	if len(resp.ToolCalls) == 0 {
		return applied
	}

	allowed := toolNameSet(extractionTools)
	accepted := 0
	for _, tc := range resp.ToolCalls {
		if maxAccepted > 0 && accepted >= maxAccepted {
			break
		}
		if !acceptTool(tc.Name) {
			continue
		}
		result, execErr := tp.attemptToolCall(ctx, tc, allowed)
		if execErr != nil {
			if errors.Is(execErr, ErrTurnPersistence) {
				return applied
			}
			tp.logger.Warn(label+" tool failed", "tool", tc.Name, "error", execErr)
			continue
		}
		atc, encErr := buildAppliedToolCall(tc, result)
		if encErr != nil {
			tp.logger.Warn(label+" encode failed", "tool", tc.Name, "error", encErr)
			continue
		}
		applied = append(applied, atc)
		accepted++
	}
	return applied
}

func toolNameSet(tools []llm.Tool) map[string]struct{} {
	allowed := make(map[string]struct{}, len(tools))
	for _, t := range tools {
		allowed[t.Name] = struct{}{}
	}
	return allowed
}

func advertisedToolNames(tools []llm.Tool) []string {
	names := make([]string, 0, len(tools))
	for _, tool := range tools {
		names = append(names, tool.Name)
	}
	return names
}

func toolCallNames(calls []llm.ToolCall) []string {
	names := make([]string, 0, len(calls))
	for _, call := range calls {
		names = append(names, call.Name)
	}
	return names
}

func appliedToolNames(calls []AppliedToolCall) []string {
	names := make([]string, 0, len(calls))
	for _, call := range calls {
		names = append(names, call.Tool)
	}
	return names
}

// attemptToolCall validates and executes a single tool call. The allowed set
// is derived from the tools advertised to the LLM; tool calls whose names are
// not in the set are rejected as hallucinations before registry lookup.
// Both validation and execution errors are returned as-is so callers can
// include them in log messages or retry prompts.
func (tp *TurnProcessor) attemptToolCall(ctx context.Context, tc llm.ToolCall, allowed map[string]struct{}) (*tools.ToolResult, error) {
	if _, ok := allowed[tc.Name]; !ok {
		return nil, fmt.Errorf("validation: tool %q was not in the advertised tool list", tc.Name)
	}
	if err := tp.validator.ValidatePreExecution(tc); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}
	result, err := tp.registry.Execute(ctx, tc.Name, tc.Arguments)
	if err != nil {
		return nil, fmt.Errorf("execution: %w", err)
	}
	tp.logger.Debug("tool execution succeeded", "tool", tc.Name, "tool_call_id", tc.ID)
	return result, nil
}

// requestRetry builds a retry conversation and calls the LLM once. The
// returned ToolCall is the corrected invocation suggested by the model.
//
// The retry context is:
//  1. The original conversation messages.
//  2. An assistant turn containing only the failed tool call.
//  3. A tool-result turn carrying the error message.
func (tp *TurnProcessor) requestRetry(
	ctx context.Context,
	failedTC llm.ToolCall,
	execErr error,
	originalMessages []llm.Message,
	assistantContent string,
	availableTools []llm.Tool,
) (llm.ToolCall, error) {
	retryStarted := time.Now()
	retryMessages := make([]llm.Message, len(originalMessages), len(originalMessages)+2)
	copy(retryMessages, originalMessages)

	retryMessages = append(retryMessages, llm.Message{
		Role:      llm.RoleAssistant,
		Content:   assistantContent,
		ToolCalls: []llm.ToolCall{failedTC},
	})
	retryMessages = append(retryMessages, llm.Message{
		Role:       llm.RoleTool,
		Content:    fmt.Sprintf("Error: %s. Please retry with corrected arguments.", execErr.Error()),
		ToolCallID: failedTC.ID,
	})

	tp.logger.Debug("issuing retry llm call", "tool", failedTC.Name, "tool_call_id", failedTC.ID, "messages", len(retryMessages))
	retryResp, err := tp.provider.Complete(ctx, retryMessages, availableTools)
	if err != nil {
		tp.logger.Error("retry llm call failed", "tool", failedTC.Name, "tool_call_id", failedTC.ID, "duration_ms", time.Since(retryStarted).Milliseconds(), "error", err)
		return llm.ToolCall{}, fmt.Errorf("retry LLM call: %w", err)
	}

	for _, tc := range retryResp.ToolCalls {
		if tc.Name == failedTC.Name {
			tp.logger.Debug("retry llm call completed", "tool", failedTC.Name, "tool_call_id", failedTC.ID, "duration_ms", time.Since(retryStarted).Milliseconds(), "returned_tool_calls", len(retryResp.ToolCalls))
			return tc, nil
		}
	}
	if len(retryResp.ToolCalls) > 0 {
		tp.logger.Debug("retry llm call completed with fallback tool", "tool", failedTC.Name, "tool_call_id", failedTC.ID, "duration_ms", time.Since(retryStarted).Milliseconds(), "returned_tool_calls", len(retryResp.ToolCalls))
		return retryResp.ToolCalls[0], nil
	}

	return llm.ToolCall{}, fmt.Errorf(
		"LLM returned no tool calls in retry response for tool %q (tool_call_id=%s)",
		failedTC.Name,
		failedTC.ID,
	)
}

// requestContinuation sends tool results back to the LLM so it can produce
// narrative text based on the outcomes. This handles models (like Gemma) that
// emit tool calls without accompanying text.
func (tp *TurnProcessor) requestContinuation(
	ctx context.Context,
	originalMessages []llm.Message,
	initialResp *llm.Response,
	applied []AppliedToolCall,
	availableTools []llm.Tool,
) (string, error) {
	contMessages := make([]llm.Message, len(originalMessages), len(originalMessages)+1+len(applied))
	copy(contMessages, originalMessages)

	// Add the assistant message with tool calls.
	contMessages = append(contMessages, llm.Message{
		Role:      llm.RoleAssistant,
		Content:   initialResp.Content,
		ToolCalls: initialResp.ToolCalls,
	})

	// Answer every tool call in the assistant message, under the call's own ID:
	// Anthropic and OpenAI reject a tool result whose ID matches no call. An
	// applied result answers the first unanswered call with the same tool name
	// (a retried call keeps its tool name but gets a new ID).
	used := make([]bool, len(applied))
	for _, tc := range initialResp.ToolCalls {
		content := `{"error":"tool call was not applied"}`
		for i, atc := range applied {
			if !used[i] && atc.Tool == tc.Name {
				used[i] = true
				content = string(atc.Result)
				break
			}
		}
		id := tc.ID
		if id == "" {
			id = tc.Name
		}
		contMessages = append(contMessages, llm.Message{
			Role:       llm.RoleTool,
			Content:    content,
			ToolCallID: id,
		})
	}

	resp, err := tp.provider.Complete(ctx, contMessages, availableTools)
	if err != nil {
		return "", fmt.Errorf("continuation LLM call: %w", err)
	}
	tp.logger.Debug("continuation call completed", "narrative_len", len(resp.Content), "tool_calls", len(resp.ToolCalls))
	return resp.Content, nil
}

// emitStatus calls the status callback if set.
func (tp *TurnProcessor) emitStatus(s api.StatusPayload) {
	if tp.StatusCallback != nil {
		tp.StatusCallback(s)
	}
}

// buildAppliedToolCall converts a raw tool call and its result into the
// engine's AppliedToolCall type, JSON-encoding the arguments and result data.
func buildAppliedToolCall(tc llm.ToolCall, result *tools.ToolResult) (AppliedToolCall, error) {
	argsJSON, err := json.Marshal(tc.Arguments)
	if err != nil {
		return AppliedToolCall{}, fmt.Errorf("marshal tool call arguments: %w", err)
	}

	var resultData map[string]any
	if result != nil {
		resultData = result.Data
	}
	resultJSON, err := json.Marshal(resultData)
	if err != nil {
		return AppliedToolCall{}, fmt.Errorf("marshal tool call result: %w", err)
	}

	return AppliedToolCall{
		Tool:      tc.Name,
		Arguments: json.RawMessage(argsJSON),
		Result:    json.RawMessage(resultJSON),
	}, nil
}
