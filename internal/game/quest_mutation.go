package game

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
)

type questMutationStore interface {
	GetQuestByID(ctx context.Context, id pgtype.UUID) (statedb.Quest, error)
	UpdateQuest(ctx context.Context, arg statedb.UpdateQuestParams) (statedb.Quest, error)
	UpdateQuestStatus(ctx context.Context, arg statedb.UpdateQuestStatusParams) (statedb.Quest, error)
	ListObjectivesByQuest(ctx context.Context, questID pgtype.UUID) ([]statedb.QuestObjective, error)
	CreateObjective(ctx context.Context, arg statedb.CreateObjectiveParams) (statedb.QuestObjective, error)
	CompleteObjective(ctx context.Context, id pgtype.UUID) (statedb.QuestObjective, error)
	ListQuestsByCampaign(ctx context.Context, campaignID pgtype.UUID) ([]statedb.Quest, error)
}

func (s *worldService) UpdateQuest(ctx context.Context, cmd domain.UpdateQuestMutationCommand) (*domain.UpdateQuestMutationResult, error) {
	quest, err := s.getCampaignQuest(ctx, cmd.CampaignID, cmd.QuestID)
	if err != nil {
		return nil, err
	}
	updated := quest
	if cmd.Description != nil {
		updated, err = s.queries.UpdateQuest(ctx, statedb.UpdateQuestParams{ID: quest.ID, ParentQuestID: quest.ParentQuestID, Title: quest.Title, Description: pgtype.Text{String: *cmd.Description, Valid: true}, QuestType: quest.QuestType, Status: quest.Status})
		if err != nil {
			return nil, err
		}
	}
	statusChanged := cmd.Status != nil && string(*cmd.Status) != quest.Status
	if statusChanged {
		updated, err = s.queries.UpdateQuestStatus(ctx, statedb.UpdateQuestStatusParams{ID: quest.ID, Status: string(*cmd.Status)})
		if err != nil {
			return nil, err
		}
	}
	cascaded, err := s.cascadeActiveSubquests(ctx, updated, cascadeStatusForUpdate(statusChanged, cmd.Status))
	if err != nil {
		return nil, err
	}
	added := make([]domain.QuestObjective, 0, len(cmd.Objectives))
	if len(cmd.Objectives) > 0 {
		objectives, err := s.queries.ListObjectivesByQuest(ctx, quest.ID)
		if err != nil {
			return nil, err
		}
		maxIdx := int32(-1)
		for _, o := range objectives {
			if o.OrderIndex > maxIdx {
				maxIdx = o.OrderIndex
			}
		}
		for i, desc := range cmd.Objectives {
			created, err := s.queries.CreateObjective(ctx, statedb.CreateObjectiveParams{QuestID: quest.ID, Description: desc, Completed: false, OrderIndex: maxIdx + 1 + int32(i)})
			if err != nil {
				return nil, err
			}
			added = append(added, toDomainObjective(created))
		}
	}
	objectives, err := s.queries.ListObjectivesByQuest(ctx, quest.ID)
	if err != nil {
		return nil, err
	}
	narrativeParts := []string{fmt.Sprintf("Quest %q updated", updated.Title)}
	if cmd.Status != nil {
		narrativeParts = append(narrativeParts, fmt.Sprintf("status is now %q", updated.Status))
	}
	if cmd.Description != nil {
		narrativeParts = append(narrativeParts, "description updated")
	}
	if len(added) > 0 {
		narrativeParts = append(narrativeParts, fmt.Sprintf("%d new objective(s) added", len(added)))
	}
	if len(cascaded) > 0 {
		narrativeParts = append(narrativeParts, fmt.Sprintf("%d subquest(s) cascaded", len(cascaded)))
	}
	return &domain.UpdateQuestMutationResult{Quest: toDomainQuest(updated), AddedObjectives: added, CascadedQuests: cascaded, Objectives: toDomainObjectives(objectives), Narrative: strings.Join(narrativeParts, ". ") + "."}, nil
}

