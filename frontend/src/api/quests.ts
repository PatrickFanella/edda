import { apiFetch } from './client';
import { buildCampaignPath } from './routes';
import type { QuestResponse, QuestNoteResponse, QuestHistoryEntry } from './types';

export interface QuestListQuery {
  type?: string;
  status?: string;
}

function buildQuestsPath(campaignId: string): string {
  return buildCampaignPath(campaignId, 'quests');
}

function buildQuestPath(campaignId: string, questId: string): string {
  return buildCampaignPath(campaignId, 'quests', questId);
}

function buildQuestListPath(campaignId: string, query?: QuestListQuery): string {
  const params = new URLSearchParams();

  if (query?.type) {
    params.set('type', query.type);
  }

  if (query?.status) {
    params.set('status', query.status);
  }

  const search = params.toString();
  return search ? `${buildQuestsPath(campaignId)}?${search}` : buildQuestsPath(campaignId);
}

export function listCampaignQuests(campaignId: string, query?: QuestListQuery): Promise<QuestResponse[]> {
  return apiFetch<QuestResponse[]>(buildQuestListPath(campaignId, query));
}

export function getCampaignQuest(campaignId: string, questId: string): Promise<QuestResponse> {
  return apiFetch<QuestResponse>(buildQuestPath(campaignId, questId));
}

export function listQuestNotes(campaignId: string, questId: string): Promise<QuestNoteResponse[]> {
  return apiFetch<QuestNoteResponse[]>(buildCampaignPath(campaignId, 'quests', questId, 'notes'));
}

export function createQuestNote(campaignId: string, questId: string, content: string): Promise<QuestNoteResponse> {
  return apiFetch<QuestNoteResponse>(buildCampaignPath(campaignId, 'quests', questId, 'notes'), {
    method: 'POST',
    body: { content },
  });
}

export function deleteQuestNote(campaignId: string, questId: string, noteId: string): Promise<void> {
  return apiFetch<void>(buildCampaignPath(campaignId, 'quests', questId, 'notes', noteId), {
    method: 'DELETE',
  });
}

export function listQuestHistory(campaignId: string, questId: string): Promise<QuestHistoryEntry[]> {
  return apiFetch<QuestHistoryEntry[]>(buildCampaignPath(campaignId, 'quests', questId, 'history'));
}
