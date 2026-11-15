import { test, expect, signIn, user } from './fixtures';

test('creation hands the opening scene to live play and completes a turn', async ({ page, game }) => {
  await signIn(page);
  await page.goto('/new');
  await page.getByRole('button', { name: /Campaign interview/ }).click();
  await page.getByRole('button', { name: 'Continue', exact: true }).click();
  await page.getByLabel('Your campaign notes').fill('A hopeful frontier');
  await page.getByRole('button', { name: 'Send reply' }).click();
  await page.getByRole('button', { name: /Pure storytelling/ }).click();
  await page.getByRole('button', { name: 'Continue', exact: true }).click();
  await page.getByRole('button', { name: /Character interview/ }).click();
  await page.getByRole('button', { name: 'Continue', exact: true }).click();
  await page.getByLabel('Your character notes').fill('Mira the scout');
  await page.getByRole('button', { name: 'Send reply' }).click();
  await page.getByRole('button', { name: 'Begin adventure' }).click();
  await expect(page).toHaveURL(/play\/alpha$/);
  await expect(page.getByText('A lantern lights the road.', { exact: true })).toBeVisible();
  await expect.poll(() => game.sockets.length).toBe(1);
  await page.getByLabel('Your move').fill('Search the road');
  await page.getByLabel('Your move').press('Enter');
  await expect(page.getByText('You find a silver key.', { exact: true })).toBeVisible();
  await expect(page.getByLabel('Your move')).toBeEnabled();
  expect(game.actions).toEqual(['Search the road']);
  expect(game.requests).toContain('POST /campaigns/start/world');
});

test('reconnects and accepts a subsequent action without duplicating it', async ({ page, game }) => {
  await signIn(page); await page.goto('/play/alpha');
  await expect.poll(() => game.sockets.length).toBe(1);
  await game.sockets[0].close({ code: 1012, reason: 'fixture interruption' });
  await expect.poll(() => game.sockets.length).toBe(2);
  await page.getByLabel('Your move').fill('Look around');
  await page.getByLabel('Your move').press('Enter');
  await expect(page.getByText('You find a silver key.', { exact: true })).toBeVisible();
  expect(game.actions).toEqual(['Look around']);
});

test('switches campaigns through the list and replays only the selected history', async ({ page, game }) => {
  await signIn(page); await page.goto('/play/alpha');
  await expect(page.getByText('The alpha gate opens.', { exact: true })).toBeVisible();
  await page.getByRole('button', { name: /Fixture Player/ }).click();
  await page.getByRole('link', { name: /Back to campaigns/i }).click();
  await page.locator('a[href="/play/beta"]').click();
  await expect(page.getByText('The beta gate opens.', { exact: true })).toBeVisible();
  await expect(page.getByText('The alpha gate opens.', { exact: true })).toHaveCount(0);
  expect(game.sockets.at(-1)?.url()).toContain('/beta/ws');
  await page.goto('/replay/beta');
  await expect(page.getByText('The beta gate opens.', { exact: true })).toBeVisible();
  await expect(page.getByText('The beta keeper welcomes you.', { exact: true })).toHaveCount(0);
  await page.getByRole('button', { name: 'Next turn' }).click();
  await expect(page.getByText('The beta keeper welcomes you.', { exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'Previous turn' }).click();
  await expect(page.getByText('The beta keeper welcomes you.', { exact: true })).toHaveCount(0);
});

test('registers, restores a session, signs out and rejects an expired session', async ({ page, game }) => {
  await page.goto('/play/alpha'); await expect(page).toHaveURL(/login$/);
  await page.goto('/register');
  await page.getByLabel('Name', { exact: true }).fill(user.name);
  await page.getByLabel('Email').fill(user.email);
  await page.getByLabel('Password', { exact: true }).fill('fixture-password');
  await page.getByRole('button', { name: /Create account/i }).click();
  await expect(page).toHaveURL('/');
  await page.reload();
  await expect(page.getByRole('heading', { name: 'Campaigns', exact: true })).toBeVisible();
  expect(game.requests).toContain('GET /auth/me');
  await page.getByRole('button', { name: /Fixture Player/ }).click();
  await page.getByRole('button', { name: 'Sign out' }).click();
  await expect(page).toHaveURL(/login$/);
  expect(await page.evaluate(() => localStorage.getItem('gm_token'))).toBeNull();
  await signIn(page);
  await page.route('**/auth/me', route => route.fulfill({ status: 401, json: { error: 'Expired' } }));
  await page.reload(); await expect(page).toHaveURL(/login$/);
  expect(await page.evaluate(() => localStorage.getItem('gm_token'))).toBeNull();
});
