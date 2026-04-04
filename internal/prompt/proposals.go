package prompt

import (
	_ "embed"
	"fmt"
)

// proposalsTemplate is the parameterized system prompt for campaign proposal
// generation. It contains three %s placeholders for player preference fields.
//
//go:embed proposals.txt
var proposalsTemplate string

// BuildProposalsPrompt returns the system prompt for campaign proposal
// generation, interpolating the given player preferences into the template.
func BuildProposalsPrompt(genre, settingStyle, tone string) string {
	return fmt.Sprintf(proposalsTemplate, genre, settingStyle, tone)
}
