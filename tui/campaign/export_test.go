package campaign

import "github.com/charmbracelet/huh"

// ExportBuildCampaignForm exposes buildCampaignForm for external tests.
func ExportBuildCampaignForm(result *CampaignFormResult) *huh.Form {
	return buildCampaignForm(result)
}
