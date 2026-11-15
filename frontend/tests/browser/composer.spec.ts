import { test, expect, signIn } from './fixtures';
import { timestamp } from '../fixtures/data';
for (const width of [360, 390, 1366]) {
 test(`retains drafts and selection across details at ${width}`, async ({ page, game }) => {
  await page.setViewportSize({ width, height: width === 1366 ? 768 : 844 });
  await signIn(page); await page.goto('/play/alpha');
  await expect.poll(() => game.sockets.length).toBe(1);
  const input = page.getByLabel('Your move');
  const draft = 'An unsent draft\nwith a second line. ' + 'Read carefully. '.repeat(100);
  await input.fill(draft);
  await input.evaluate((el: HTMLTextAreaElement) => el.setSelectionRange(3, 15));
  const choices = Array.from({ length: 8 }, (_, i) => ({ id: `choice-${i}`, text: `Option ${i + 1}: ${'The gate waits. '.repeat(40)}` }));
  game.sockets[0].send(JSON.stringify({ type: 'result', timestamp, payload: { narrative: 'The keeper waits.', state_changes: [], choices } }));
  await page.getByRole('tab', { name: 'Character', exact: true }).click();
  await page.getByRole('tab', { name: 'Narrative', exact: true }).click();
  await expect(input).toHaveValue(draft);
  expect(await input.evaluate((el: HTMLTextAreaElement) => [el.selectionStart, el.selectionEnd])).toEqual([3, 15]);
  await expect(page.getByRole('button', { name: 'Send action' })).toBeInViewport();
  expect(await page.evaluate(() => document.documentElement.scrollHeight)).toBeLessThanOrEqual(await page.evaluate(() => innerHeight));
  await page.getByRole('region', { name: 'Suggested choices' }).getByRole('button').last().click();
  await expect(page.getByText('You find a silver key.', { exact: true })).toBeVisible();
 });
}
