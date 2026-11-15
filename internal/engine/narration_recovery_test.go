package engine

import (
	"errors"
	"testing"

	"git.subcult.tv/subculture-collective/edda/internal/llm"
)

func TestRecoveryNarrationRejectsNewMechanics(t *testing.T) {
	for _, tt := range []struct {
		name string
		resp *llm.Response
		want error
	}{
		{"nil", nil, ErrEmptyTurnResponse},
		{"empty", &llm.Response{}, ErrEmptyTurnResponse},
		{"only choices", &llm.Response{Content: "Choices:\n1. Wait.\n2. Look.\n3. Listen."}, ErrEmptyTurnResponse},
		{"tool", &llm.Response{Content: "A pause.", ToolCalls: []llm.ToolCall{{Name: "update_player_hp"}}}, ErrUnresolvedDurableClaims},
		{"unbacked claim", &llm.Response{Content: "The silver key is now in your inventory."}, ErrUnresolvedDurableClaims},
		{"valid", &llm.Response{Content: "A quiet pause.\n\nChoices:\n1. Wait.\n2. Look.\n3. Listen."}, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateRecoveryNarration(tt.resp, nil); !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}
