package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type QuestType string

const (
	QuestTypeShortTerm  QuestType = "short_term"
	QuestTypeMediumTerm QuestType = "medium_term"
	QuestTypeLongTerm   QuestType = "long_term"
)

type QuestStatus string

const (
	QuestStatusActive    QuestStatus = "active"
	QuestStatusCompleted QuestStatus = "completed"
	QuestStatusFailed    QuestStatus = "failed"
	QuestStatusAbandoned QuestStatus = "abandoned"
)

type Quest struct {
	ID            uuid.UUID
	CampaignID    uuid.UUID
	ParentQuestID *uuid.UUID
	Title         string
	Description   string
	QuestType     QuestType
	Status        QuestStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (q *Quest) Validate() error {
	if q.Title == "" {
		return errors.New("quest title is required")
	}
	if q.CampaignID == uuid.Nil {
		return errors.New("quest campaign_id is required")
	}
	return nil
}

type QuestObjective struct {
	ID          uuid.UUID
	QuestID     uuid.UUID
	Description string
	Completed   bool
	OrderIndex  int
}
