import { test, expect, signIn } from './fixtures';
import { timestamp } from '../fixtures/data';
for (const width of [390, 1440]) {
 test(`reads long history without losing position at ${width}`, async ({ page, game }) => {
  await page.setViewportSize({ width, height: 900 });
  await page.route('**/campaigns/alpha/history', route => route.fulfill({ json: { entries: Array.from({ length: 300 }, (_, index) => ({ turn_number: index + 1, player_input: `Move ${index + 1}`, input_type: 'action', llm_response: `Scene ${index + 1}: The old road winds ahead.\n\nA keeper waits in silence.`, created_at: timestamp })) } }));
  await signIn(page); await page.goto('/play/alpha');
  await expect.poll(() => game.sockets.length).toBe(1);
  const log = page.getByRole('log');
  await expect(log.locator('article')).toHaveCount(600);
  const ids = await log.locator('article').evaluateAll(els => els.map(el => el.getAttribute('data-entry-id')));
  expect(new Set(ids).size).toBe(600);
  await expect.poll(() => log.evaluate(el => Math.abs(el.scrollHeight - el.scrollTop - el.clientHeight))).toBeLessThan(2);
  const send = page.getByRole('button', { name: 'Send action' }); const before = await send.boundingBox();
  await log.evaluate(el => { el.scrollTop = 100; el.dispatchEvent(new Event('scroll')); });
  await expect(page.getByRole('button', { name: 'Return to latest', exact: true })).toBeVisible();
  game.sockets[0].send(JSON.stringify({ type: 'result', timestamp, payload: { turn_number: 301, narrative: 'A new arrival waits.', state_changes: [], choices: [] } }));
  await expect(log.locator('article')).toHaveCount(601);
  expect(await log.evaluate(el => el.scrollTop)).toBe(100);
  expect((await send.boundingBox())!.y).toBe(before!.y);
  expect(await log.locator('article').evaluateAll(els => els.slice(0,600).map(el => el.getAttribute('data-entry-id')))).toEqual(ids);
  await page.getByRole('button', { name: /Return to latest/ }).click();
  await expect(page.getByText('A new arrival waits.', { exact: true })).toBeInViewport();
  expect(await page.evaluate(() => scrollY)).toBe(0);
 });
}
