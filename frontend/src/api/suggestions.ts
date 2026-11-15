import { apiFetch } from './client';
import { buildCampaignPath } from './routes';
import type { ActionIntent } from './types';

export interface ActionSuggestion {
  id: string;
  text: string;
  reason: string;
  intent?: ActionIntent;
  source_ids: string[];
}
export interface ActionSuggestions {
  version: string;
  candidates: ActionSuggestion[];
}
export function getActionSuggestions(campaignId: string): Promise<ActionSuggestions> {
  return apiFetch(buildCampaignPath(campaignId, 'action-suggestions'));
}

// A local word-prefix match: typing never calls a provider or another endpoint.
export function matchSuggestions(candidates: readonly ActionSuggestion[], draft: string): ActionSuggestion[] {
  const words = draft.toLocaleLowerCase().match(/[\p{L}\p{N}]+/gu)?.filter(word => !['i', 'my', 'the', 'a', 'to', 'using'].includes(word)) ?? [];
  return candidates.filter(candidate => {
    const tokens = `${candidate.text} ${candidate.reason}`.toLocaleLowerCase().match(/[\p{L}\p{N}]+/gu) ?? [];
    return words.every(word => tokens.some(token => token.startsWith(word)));
  }).slice(0, 2);
}
