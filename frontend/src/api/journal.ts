import { apiFetch, apiFetchVoid } from './client';
import { buildCampaignPath } from './routes';
import type { ChapterResponse, ChapterRevisionResponse, JournalEntryResponse, SessionSummaryResponse } from './types';

// ---------- Summaries ----------

export function listSummaries(campaignId: string): Promise<SessionSummaryResponse[]> {
  return apiFetch<SessionSummaryResponse[]>(buildCampaignPath(campaignId, 'journal', 'summaries'));
}

export interface SummarizeRequest {
  from_turn?: number;
  to_turn?: number;
}

export function triggerSummarize(
  campaignId: string,
  request?: SummarizeRequest,
): Promise<SessionSummaryResponse> {
  return apiFetch<SessionSummaryResponse>(buildCampaignPath(campaignId, 'journal', 'summarize'), {
    method: 'POST',
    body: request ?? {},
  });
}

// ---------- Journal Entries ----------

export function listEntries(campaignId: string): Promise<JournalEntryResponse[]> {
  return apiFetch<JournalEntryResponse[]>(buildCampaignPath(campaignId, 'journal', 'entries'));
}

export interface CreateEntryRequest {
  title: string;
  content: string;
}

export function createEntry(
  campaignId: string,
  request: CreateEntryRequest,
): Promise<JournalEntryResponse> {
  return apiFetch<JournalEntryResponse>(buildCampaignPath(campaignId, 'journal', 'entries'), {
    method: 'POST',
    body: request,
  });
}

export function deleteEntry(campaignId: string, entryId: string): Promise<void> {
  return apiFetchVoid(buildCampaignPath(campaignId, 'journal', 'entries', entryId), {
    method: 'DELETE',
  });
}

export interface GenerateChapterRequest {
  request_id: string;
  chapter_id?: string;
}
export function listChapters(campaignId: string): Promise<ChapterResponse[]> {
  return apiFetch(buildCampaignPath(campaignId, 'journal', 'chapters'));
}
export function generateChapter(campaignId: string, request: GenerateChapterRequest): Promise<ChapterRevisionResponse | { message: string }> {
  return apiFetch(buildCampaignPath(campaignId, 'journal', 'chapters'), { method: 'POST', body: request });
}
