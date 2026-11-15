import { test, expect, signIn } from './fixtures';
import { character } from '../fixtures/data';
import fixture from '../fixtures/srd-check-sheet.json' with { type: 'json' };

for (const width of [1440, 390]) {
  test(`Fighter advancement requires an explicit HP choice and confirmation at ${width}px`, async ({ page, game }) => {
    await page.setViewportSize({ width, height: 844 });
    let level = 1, confirmations = 0;
    await page.route('**/api/v1/campaigns/alpha/character', route => route.fulfill({ json: { ...character(), rules_version: 'srd-5.2.1', level, sheet: { ...fixture.sheet, scope: `fighter-${level}-chain-mail-defense-fixture`, max_hp: { value: level === 1 ? 12 : 20, derivation: [] }, progression: { ...fixture.sheet.progression, current_xp: 300, next_level_xp: level === 1 ? 300 : 900 } } } }));
    await page.route('**/api/v1/campaigns/alpha/character/action-preview?*', route => {
      const params = new URL(route.request().url()).searchParams;
      expect(params.get('kind')).toBe('advance-level');
      expect(params.get('entry_id')).toBe('fixed');
      return route.fulfill({ json: { intent: { actor_id: 'hero', kind: 'advance-level', entry_id: 'fixed' }, rules_version: 'srd-5.2.1', expected_revision: 'advancement', available: level === 1, reason: 'Only level 1 Fighter advancement is supported.', targets: [], options: ['Fixed HP: 6 + Constitution; retain Defense and mastery. Action Surge and Tactical Mind effects unavailable.'], costs: [], cost_status: 'fixture-only', source_ids: [] } });
    });
    await page.route('**/api/v1/campaigns/alpha/action?*', route => route.fulfill({ json: null }));
    await page.route('**/api/v1/campaigns/alpha/action/confirm', route => {
      const request = route.request().postDataJSON();
      expect(request.intent).toEqual({ actor_id: 'hero', kind: 'advance-level', entry_id: 'fixed' });
      confirmations++; level = 2;
      return route.fulfill({ json: { action_id: request.action_id, turn_number: 3, narrative: 'Fighter level 2 saved. Maximum HP 12 to 20; current HP unchanged.', combat_active: false, state_changes: [], resolution_events: [{ type: "srd_advancement", label: "Fighter level 2", outcome: "advanced", details: { receipt: { rules_version: "srd-5.2.1", kind: "advance-level", choice: "fixed", die: 6, hp_gain: 8, level_before: 1, level_after: 2, max_hp_before: 12, max_hp_after: 20, source_ids: ["srd-5.2.1:content/character/level-advancement"] } } }] } });
    });
    await signIn(page); await page.goto('/play/alpha');
    await page.getByRole('tab', { name: 'Character', exact: true }).click();
    await page.getByLabel('Character action').selectOption('advance-level:fixed');
    await page.getByRole('button', { name: 'Preview action', exact: true }).click();
    await expect(page.getByText(/Fixed HP: 6/)).toBeVisible();
    expect(confirmations).toBe(0);
    await page.getByRole('button', { name: 'Confirm action', exact: true }).click();
    await expect(page.getByLabel('SRD character sheet')).toContainText('Level 2 Fighter');
    expect(confirmations).toBe(1);
    await expect(page.getByRole('tabpanel', { name: 'Character', exact: true }).getByLabel('Saved advancement receipt')).toContainText('d10 value 6');
    await page.reload(); await page.getByRole('tab', { name: 'Character', exact: true }).click();
    await expect(page.getByLabel('SRD character sheet')).toContainText('Level 2 Fighter');
    await expect(page.getByLabel('Character action').locator('option[value="advance-level:fixed"]')).toHaveCount(0);
    expect(confirmations).toBe(1);
    expect(game.actions).toEqual([]);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  });
}
