package prompt

import _ "embed"

// CharacterInterview is the system prompt that instructs the LLM how to
// conduct the campaign-creation interview, gathering player preferences across
// genre, tone, themes, world type, danger level, and political complexity.
//
//go:embed character_interview.txt
var CharacterInterview string
