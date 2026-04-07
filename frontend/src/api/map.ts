import { apiFetch } from './client';
import { buildCampaignPath } from './routes';
import type { MapDataResponse } from './types';

export function getMapData(campaignId: string): Promise<MapDataResponse> {
  return apiFetch<MapDataResponse>(buildCampaignPath(campaignId, 'map'));
}
