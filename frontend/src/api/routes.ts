export function encodePathSegment(segment: string, label: string): string {
  const cleaned = segment.trim().replace(/^\/+|\/+$/g, '');
  if (!cleaned) {
    throw new Error(`${label} is required`);
  }

  return encodeURIComponent(cleaned);
}

export function buildCampaignsPath(): string {
  return '/campaigns';
}

export function buildCampaignPath(campaignId: string, ...segments: string[]): string {
  const suffix = segments.map((segment, index) => encodePathSegment(segment, `segment_${index}`));
  return `${buildCampaignsPath()}/${encodePathSegment(campaignId, 'campaignId')}${suffix.length > 0 ? `/${suffix.join('/')}` : ''}`;
}
