import { apiFetch } from './client';
import { buildCampaignPath } from './routes';

export interface ReminderItem { text: string; source_id: string }
export interface ReturnReminder {
  scene?: ReminderItem;
  objective?: ReminderItem;
  known_facts: ReminderItem[];
}
export function getReturnReminder(campaignId: string): Promise<ReturnReminder> {
  return apiFetch(buildCampaignPath(campaignId, 'journal', 'reminder'));
}
