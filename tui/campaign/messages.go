package campaign

import (
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
	"github.com/PatrickFanella/game-master/internal/world"
)

// --- Messages emitted by campaign wizard sub-models ---

// SelectedMsg is sent when the player selects an existing campaign from the list.
type SelectedMsg struct {
	Campaign statedb.Campaign
}

// NewCampaignMsg is sent when the player chooses "New Campaign" from the picker.
type NewCampaignMsg struct{}

// MethodChosenMsg is sent when the player picks a campaign creation method.
type MethodChosenMsg struct {
	Method CreationMethod
}

// CreationMethod identifies which path the player chose.
type CreationMethod int

const (
	MethodDescribe   CreationMethod = iota // free-form interview
	MethodAttributes                       // guided attribute selection
)

// ProfileReadyMsg is sent when a CampaignProfile has been gathered
// (from either the interview or the proposals path).
type ProfileReadyMsg struct {
	Profile world.CampaignProfile
	Name    string // populated for attribute path (proposal name)
	Summary string // populated for attribute path (proposal summary)
}

// AttributesReadyMsg is sent when the attribute form is completed.
type AttributesReadyMsg struct {
	Genre        string
	SettingStyle string
	Tone         string
}

// ProposalSelectedMsg is sent when the player picks one of the 3 proposals.
type ProposalSelectedMsg struct {
	Proposal world.CampaignProposal
}

// CharMethodChosenMsg is sent when the player picks a character creation method.
type CharMethodChosenMsg struct {
	Method CreationMethod
}

// CharacterReadyMsg is sent when a CharacterProfile has been gathered.
type CharacterReadyMsg struct {
	Profile world.CharacterProfile
}

// ConfirmedMsg is sent when the player confirms the campaign + character.
type ConfirmedMsg struct {
	Name             string
	Summary          string
	Profile          world.CampaignProfile
	CharacterProfile world.CharacterProfile
}

// ChangeMsg is sent when the player wants to go back from confirmation.
type ChangeMsg struct{}

// WorldReadyMsg is sent when the orchestrator finishes world building.
type WorldReadyMsg struct {
	Result *world.OrchestratorResult
}

// WorldErrorMsg is sent when world building fails.
type WorldErrorMsg struct {
	Err error
}

// BackMsg is sent when the player wants to go back to the previous step.
type BackMsg struct{}
