import { test, expect, signIn } from './fixtures';
import { history } from '../fixtures/data';

const reminder = { scene: { text: 'North gate', source_id: 'known-location:gate' }, objective: { text: 'Find the keeper', source_id: 'quest:keeper' }, known_facts: [{ text: 'The blue seal is broken.', source_id: 'known-fact:seal' }] };
for (const width of [360, 390, 1366]) {
  test(`reminder opens and dismisses without disturbing draft or story at ${width}`, async ({ page, game }, testInfo) => {
    await page.setViewportSize({ width, height: width === 1366 ? 768 : 844 });
    await page.route('**/api/v1/campaigns/alpha/journal/reminder', route => route.fulfill({ json: reminder }));
    await page.route('**/campaigns/alpha/history', route => route.fulfill({ json: { entries: history().map(entry => ({ ...entry, llm_response: entry.llm_response + ' A long road.'.repeat(300) })) } }));
    await signIn(page); await page.goto('/play/alpha');
    await page.getByRole('button', { name: 'Where was I?' }).click();
    const panel = page.getByRole('region', { name: 'Return reminder' });
    await expect(panel).toContainText('North gate');
    await expect(panel).toContainText('Find the keeper');
    const content = page.getByRole('region', { name: 'Current situation' });
    const bounds = await content.boundingBox();
    expect(bounds).not.toBeNull();
    expect(bounds!.x).toBeGreaterThanOrEqual(0);
    expect(bounds!.x + bounds!.width).toBeLessThanOrEqual(width);
    await page.screenshot({ path: testInfo.outputPath('reminder-open.png') });
    const input = page.getByLabel('Your move'); await input.fill('My unsent action');
    await input.evaluate((el: HTMLTextAreaElement) => el.setSelectionRange(3, 10));
    const log = page.getByRole('tabpanel', { name: 'Play', exact: true }).getByRole('log');
    await expect(log).toBeVisible();
    await log.evaluate(el => { el.scrollTo({ top: 100, behavior: 'instant' }); el.dispatchEvent(new Event('scroll')); });
    await content.focus(); await content.press('Escape');
    await expect(page.getByRole('button', { name: 'Where was I?' })).toBeFocused();
    await page.getByRole('button', { name: 'Where was I?' }).click();
    await expect(input).toHaveValue('My unsent action');
    expect(await input.evaluate((el: HTMLTextAreaElement) => [el.selectionStart, el.selectionEnd])).toEqual([3, 10]);
    expect(await log.evaluate(el => el.scrollTop)).toBe(100);
    expect(game.actions).toEqual([]);
    await expect(page.getByRole('button', { name: 'Send action' })).toBeInViewport();
    expect(await page.evaluate(() => document.documentElement.scrollHeight <= innerHeight + 1)).toBe(true);
  });
}

test('a delayed or failed reminder never blocks play and refreshes after a completed turn', async ({ page, game }) => {
  let finish!: () => void; let requests = 0;
  await page.route('**/api/v1/campaigns/alpha/journal/reminder', async route => {
    requests++;
    if (requests === 1) {
      await new Promise<void>(resolve => { finish = resolve; });
      await route.fulfill({ status: 503, json: { error: 'private diagnostic' } });
    } else await route.fulfill({ json: reminder });
  });
  await signIn(page); await page.goto('/play/alpha');
  await page.getByRole('button', { name: 'Where was I?' }).click();
  await expect(page.getByText(/Recalling your situation/)).toBeVisible();
  const input = page.getByLabel('Your move'); await input.fill('Wait by the gate');
  await expect(input).toBeEnabled(); await expect(page.getByRole('button', { name: 'Send action' })).toBeEnabled();
  finish();
  await expect(page.getByRole('button', { name: 'Retry reminder' })).toBeVisible();
  await input.press('Enter'); await expect.poll(() => game.actions).toContain('Wait by the gate');
  await expect(page.getByRole('region', { name: 'Return reminder' })).toContainText('North gate');
  expect(requests).toBeGreaterThan(1);
  await page.goto('/play/beta');
  await expect(page.getByRole('region', { name: 'Return reminder' })).not.toContainText('North gate');
});
