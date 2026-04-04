import { apiFetch } from './client';
import { buildCampaignPath } from './routes';
import type { CharacterAbility, CharacterResponse, ItemResponse } from './types';

function buildCharacterPath(campaignId: string, suffix?: 'inventory' | 'abilities'): string {
  return suffix ? buildCampaignPath(campaignId, 'character', suffix) : buildCampaignPath(campaignId, 'character');
}

export function getCampaignCharacter(campaignId: string): Promise<CharacterResponse> {
  return apiFetch<CharacterResponse>(buildCharacterPath(campaignId));
}

export function getCampaignCharacterInventory(campaignId: string): Promise<ItemResponse[]> {
  return apiFetch<ItemResponse[]>(buildCharacterPath(campaignId, 'inventory'));
}

export function getCampaignCharacterAbilities(campaignId: string): Promise<CharacterAbility[]> {
  return apiFetch<CharacterAbility[]>(buildCharacterPath(campaignId, 'abilities'));
}
