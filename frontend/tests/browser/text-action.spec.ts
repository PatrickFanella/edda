import { expect, signIn, test } from './fixtures';
import { campaign, character } from '../fixtures/data';
import fixture from '../fixtures/srd-check-sheet.json' with { type: 'json' };

const athletics = {
  intent: { actor_id: 'hero', kind: 'ability-check', skill: 'athletics', ability: 'strength' },
  rules_version: 'srd-5.2.1', expected_revision: 'athletics-v1', available: true,
  modifier: fixture.sheet.skills.athletics, targets: [], options: ['ordinary'], costs: [],
  cost_status: 'requires-adjudication', source_ids: ['srd-5.2.1:content/d20-tests/ability-checks'],
};

for (const viewport of [{ width: 1440, height: 900, earlyFrame: false }, { width: 390, height: 844, earlyFrame: false }, { width: 390, height: 844, earlyFrame: true }]) {
  test(`prepares the exact SRD Athletics text at ${viewport.width}px without websocket fallback${viewport.earlyFrame ? ' when a frame precedes the panel commit' : ''}`, async ({ page, game }) => {
    if (viewport.earlyFrame) {
      // A frame may run before React commits the newly selected panel. Model
      // that ordering without relying on CPU load or wall-clock timing.
      await page.addInitScript(() => {
        const requestFrame = window.requestAnimationFrame.bind(window);
        window.requestAnimationFrame = callback => {
          if (document.querySelector('[role="tabpanel"][id$="-panel-character"][hidden]')) {
            queueMicrotask(() => callback(performance.now()));
            return 0;
          }
          return requestFrame(callback);
        };
      });
    }
    await page.setViewportSize(viewport);
    await page.route('**/api/v1/campaigns/alpha', route => route.fulfill({ json: { ...campaign(), rules_mode: 'crunch', rules_version: 'srd-5.2.1' } }));
    await page.route('**/api/v1/campaigns/alpha/character', route => route.fulfill({ json: { ...character(), rules_version: 'srd-5.2.1', sheet: fixture.sheet } }));
    await page.route('**/api/v1/campaigns/alpha/character/action-preview?*', route => route.fulfill({ json: { intent: { actor_id: 'hero', kind: 'combat-end-turn' }, rules_version: 'srd-5.2.1', available: false, targets: [], options: [], costs: [], cost_status: 'encounter', source_ids: [] } }));
    let confirms = 0;
    await page.route('**/api/v1/campaigns/alpha/character/action-preview', route => {
      expect(route.request().method()).toBe('POST');
      expect(route.request().postDataJSON()).toEqual({ actor_id: 'hero', input: 'Make an Athletics check' });
      return route.fulfill({ headers: { 'Cache-Control': 'no-store' }, json: athletics });
    });
    await page.route('**/api/v1/campaigns/alpha/action**', route => route.fulfill({ json: null }));
    await page.route('**/api/v1/campaigns/alpha/action/confirm', route => {
      confirms++;
      expect(route.request().postDataJSON()).toMatchObject({ intent: athletics.intent });
      return route.fulfill({ json: { action_id: 'saved', turn_number: 3, narrative: 'The Athletics check is recorded.', combat_active: false, state_changes: [], resolution_events: [] } });
    });
    await signIn(page); await page.goto('/play/alpha');
    const input = page.getByLabel('Your move');
    await input.fill('Make an Athletics check');
    await expect(page.getByRole('button', { name: 'Prepare action', exact: true })).toBeEnabled();
    await input.press('Enter');
    await expect(page.getByRole('tab', { name: 'Character', exact: true })).toHaveAttribute('aria-selected', 'true');
    await expect(page.getByRole('tabpanel', { name: 'Character', exact: true })).toBeFocused();
    await expect(page.getByText('Preview ready. No action has been performed.')).toBeVisible();
    expect(game.actions).toEqual([]); expect(confirms).toBe(0);
    const confirm = page.getByRole('button', { name: 'Confirm action', exact: true });
    await confirm.scrollIntoViewIfNeeded();
    expect(await confirm.evaluate(node => { const box = node.getBoundingClientRect(); return box.left >= 0 && box.right <= innerWidth && box.top >= 0 && box.bottom <= innerHeight; })).toBe(true);
    await page.getByRole('button', { name: 'Confirm action', exact: true }).click();
    await expect(page.getByText('Action saved.')).toBeVisible();
    expect(confirms).toBe(1);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth && document.documentElement.scrollHeight >= innerHeight)).toBe(true);
  });
}

test('keeps unsupported SRD text in the draft and never sends it to the websocket', async ({ page, game }) => {
  await page.route('**/api/v1/campaigns/alpha', route => route.fulfill({ json: { ...campaign(), rules_mode: 'crunch', rules_version: 'srd-5.2.1' } }));
  await page.route('**/api/v1/campaigns/alpha/character', route => route.fulfill({ json: { ...character(), rules_version: 'srd-5.2.1', sheet: fixture.sheet } }));
  await page.route('**/api/v1/campaigns/alpha/character/action-preview?*', route => route.fulfill({ json: { intent: { actor_id: 'hero', kind: 'combat-end-turn' }, rules_version: 'srd-5.2.1', available: false, targets: [], options: [], costs: [], cost_status: 'encounter', source_ids: [] } }));
  await page.route('**/api/v1/campaigns/alpha/character/action-preview', route => route.fulfill({ status: 400, json: { error: 'Use the exact supported text: Make an Athletics check.' } }));
  await page.route('**/api/v1/campaigns/alpha/action**', route => route.fulfill({ json: null }));
  await signIn(page); await page.goto('/play/alpha');
  const input = page.getByLabel('Your move');
  await input.fill('Make an athletics check'); await input.press('Enter');
  await expect(page.getByText('Use the exact supported text: Make an Athletics check.', { exact: true })).toBeVisible();
  await expect(input).toHaveValue('Make an athletics check');
  expect(game.actions).toEqual([]);
});