func (s *worldService) CompleteObjective(ctx context.Context, cmd domain.CompleteObjectiveMutationCommand) (*domain.CompleteObjectiveMutationResult, error) {
	quest, err := s.getCampaignQuest(ctx, cmd.CampaignID, cmd.QuestID)
	if err != nil {
		return nil, err
	}
	objectives, err := s.queries.ListObjectivesByQuest(ctx, quest.ID)
	if err != nil {
		return nil, err
	}
	target, err := selectObjectiveForMutation(objectives, cmd)
	if err != nil {
		return nil, err
	}
	wasCompleted := target.Completed
	if !target.Completed {
		target, err = s.queries.CompleteObjective(ctx, target.ID)
		if err != nil {
			return nil, err
		}
	}
	completed := 0
	for _, o := range objectives {
		if o.Completed || o.ID == target.ID {
			completed++
		}
	}
	all := completed == len(objectives)
	questStatus := toDomainQuest(quest).Status
	updatedQuest := quest
	autoCompleted := false
	if all && cmd.AutoCompleteQuest && questStatus != domain.QuestStatusCompleted {
		updatedQuest, err = s.queries.UpdateQuestStatus(ctx, statedb.UpdateQuestStatusParams{ID: quest.ID, Status: string(domain.QuestStatusCompleted)})
		if err != nil {
			return nil, err
		}
		autoCompleted = true
		questStatus = domain.QuestStatusCompleted
		_, err := s.cascadeActiveSubquests(ctx, updatedQuest, &questStatus)
		if err != nil {
			return nil, err
		}
	}
	questReady := all && questStatus != domain.QuestStatusCompleted
	narrativePrefix := fmt.Sprintf("Objective %q completed", target.Description)
	if wasCompleted {
		narrativePrefix = fmt.Sprintf("Objective %q was already complete", target.Description)
	}
	narrative := fmt.Sprintf("%s. Progress: %d/%d complete.", narrativePrefix, completed, len(objectives))
	if questReady {
		narrative += fmt.Sprintf(" All objectives are complete; quest %q is ready for completion.", quest.Title)
	}
	if autoCompleted {
		narrative += fmt.Sprintf(" Quest %q has been auto-completed.", quest.Title)
	}
	return &domain.CompleteObjectiveMutationResult{Quest: toDomainQuest(updatedQuest), Objective: toDomainObjective(target), ObjectiveWasComplete: wasCompleted, ObjectivesCompleted: completed, ObjectivesTotal: len(objectives), Progress: fmt.Sprintf("%d/%d", completed, len(objectives)), AllObjectivesComplete: all, QuestReadyForCompletion: questReady, QuestAutoCompleted: autoCompleted, Narrative: strings.TrimSpace(narrative)}, nil
}

func cascadeStatusForUpdate(statusChanged bool, status *domain.QuestStatus) *domain.QuestStatus {
	if !statusChanged || status == nil {
		return nil
	}
	if *status != domain.QuestStatusCompleted && *status != domain.QuestStatusFailed {
		return nil
	}
	return status
}

func (s *worldService) getCampaignQuest(ctx context.Context, campaignID, questID uuid.UUID) (statedb.Quest, error) {
	quest, err := s.queries.GetQuestByID(ctx, statedb.GetQuestByIDParams{ID: dbutil.ToPgtype(questID)})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return statedb.Quest{}, errors.New("quest_id does not reference an existing quest")
		}
		return statedb.Quest{}, err
	}
	if dbutil.FromPgtype(quest.CampaignID) != campaignID {
		return statedb.Quest{}, errors.New("quest_id does not reference an existing quest")
	}
	return quest, nil
}

