package prompt

import (
	_ "embed"
	"fmt"
)

// sceneTemplate is the parameterized system prompt for opening scene
// generation. It contains six %s placeholders for campaign and world details.
//
//go:embed scene.txt
var sceneTemplate string

// BuildScenePrompt returns the system prompt for opening scene generation,
// interpolating the given campaign and world details into the template.
func BuildScenePrompt(genre, tone, themes, startingLocation, npcList, factList string) string {
	return fmt.Sprintf(sceneTemplate, genre, tone, themes, startingLocation, npcList, factList)
}
