import { apiFetch } from './client';
import { buildCampaignPath } from './routes';
import type { NPCResponse } from './types';

function buildNPCsPath(campaignId: string): string {
  return buildCampaignPath(campaignId, 'npcs');
}

function buildNPCPath(campaignId: string, npcId: string): string {
  return buildCampaignPath(campaignId, 'npcs', npcId);
}

export function listCampaignNPCs(campaignId: string): Promise<NPCResponse[]> {
  return apiFetch<NPCResponse[]>(buildNPCsPath(campaignId));
}

export function getCampaignNPC(campaignId: string, npcId: string): Promise<NPCResponse> {
  return apiFetch<NPCResponse>(buildNPCPath(campaignId, npcId));
}
