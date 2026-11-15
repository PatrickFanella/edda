import { test, expect, signIn } from './fixtures';
import { character } from '../fixtures/data';
import fixture from '../fixtures/srd-check-sheet.json' with { type: 'json' };

for (const viewport of [{ width: 1440, height: 900 }, { width: 390, height: 844 }]) {
  test(`prepares a sheet action without performing it at ${viewport.width}px`, async ({ page, game }) => {
    await page.setViewportSize(viewport);
    await page.route('**/api/v1/campaigns/alpha/character', route => route.fulfill({ json: { ...character(), rules_version: 'srd-5.2.1', sheet: fixture.sheet } }));
    const previews: URL[] = [];
    await page.route('**/api/v1/campaigns/alpha/character/action-preview?*', route => {
      const url = new URL(route.request().url());
      previews.push(url);
      expect(route.request().method()).toBe('GET');
      return route.fulfill({ json: {
        intent: { actor_id: 'hero', kind: 'ability-check', ability: 'strength', skill: 'athletics' },
        rules_version: 'srd-5.2.1', expected_revision: 'actor-v1:fixture', available: true,
        modifier: fixture.sheet.skills.athletics, targets: [], options: ['ordinary'], costs: [],
        cost_status: 'requires-adjudication', source_ids: ['srd-5.2.1:content/d20-tests/ability-checks'],
      } });
    });
    await page.route('**/api/v1/campaigns/alpha/action?*', route => route.fulfill({ json: null }));
    await page.route('**/api/v1/campaigns/alpha/action/confirm', route => {
      const request = route.request().postDataJSON();
      return route.fulfill({ json: { action_id: request.action_id, turn_number: 3, narrative: 'The ordinary check is recorded.', combat_active: false, state_changes: [], resolution_events: [] } });
    });
    await signIn(page);
    await page.goto('/play/alpha');
    await page.getByRole('tab', { name: 'Character', exact: true }).click();
    await page.getByLabel('Character action').selectOption('skill:athletics');
    await page.getByRole('button', { name: 'Preview action', exact: true }).click();
    await expect(page.getByText('Preview ready. No action has been performed.')).toBeVisible();
    expect(previews).toHaveLength(1);
    expect(previews[0].searchParams.get('skill')).toBe('athletics');
    expect(previews[0].searchParams.get('actor_id')).toBe('hero');
    expect(game.actions).toEqual([]);
    await expect(page.getByText('Modifier: 5')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Confirm action', exact: true })).toBeEnabled();
    await page.getByRole('button', { name: 'Confirm action', exact: true }).click();
    await expect(page.getByText('Action saved.')).toBeVisible();
    await page.getByRole('tab', { name: 'Play', exact: true }).click();
    await expect(page.getByText('The ordinary check is recorded.', { exact: true })).toBeVisible();
    await expect(page.getByText('Confirm athletics check', { exact: true })).toBeVisible();
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  });
}
