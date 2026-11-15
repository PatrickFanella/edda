package tools

import "git.subcult.tv/subculture-collective/edda/internal/rules/srd521"

// SRDCheckResult adapts an already resolved check to the shared engine journal
// shape. It never receives raw difficulty, facts or model-provided modifiers.
// No SRD tool is registered in the legacy campaign registry by this adapter.
func SRDCheckResult(resolution srd521.CheckResolution) *ToolResult {
	return &ToolResult{
		Success: true,
		Data:    map[string]any{"receipt": resolution.PlayerReceipt()},
	}
}
