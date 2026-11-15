import type { ChapterResponse } from '../../src/api/types';

const id = (value: number) => `00000000-0000-4000-8000-${String(value).padStart(12, '0')}`;
export const chapterExportFixture = (number: number): ChapterResponse => ({
  id: id(number), campaign_id: 'alpha', from_turn: number, to_turn: number,
  source_hash: `source-hash-${number}`, knowledge_boundary: 'player-visible-transcript-through-to-turn-v1',
  rules_version: 'srd-5.2.1', rules_mode: 'light', created_at: `2026-01-0${number}T00:00:00Z`,
  source_turns: [{ log_id: id(number + 10), attempt_id: id(number + 20), turn_number: number, player_input: 'Try the gate', narrative: 'The gate remained closed.' }],
  revisions: [1, 2, 3, 4].map(revision => ({ id: id(number * 100 + revision), chapter_id: id(number), request_id: id(number * 1000 + revision), revision,
    status: revision < 3 ? 'completed' : revision === 3 ? 'failed' : 'generating', content: revision < 3 ? `Saved chapter ${number}, revision ${revision}.\n\nThe gate remained closed.` : 'UNFINISHED_PROSE',
    error: revision === 3 ? 'FAILED_DETAILS' : '', prompt_version: 'faithful-chapter-v1', created_at: '2026-01-03T00:00:00Z' })),
});
