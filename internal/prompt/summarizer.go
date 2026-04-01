package prompt

import _ "embed"

// Summarizer is the system prompt that instructs the LLM to produce structured
// JSON summaries of individual game turns for long-term memory storage.
//
//go:embed summarizer.txt
var Summarizer string
