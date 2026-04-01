package campaign

import (
	"errors"
	"strings"

	"github.com/charmbracelet/huh"
)

// GenreOptions lists the genres available when creating a new campaign.
var GenreOptions = []string{"Fantasy", "Sci-Fi", "Horror", "Historical", "Modern", "Post-Apocalyptic", "Steampunk"}

// DifficultyOptions lists the difficulty levels available when creating a new campaign.
var DifficultyOptions = []string{"Casual", "Moderate", "Deadly"}

// CampaignFormResult holds the values collected by the new-campaign form.
type CampaignFormResult struct {
	Name       string
	Genre      string
	Difficulty string
}

// NewCampaignFormMsg is sent after the player has completed the new-campaign form.
type NewCampaignFormMsg struct {
	Result CampaignFormResult
}

var errEmptyName = errors.New("campaign name cannot be empty")

// buildCampaignForm constructs the multi-step Huh form for creating a new
// campaign. It collects name, genre, and difficulty across three groups.
func buildCampaignForm(result *CampaignFormResult) *huh.Form {
	genreOpts := make([]huh.Option[string], len(GenreOptions))
	for i, g := range GenreOptions {
		genreOpts[i] = huh.NewOption(g, g)
	}

	diffOpts := make([]huh.Option[string], len(DifficultyOptions))
	for i, d := range DifficultyOptions {
		diffOpts[i] = huh.NewOption(d, d)
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Campaign name").
				Placeholder("e.g. Shadows of the East").
				Value(&result.Name).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errEmptyName
					}
					return nil
				}),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Genre").
				Options(genreOpts...).
				Value(&result.Genre),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Difficulty").
				Options(diffOpts...).
				Value(&result.Difficulty),
		),
	).WithShowHelp(false)
}
