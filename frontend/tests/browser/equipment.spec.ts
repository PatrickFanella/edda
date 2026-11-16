import { test, expect, signIn } from './fixtures';
import { character } from '../fixtures/data';
import fixture from '../fixtures/srd-check-sheet.json' with { type: 'json' };

for (const width of [1440, 390]) {
  test(`equipment preview, explicit confirmation and lost-response reload at ${width}px`, async ({ page, game }) => {
    await page.setViewportSize({ width, height: 844 });
    let equipped = true, confirmations = 0;
    let saved: Record<string, unknown> | null = null;
    await page.route('**/api/v1/campaigns/alpha/character', route => route.fulfill({ json: { ...character(), rules_version: 'srd-5.2.1', sheet: { ...fixture.sheet, defenses: { ...fixture.sheet.defenses, 'armor-class': { value: equipped ? 19 : 13, derivation: [] } } } } }));
    await page.route('**/api/v1/campaigns/alpha/character/inventory', route => route.fulfill({ json: [{ id: 'mail', name: 'Chain Mail', item_type: 'armor', rarity: 'common', description: 'Ordinary fixture', quantity: 1, equipped, properties: { srd521_id: 'chain-mail' } }] }));
    await page.route('**/api/v1/campaigns/alpha/character/action-preview?*', route => {
      const params = new URL(route.request().url()).searchParams;
      expect(params.get('kind')).toBe('equipment-remove'); expect(params.get('entry_id')).toBe('mail');
      return route.fulfill({ json: { intent: { actor_id: 'hero', kind: 'equipment-remove', entry_id: 'mail' }, rules_version: 'srd-5.2.1', expected_revision: 'equipment-fixture', available: true, targets: [], options: ['Remove chain-mail; AC 19 to 13; 300 seconds of uninterrupted downtime.'], costs: [{ resource_id: 'campaign-seconds', amount: 300 }], cost_status: 'fixture-only', source_ids: [] } });
    });
    await page.route('**/api/v1/campaigns/alpha/action?*', route => route.fulfill({ json: saved ? { result: saved } : null }));
    await page.route('**/api/v1/campaigns/alpha/action/confirm', async route => {
      const request = route.request().postDataJSON();
      expect(request.intent).toEqual({ actor_id: 'hero', kind: 'equipment-remove', entry_id: 'mail' });
      confirmations++; equipped = false;
      saved = { action_id: request.action_id, turn_number: 3, narrative: 'Chain mail removed; AC 19 to 13; 300 seconds elapsed.', combat_active: false, state_changes: [], resolution_events: [{ type: 'srd_equipment', label: 'Equipment change', outcome: 'equipment-remove', details: { receipt: { rules_version: 'srd-5.2.1', kind: 'equipment-remove', item_id: 'mail', equipment: 'chain-mail', equipped_before: true, equipped_after: false, ac_before: 19, ac_after: 13, time_before: 28800, time_after: 29100, source_ids: [] } } }] };
      // Commit succeeded, response was lost. The retained identity reconciles on reload.
      await route.abort('failed');
    });
    await signIn(page); await page.goto('/play/alpha');
    await page.getByRole('tab', { name: 'Character', exact: true }).click();
    await page.getByLabel('Character action').selectOption('equipment-remove:mail');
    await page.getByRole('button', { name: 'Preview action', exact: true }).click();
    await expect(page.getByText(/Remove chain-mail; AC 19 to 13/)).toBeVisible();
    expect(confirmations).toBe(0);
    await page.getByRole('button', { name: 'Confirm action', exact: true }).click();
    await expect(page.getByRole('alert')).toContainText('Retry the same action');
    await page.reload(); await page.getByRole('tab', { name: 'Character', exact: true }).click();
    await expect(page.getByRole('tabpanel', { name: 'Character', exact: true }).getByLabel('Saved equipment receipt')).toContainText('AC 19 to 13');
    await expect(page.getByLabel('Character action').locator('option[value="equipment-equip:mail"]')).toHaveCount(1);
    expect(confirmations).toBe(1); expect(game.actions).toEqual([]);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  });
}
