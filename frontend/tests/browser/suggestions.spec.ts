import { test, expect, signIn } from './fixtures';

for (const width of [390, 1440]) {
  test(`matches grounded suggestions without per-keystroke requests at ${width}`, async ({ page, game }) => {
    await page.setViewportSize({ width, height: 900 });
    let requests = 0;
    await page.route('**/api/v1/campaigns/alpha/action-suggestions', route => {
      requests++;
      return route.fulfill({ json: { version: `suggestions-v1:${requests}`, candidates: [{ id: 'known-clue', text: 'Investigate the blue seal', reason: 'You already know this clue. Its outcome is uncertain.', source_ids: ['known-fact:seal'] }] } });
    });
    await signIn(page); await page.goto('/play/alpha');
    const candidate = page.getByRole('button', { name: /Investigate the blue seal/ });
    await expect(candidate).toBeVisible();
    const initial = requests;
    await page.getByLabel('Your move').fill('I inv');
    await expect(candidate).toBeVisible();
    await candidate.click();
    await expect(page.getByLabel('Your move')).toHaveValue('Investigate the blue seal');
    expect(game.actions).toEqual([]); expect(requests).toBe(initial);
    await page.getByLabel('Your move').press('Enter');
    await expect.poll(() => game.actions).toContain('Investigate the blue seal');
    await expect.poll(() => requests).toBeGreaterThan(initial);
    expect(await page.evaluate(() => document.documentElement.scrollHeight <= innerHeight + 1)).toBe(true);
  });
}
test('assistance failure leaves ordinary play available', async ({ page, game }) => {
  await page.route('**/api/v1/campaigns/alpha/action-suggestions', route => route.fulfill({ status: 503, json: { error: 'assistance unavailable' } }));
  await signIn(page); await page.goto('/play/alpha');
  await page.getByLabel('Your move').fill('Wait by the gate');
  await page.getByLabel('Your move').press('Enter');
  await expect.poll(() => game.actions).toContain('Wait by the gate');
});
