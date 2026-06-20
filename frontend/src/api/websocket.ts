import { API_BASE } from './backend';
export { decodeWebSocketEvent, type ParsedWebSocketEvent } from './websocketParser';

export function buildCampaignWebSocketURL(campaignId: string): string {
  const origin = window.location.origin;
  const base = API_BASE.startsWith('http://') || API_BASE.startsWith('https://') ? API_BASE : `${origin}${API_BASE}`;
  const url = new URL(base);

  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
  url.pathname = `${url.pathname.replace(/\/$/, '')}/campaigns/${encodeURIComponent(campaignId)}/ws`;
  url.search = '';
  url.hash = '';

  return url.toString();
}
