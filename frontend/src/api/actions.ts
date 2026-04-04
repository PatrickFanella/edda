import { apiFetch } from './client';
import { buildCampaignPath } from './routes';
import type { ActionRequest, TurnResponse } from './types';

function buildActionPath(campaignId: string): string {
  return buildCampaignPath(campaignId, 'action');
}

export function submitCampaignAction(campaignId: string, request: ActionRequest): Promise<TurnResponse> {
  return apiFetch<TurnResponse>(buildActionPath(campaignId), {
    method: 'POST',
    body: request,
  });
}
