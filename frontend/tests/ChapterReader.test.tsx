import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, expect, it, vi } from 'vitest';
import { ChapterReader } from '../src/components/journal/ChapterReader';
import { getSessionHistory } from '../src/api/campaigns';
import type { ChapterResponse } from '../src/api/types';

vi.mock('../src/api/campaigns', () => ({ getSessionHistory: vi.fn() }));
const chapter = (id: string, turn: number): ChapterResponse => ({ id, campaign_id: 'alpha', from_turn: turn, to_turn: turn, source_hash: `hash-${id}`, source_turns: [{ log_id: `source-${turn}`, attempt_id: `action-${turn}`, turn_number: turn, player_input: `Input ${turn}`, narrative: `Recorded source ${turn}` }], knowledge_boundary: 'player-visible', rules_version: 'srd-5.2.1', rules_mode: 'light', created_at: '2026-01-02T00:00:00Z', revisions: [{ id: `revision-${id}`, chapter_id: id, request_id: `request-${id}`, revision: 1, status: 'completed', content: `Chapter prose ${turn}`, error: '', prompt_version: 'faithful-chapter-v1', created_at: '2026-01-02T00:00:00Z' }] });
function mount(chapters = [chapter('second', 2), chapter('first', 1)], campaignId = 'alpha') {
  return render(<QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}><ChapterReader key={campaignId} campaignId={campaignId} chapters={chapters} /></QueryClientProvider>);
}
beforeEach(() => { vi.mocked(getSessionHistory).mockResolvedValue({ entries: [{ turn_number: 3, player_input: 'New play', input_type: 'action', llm_response: 'A new event', created_at: '2026-01-03T00:00:00Z' }] }); });
it('sorts by played source order, navigates, shows freshness and inspects captured sources', async () => {
  mount();
  expect(await screen.findByText('Chapter prose 1')).toBeVisible();
  expect(await screen.findByText('1 completed turn awaits a saved chapter.')).toBeVisible();
  expect(screen.getByRole('button', { name: 'Previous chapter' })).toBeDisabled();
  await userEvent.click(screen.getByRole('button', { name: 'Next chapter' }));
  expect(screen.getByText('Chapter prose 2')).toBeVisible();
  // jsdom has no scrollIntoView implementation.
  HTMLElement.prototype.scrollIntoView = vi.fn();
  await userEvent.click(screen.getByRole('link', { name: 'Read source history' }));
  expect(screen.getByText('Recorded source 2')).toBeVisible();
  expect(screen.getByText(/Source source-2 · Action action-2/)).toBeVisible();
});
it('retains chapter, revision and relative reading offset across unmounts, with campaign isolation', async () => {
  const view = mount();
  await userEvent.click(screen.getByRole('button', { name: 'Next chapter' }));
  const pane = screen.getByRole('region', { name: 'Chapter prose' });
  Object.defineProperties(pane, { scrollHeight: { value: 1000, configurable: true }, clientHeight: { value: 200, configurable: true } });
  pane.scrollTop = 400; fireEvent.scroll(pane);
  expect(localStorage.getItem('edda:chapter-offset:alpha:second:revision-second')).toBe('0.5');
  view.unmount(); mount();
  expect(screen.getByLabelText('Chapter', { exact: true })).toHaveValue('second');
  await waitFor(() => expect(localStorage.getItem('edda:chapter-reader:alpha')).toContain('second'));
});
it('keeps saved prose readable during freshness failure and excludes failed revisions', async () => {
  vi.mocked(getSessionHistory).mockRejectedValue(new Error('offline'));
  const c = chapter('first', 1); c.revisions.push({ ...c.revisions[0], id: 'failed', revision: 2, status: 'failed', content: 'Unvalidated prose' });
  mount([c]);
  expect(await screen.findByText(/Freshness unavailable/)).toBeVisible();
  expect(screen.getByText('Chapter prose 1')).toBeVisible();
  expect(screen.queryByText('Unvalidated prose')).not.toBeInTheDocument();
});
it('does not count chapters from before restarted play as coverage of new turns', async () => {
  vi.mocked(getSessionHistory).mockResolvedValue({ entries: [{ turn_number: 1, player_input: 'Restart', input_type: 'action', llm_response: 'Different play', created_at: '2026-02-01T00:00:00Z' }] });
  mount([chapter('first', 1)]);
  expect(await screen.findByText('1 completed turn awaits a saved chapter.')).toBeVisible();
});

it('keeps pre-reset sources before newly captured reused turn ranges', async () => {
  const old = chapter('old', 1); old.to_turn = 10;
  const replacement = chapter('new', 1); replacement.created_at = '2026-02-01T00:00:00Z';
  replacement.revisions[0].content = 'Replacement play';
  old.revisions[0].content = 'Original play';
  mount([replacement, old]);
  expect(screen.getByText('Original play')).toBeVisible();
  await userEvent.click(screen.getByRole('button', { name: 'Next chapter' }));
  expect(screen.getByText('Replacement play')).toBeVisible();
});
