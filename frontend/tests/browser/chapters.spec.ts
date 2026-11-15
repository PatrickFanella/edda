import { test, expect, signIn } from './fixtures';

test('Journal generates, retries and regenerates chapters while Play remains available', async ({ page, game }) => {
  const requests: { request_id: string; chapter_id?: string }[] = [];
  const revision = { id: 'revision-1', chapter_id: 'chapter', request_id: 'key', revision: 1, status: 'failed', content: '', error: 'Generation failed. Play remains available.', prompt_version: 'faithful-chapter-v1', created_at: '2026-01-01T00:00:00Z' };
  const chapter = { id: 'chapter', campaign_id: 'alpha', from_turn: 1, to_turn: 2, source_hash: 'fixture', source_turns: [], knowledge_boundary: 'player-visible-transcript-through-to-turn-v1', rules_version: 'srd-5.2.1', rules_mode: 'strict', created_at: revision.created_at, revisions: [revision] };
  let generated = false;
  await page.route('**/api/v1/campaigns/alpha/journal/chapters', async route => {
    if (route.request().method() === 'GET') return route.fulfill({ json: generated ? [chapter] : [] });
    const request = route.request().postDataJSON(); requests.push(request);
    if (!generated) { generated = true; revision.request_id = request.request_id; }
    else if (requests.length === 2) { revision.status = 'completed'; revision.content = 'The gate stayed shut.'; revision.error = ''; }
    else chapter.revisions.push({ ...revision, id: 'revision-2', revision: 2, request_id: request.request_id, content: 'Ada tried the gate. It stayed shut.' });
    return route.fulfill({ json: chapter.revisions.at(-1) });
  });
  await signIn(page); await page.goto('/play/alpha');
  await page.getByRole('tab', { name: 'Journal', exact: true }).click();
  await page.getByRole('button', { name: 'Generate next chapter' }).click();
  await expect(page.getByText('Generation failed. Play remains available.', { exact: true })).toBeVisible();
  await page.getByRole('tab', { name: 'Play', exact: true }).click();
  const input = page.getByLabel('Your move');
  await input.fill('Look at the gate'); await input.press('Enter');
  await expect.poll(() => game.actions).toContain('Look at the gate');
  await page.getByRole('tab', { name: 'Journal', exact: true }).click();
  await page.getByRole('button', { name: 'Retry revision 1' }).click();
  await expect(page.getByText('The gate stayed shut.', { exact: true })).toBeVisible();
  expect(requests[1].request_id).toBe(requests[0].request_id);
  await page.getByRole('button', { name: 'Regenerate chapter' }).click();
  await expect(page.getByText('Ada tried the gate. It stayed shut.', { exact: true })).toBeVisible();
  expect(requests[2].request_id).not.toBe(requests[0].request_id);
  expect(requests[2].chapter_id).toBe('chapter');
  await page.reload();
  await page.getByRole('tab', { name: 'Journal', exact: true }).click();
  await expect(page.getByText('The gate stayed shut.', { exact: true })).toBeVisible();
  await expect(page.getByText('Revision 2 · completed')).toBeVisible();
});
