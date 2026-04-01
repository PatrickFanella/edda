package llmutil

import "strings"

// StripMarkdownFences removes ```json ... ``` wrapping that LLMs
// sometimes add around JSON output.
func StripMarkdownFences(s string) string {
	trimmed := strings.TrimSpace(s)
	if !strings.HasPrefix(trimmed, "```") {
		return s
	}
	// Remove opening fence line.
	if idx := strings.Index(trimmed, "\n"); idx != -1 {
		trimmed = trimmed[idx+1:]
	}
	// Remove closing fence.
	if idx := strings.LastIndex(trimmed, "```"); idx != -1 {
		trimmed = trimmed[:idx]
	}
	return strings.TrimSpace(trimmed)
}
