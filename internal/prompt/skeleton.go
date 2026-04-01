package prompt

import (
	_ "embed"
	"fmt"
)

// skeletonTemplate is the parameterized system prompt for world skeleton
// generation. It contains six %s placeholders for campaign profile fields.
//
//go:embed skeleton.txt
var skeletonTemplate string

// BuildSkeletonPrompt returns the system prompt for world skeleton generation,
// interpolating the given campaign profile fields into the template.
func BuildSkeletonPrompt(genre, tone, themes, worldType, dangerLevel, politicalComplexity string) string {
	return fmt.Sprintf(skeletonTemplate, genre, tone, themes, worldType, dangerLevel, politicalComplexity)
}
