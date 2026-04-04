package prompt

import (
	_ "embed"
	"fmt"
)

// campaignNameTemplate is the parameterized system prompt for campaign name
// generation. It contains four %s placeholders for genre, tone, themes, and
// world type.
//
//go:embed campaign_name.txt
var campaignNameTemplate string

// BuildCampaignNamePrompt returns the system prompt for campaign name
// generation, interpolating the given campaign profile fields into the
// template.
func BuildCampaignNamePrompt(genre, tone, themes, worldType string) string {
	return fmt.Sprintf(campaignNameTemplate, genre, tone, themes, worldType)
}
