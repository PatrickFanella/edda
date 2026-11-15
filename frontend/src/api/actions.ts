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

export interface PreparedAction {
  action_id: string;
  expected_revision: string;
  intent: import('./types').ActionIntent;
  target_id?: string;
}
export function confirmCharacterAction(campaignId: string, request: PreparedAction): Promise<TurnResponse> {
  return apiFetch(buildCampaignPath(campaignId, 'action', 'confirm'), { method: 'POST', body: request });
}
export function readActionReceipt(campaignId: string, actionId: string): Promise<{ result?: TurnResponse } | null> {
  return apiFetch(`${buildActionPath(campaignId)}?action_id=${encodeURIComponent(actionId)}`);
}

export function publishConfirmedAction(campaignId: string, result: TurnResponse, input: string) {
  window.dispatchEvent(new CustomEvent('edda:confirmed-action', { detail: { campaignId, result, input } }));
}
