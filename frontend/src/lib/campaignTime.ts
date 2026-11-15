export function formatCampaignSeconds(seconds: number): string {
  const day = Math.floor(seconds / 86400) + 1;
  const hour = Math.floor(seconds / 3600) % 24;
  const minute = Math.floor(seconds / 60) % 60;
  const second = seconds % 60;
  return `Day ${day}, ${[hour, minute, second].map(value => String(value).padStart(2, '0')).join(':')}`;
}
