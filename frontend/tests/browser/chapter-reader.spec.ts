import { test, expect, signIn } from './fixtures';
import { history } from '../fixtures/data';

for (const width of [390, 1440]) {
  test(`reads saved chapters with navigation and retained position at ${width}`, async ({ page, game }) => {
    await page.setViewportSize({ width, height: 900 });
    const makeChapter = (turn: number) => ({ id: `chapter-${turn}`, campaign_id: 'alpha', from_turn: turn, to_turn: turn, source_hash: `hash-${turn}`, source_turns: [{ log_id: `source-${turn}`, attempt_id: `action-${turn}`, turn_number: turn, player_input: 'Inspect the gate', narrative: `The source outcome ${turn}.` }], knowledge_boundary: 'player-visible', rules_version: 'srd-5.2.1', rules_mode: 'light', created_at: '2026-02-01T00:00:00Z', revisions: [{ id: `revision-${turn}`, chapter_id: `chapter-${turn}`, request_id: `request-${turn}`, revision: 1, status: 'completed', content: Array.from({ length: 80 }, (_, i) => `Chapter ${turn}, paragraph ${i + 1}. The traveler followed the recorded path beside the gate.`).join('\n\n'), error: '', prompt_version: 'faithful-chapter-v1', created_at: '2026-02-01T00:00:00Z' }] });
    await page.route('**/api/v1/campaigns/alpha/journal/chapters', route => route.fulfill({ json: [makeChapter(2), makeChapter(1)] }));
    await page.route('**/api/v1/campaigns/alpha/history', route => route.fulfill({ json: { entries: [...history(), { turn_number: 3, player_input: 'Continue', input_type: 'action', llm_response: 'New play.', created_at: '2026-03-01T00:00:00Z' }] } }));
    await signIn(page); await page.goto('/play/alpha');
    await page.getByRole('tab', { name: 'Journal', exact: true }).click();
    const pane = page.getByRole('region', { name: 'Chapter prose' });
    await expect(page.getByRole('combobox', { name: 'Chapter', exact: true })).toHaveValue('chapter-1');
    await expect(page.getByText('1 completed turn awaits a saved chapter.')).toBeVisible();
    await pane.evaluate(element => { element.scrollTop = (element.scrollHeight - element.clientHeight) * 0.45; element.dispatchEvent(new Event('scroll', { bubbles: true })); });
    await expect.poll(() => page.evaluate(() => localStorage.getItem('edda:chapter-offset:alpha:chapter-1:revision-1'))).not.toBeNull();
    const offset = await pane.evaluate(element => element.scrollTop);
    await page.getByRole('button', { name: 'Next chapter', exact: true }).click();
    await expect(page.getByRole('combobox', { name: 'Chapter', exact: true })).toHaveValue('chapter-2');
    await page.getByRole('button', { name: 'Previous chapter' }).click();
    await expect.poll(() => pane.evaluate(element => element.scrollTop)).toBeCloseTo(offset, 0);
    await page.getByRole('tab', { name: 'Play', exact: true }).click();
    await page.getByRole('tab', { name: 'Journal', exact: true }).click();
    await expect.poll(() => pane.evaluate(element => element.scrollTop)).toBeCloseTo(offset, 0);
    await page.reload();
    await page.getByRole('tab', { name: 'Journal', exact: true }).click();
    await expect.poll(() => pane.evaluate(element => element.scrollTop)).toBeCloseTo(offset, 0);
    await pane.evaluate(element => { element.scrollTop = 0; });
    await page.getByRole('link', { name: 'Read source history' }).click();
    await expect(page.getByText('The source outcome 1.', { exact: true })).toBeVisible();
    await expect(page.getByText('Source source-1 · Action action-1', { exact: true })).toBeVisible();
    expect(game.actions).toEqual([]);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth && document.documentElement.scrollHeight <= innerHeight + 1)).toBe(true);
  });
}
