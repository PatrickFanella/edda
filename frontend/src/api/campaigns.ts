import { apiFetch, apiFetchVoid } from './client';
import { buildCampaignPath, buildCampaignsPath } from './routes';
import type { CampaignCreateRequest, CampaignListResponse, CampaignResponse, CampaignTimeResponse, SavePointResponse, SessionHistoryResponse } from './types';

export function listCampaigns(): Promise<CampaignListResponse> {
  return apiFetch<CampaignListResponse>(buildCampaignsPath());
}

export function createCampaign(request: CampaignCreateRequest): Promise<CampaignResponse> {
  return apiFetch<CampaignResponse>(buildCampaignsPath(), {
    method: 'POST',
    body: request,
  });
}

export function getCampaign(campaignId: string): Promise<CampaignResponse> {
  return apiFetch<CampaignResponse>(buildCampaignPath(campaignId));
}

export function updateCampaign(campaignId: string, request: CampaignCreateRequest): Promise<CampaignResponse> {
  return apiFetch<CampaignResponse>(buildCampaignPath(campaignId), {
    method: 'PUT',
    body: request,
  });
}

export function deleteCampaign(campaignId: string): Promise<void> {
  return apiFetchVoid(buildCampaignPath(campaignId), {
    method: 'DELETE',
  });
}

export function getSessionHistory(campaignId: string): Promise<SessionHistoryResponse> {
  return apiFetch<SessionHistoryResponse>(buildCampaignPath(campaignId, 'history'));
}

export function startOverCampaign(campaignId: string): Promise<void> {
  return apiFetchVoid(buildCampaignPath(campaignId, 'start-over'), { method: 'POST' });
}

export function listSavePoints(campaignId: string): Promise<SavePointResponse[]> {
  return apiFetch<SavePointResponse[]>(buildCampaignPath(campaignId, 'saves'));
}

export function createManualSave(campaignId: string, name: string): Promise<SavePointResponse> {
  return apiFetch<SavePointResponse>(buildCampaignPath(campaignId, 'saves'), {
    method: 'POST',
    body: { name },
  });
}

export function getCampaignTime(campaignId: string): Promise<CampaignTimeResponse> {
  return apiFetch<CampaignTimeResponse>(buildCampaignPath(campaignId, 'time'));
}
