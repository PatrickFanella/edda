import { test, expect, signIn } from './fixtures';

for (const width of [360, 390, 1366]) {
  test(`accepts visible completions into an editable draft at ${width}`, async ({ page, game }) => {
    await page.setViewportSize({ width, height: width === 1366 ? 768 : 844 });
    await page.route('**/api/v1/campaigns/alpha/action-suggestions', route => route.fulfill({ json: {
      version: 'suggestions-v1:fixture', candidates: [{ id: 'clue', text: 'Investigate the blue seal', reason: 'A known clue.', source_ids: ['known-fact:seal'] }],
    } }));
    await signIn(page); await page.goto('/play/alpha');
    const input = page.getByLabel('Your move');
    await input.fill('Inv');
    await expect(page.getByTestId('action-completion')).toBeVisible();
    await input.press('Shift+Tab'); await expect(input).not.toBeFocused();
    await input.focus(); await input.press('Escape');
    await expect(page.getByTestId('action-completion')).toHaveCount(0);
    await input.press('Tab'); await expect(input).not.toBeFocused();
    await input.fill('Inve'); await input.press('Tab');
    await expect(input).toBeFocused(); await expect(input).toHaveValue('Investigate the blue seal');
    expect(game.actions).toEqual([]);
    await input.press('End'); await input.press('Space'); await input.press('c');
    await expect(input).toHaveValue('Investigate the blue seal c');
    await page.getByRole('tab', { name: 'World', exact: true }).click();
    await page.getByRole('tab', { name: 'Play', exact: true }).click();
    await expect(input).toHaveValue('Investigate the blue seal c');
    await expect(page.getByRole('tab', { name: 'Play', exact: true })).toBeFocused();
    await expect(page.getByRole('button', { name: 'Send action' })).toBeInViewport();
    expect(await page.evaluate(() => document.documentElement.scrollHeight <= innerHeight + 1)).toBe(true);
    await input.press('Enter'); await expect.poll(() => game.actions).toEqual(['Investigate the blue seal c']);
  });
}

test.describe('touch completion', () => {
  test.use({ hasTouch: true, isMobile: true, viewport: { width: 390, height: 844 } });
  test('a phone tap inserts without submitting and retains input focus', async ({ page, game }) => {
    await page.route('**/api/v1/campaigns/alpha/action-suggestions', route => route.fulfill({ json: {
      version: 'suggestions-v1:fixture', candidates: [{ id: 'clue', text: 'Investigate the blue seal', reason: 'A known clue.', source_ids: ['known-fact:seal'] }],
    } }));
    await signIn(page); await page.goto('/play/alpha');
    await page.getByRole('button', { name: /Investigate the blue seal/ }).tap();
    await expect(page.getByLabel('Your move')).toHaveValue('Investigate the blue seal');
    await expect(page.getByLabel('Your move')).toBeFocused(); expect(game.actions).toEqual([]);
  });
});
