import { apiFetch } from './client';
import { buildCampaignPath } from './routes';
import type { LocationResponse } from './types';

function buildLocationsPath(campaignId: string): string {
  return buildCampaignPath(campaignId, 'locations');
}

function buildLocationPath(campaignId: string, locationId: string): string {
  return buildCampaignPath(campaignId, 'locations', locationId);
}

export function listCampaignLocations(campaignId: string): Promise<LocationResponse[]> {
  return apiFetch<LocationResponse[]>(buildLocationsPath(campaignId));
}

export function getCampaignLocation(campaignId: string, locationId: string): Promise<LocationResponse> {
  return apiFetch<LocationResponse>(buildLocationPath(campaignId, locationId));
}
