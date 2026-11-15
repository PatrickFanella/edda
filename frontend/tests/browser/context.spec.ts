import { test, expect, signIn } from './fixtures';
import { history } from '../fixtures/data';
for (const width of [390, 1366]) {
 test(`context restores reading position and focus at ${width}`, async ({ page, game }) => {
  await page.setViewportSize({ width, height: 844 });
  await page.route('**/campaigns/alpha/history', route => route.fulfill({ json: { entries: history().map(entry => ({ ...entry, llm_response: entry.llm_response + ' A long road.'.repeat(500) })) } }));
  await signIn(page); await page.goto('/play/alpha');
  await expect.poll(() => game.sockets.length).toBe(1);
  const story = page.getByRole('tabpanel', { name: 'Play', exact: true });
  const log = story.getByRole('log');
  await expect(log).toBeVisible();
  await expect.poll(() => log.evaluate(el => Math.abs(el.scrollHeight - el.clientHeight - el.scrollTop))).toBeLessThan(20);
  await log.evaluate(el => { el.scrollTo({ top: 100, behavior: 'instant' }); el.dispatchEvent(new Event('scroll')); });
  const input = page.getByLabel('Your move'); await input.fill('My retained action'); await input.focus();
  await page.getByRole('tab', { name: 'Character', exact: true }).click();
  if (width >= 1280) await expect(page.getByRole('region', { name: 'Story', exact: true })).toBeVisible(); else await expect(story).toBeHidden();
  await expect(page.getByRole('region', { name: 'Inventory', exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'Close details' }).click();
  await expect(input).toHaveValue('My retained action'); await expect(input).toBeFocused();
  expect(await log.evaluate(el => el.scrollTop)).toBe(100);
  await page.getByRole('tab', { name: 'World', exact: true }).click();
  await expect(page.getByRole('region', { name: 'Quests', exact: true })).toBeVisible();
  await expect(page.getByRole('region', { name: 'NPCs', exact: true })).toBeVisible();
  await page.getByRole('tabpanel', { name: 'World', exact: true }).press('Escape');
  await expect(page.getByRole('tab', { name: 'Play', exact: true })).toHaveAttribute('aria-selected', 'true');
 });
}
