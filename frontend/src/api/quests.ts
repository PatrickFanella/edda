import { apiFetch } from './client';
import { buildCampaignPath } from './routes';
import type { QuestResponse } from './types';

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
