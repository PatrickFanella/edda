package domain

import (
	"context"

	"github.com/google/uuid"
)

type UpdateQuestMutationCommand struct {
	CampaignID  uuid.UUID
	QuestID     uuid.UUID
	Status      *QuestStatus
	Description *string
	Objectives  []string
}

type CompleteObjectiveMutationCommand struct {
	CampaignID           uuid.UUID
	QuestID              uuid.UUID
	ObjectiveID          *uuid.UUID
	ObjectiveDescription *string
	AutoCompleteQuest    bool
}

type UpdateQuestMutationResult struct {
	Quest           QuestSummary
	AddedObjectives []QuestObjective
	CascadedQuests  []QuestStatusCascade
	Objectives      []QuestObjective
	Narrative       string
}

type CompleteObjectiveMutationResult struct {
	Quest                   QuestSummary
	Objective               QuestObjective
	ObjectiveWasComplete    bool
	ObjectivesCompleted     int
	ObjectivesTotal         int
	Progress                string
	AllObjectivesComplete   bool
	QuestReadyForCompletion bool
	QuestAutoCompleted      bool
	Narrative               string
}

type QuestSummary struct {
	ID            uuid.UUID
	CampaignID    uuid.UUID
	ParentQuestID *uuid.UUID
	Title         string
	Description   string
	QuestType     QuestType
	Status        QuestStatus
}

type QuestStatusCascade struct {
	ID        uuid.UUID
	Title     string
	OldStatus QuestStatus
	NewStatus QuestStatus
}

type QuestMutationStore interface {
	UpdateQuest(ctx context.Context, cmd UpdateQuestMutationCommand) (*UpdateQuestMutationResult, error)
	CompleteObjective(ctx context.Context, cmd CompleteObjectiveMutationCommand) (*CompleteObjectiveMutationResult, error)
}