func (s *worldService) cascadeActiveSubquests(ctx context.Context, parent statedb.Quest, status *domain.QuestStatus) ([]domain.QuestStatusCascade, error) {
	if status == nil || *status == domain.QuestStatusActive {
		return nil, nil
	}
	quests, err := s.queries.ListQuestsByCampaign(ctx, parent.CampaignID)
	if err != nil {
		return nil, err
	}
	children := map[[16]byte][]statedb.Quest{}
	for _, q := range quests {
		if q.ParentQuestID.Valid {
			children[q.ParentQuestID.Bytes] = append(children[q.ParentQuestID.Bytes], q)
		}
	}
	cascadeStatus := domain.QuestStatusFailed
	if *status == domain.QuestStatusCompleted {
		cascadeStatus = domain.QuestStatusAbandoned
	}
	queue := append([]statedb.Quest(nil), children[parent.ID.Bytes]...)
	out := []domain.QuestStatusCascade{}
	for len(queue) > 0 {
		q := queue[0]
		queue = queue[1:]
		if q.Status == string(domain.QuestStatusActive) {
			updated, err := s.queries.UpdateQuestStatus(ctx, statedb.UpdateQuestStatusParams{ID: q.ID, Status: string(cascadeStatus)})
			if err != nil {
				return nil, err
			}
			out = append(out, domain.QuestStatusCascade{ID: dbutil.FromPgtype(updated.ID), Title: updated.Title, OldStatus: domain.QuestStatusActive, NewStatus: cascadeStatus})
		}
		queue = append(queue, children[q.ID.Bytes]...)
	}
	return out, nil
}

func selectObjectiveForMutation(objectives []statedb.QuestObjective, cmd domain.CompleteObjectiveMutationCommand) (statedb.QuestObjective, error) {
	if cmd.ObjectiveID != nil {
		for _, o := range objectives {
			if dbutil.FromPgtype(o.ID) == *cmd.ObjectiveID {
				return o, nil
			}
		}
		return statedb.QuestObjective{}, errors.New("objective_id does not belong to the specified quest")
	}
	needle := normalizeObjectiveText(*cmd.ObjectiveDescription)
	var exact, contains []statedb.QuestObjective
	for _, o := range objectives {
		c := normalizeObjectiveText(o.Description)
		if c == needle {
			exact = append(exact, o)
		} else if strings.Contains(c, needle) {
			contains = append(contains, o)
		}
	}
	if len(exact) == 1 {
		return exact[0], nil
	}
	if len(exact) > 1 || len(contains) > 1 {
		return statedb.QuestObjective{}, errors.New("objective_description matches multiple objectives; provide objective_id")
	}
	if len(contains) == 1 {
		return contains[0], nil
	}
	return statedb.QuestObjective{}, errors.New("objective_description did not match any objective in the quest")
}

func toDomainQuest(q statedb.Quest) domain.QuestSummary {
	return domain.QuestSummary{ID: dbutil.FromPgtype(q.ID), CampaignID: dbutil.FromPgtype(q.CampaignID), ParentQuestID: optionalUUIDFromPgtype(q.ParentQuestID), Title: q.Title, Description: q.Description.String, QuestType: domain.QuestType(q.QuestType), Status: domain.QuestStatus(q.Status)}
}
func toDomainObjective(o statedb.QuestObjective) domain.QuestObjective {
	return domain.QuestObjective{ID: dbutil.FromPgtype(o.ID), QuestID: dbutil.FromPgtype(o.QuestID), Description: o.Description, Completed: o.Completed, OrderIndex: int(o.OrderIndex)}
}
func toDomainObjectives(os []statedb.QuestObjective) []domain.QuestObjective {
	out := make([]domain.QuestObjective, 0, len(os))
	for _, o := range os {
		out = append(out, toDomainObjective(o))
	}
	return out
}
func optionalUUIDFromPgtype(id pgtype.UUID) *uuid.UUID {
	if !id.Valid {
		return nil
	}
	v := dbutil.FromPgtype(id)
	return &v
}
func normalizeObjectiveText(text string) string {
	text = strings.ToLower(strings.TrimSpace(text))
	text = strings.ReplaceAll(text, "_", " ")
	return strings.Join(strings.Fields(text), " ")
}
