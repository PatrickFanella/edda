import { test, expect, signIn } from './fixtures';
import { character } from '../fixtures/data';
import fixture from '../fixtures/srd-check-sheet.json' with { type: 'json' };

for (const width of [1440, 390]) {
  test(`Tactical Mind explicit confirm and lost-response reconnect ${width}px`, async ({ page, game }) => {
    await page.setViewportSize({ width, height: 844 });
    let confirmations = 0;
    let saved: Record<string, unknown> | null = null;
    await page.route('**/api/v1/campaigns/alpha/character', route => route.fulfill({ json: { ...character(), rules_version: 'srd-5.2.1', level: 2, sheet: { ...fixture.sheet, scope: 'fighter-2-chain-mail-defense-fixture' } } }));
    await page.route('**/api/v1/campaigns/alpha/character/action-preview?*', route => {
      const params = new URL(route.request().url()).searchParams;
      expect(params.get('kind')).toBe('tactical-mind');
      expect(params.get('entry_id')).toBe('use');
      return route.fulfill({ json: { intent: { actor_id: 'hero', kind: 'tactical-mind', entry_id: 'use' }, rules_version: 'srd-5.2.1', expected_revision: 'tactical-v1:fixture', available: true, targets: [], options: ['Apply one d10 to the saved failed check; Second Wind only on success.'], costs: [], cost_status: 'conditional-on-success', source_ids: ['srd-5.2.1:content/fighter/tactical-mind'] } });
    });
    await page.route('**/api/v1/campaigns/alpha/action?*', route => route.fulfill({ json: saved ? { result: saved } : null }));
    await page.route('**/api/v1/campaigns/alpha/action/confirm', async route => {
      const request = route.request().postDataJSON();
      expect(request.intent).toEqual({ actor_id: 'hero', kind: 'tactical-mind', entry_id: 'use' });
      confirmations++;
      saved = { action_id: request.action_id, turn_number: 2, narrative: 'Tactical Mind saved: success. Original roll retained.', combat_active: false, state_changes: [], resolution_events: [{ type: 'srd_tactical_mind', label: 'Tactical Mind', outcome: 'success', details: { receipt: { base: { ...fixture.receipts[0], outcome: 'failure' }, die: 1, total: 10, outcome: 'success', second_wind_before: 2, second_wind_after: 1, source_ids: ['srd-5.2.1:content/fighter/tactical-mind'] } } }] };
      await route.abort('failed');
    });
    await signIn(page); await page.goto('/play/alpha');
    await page.getByRole('tab', { name: 'Character', exact: true }).click();
    await page.getByLabel('Character action').selectOption('tactical-mind:use');
    await page.getByRole('button', { name: 'Preview action', exact: true }).click();
    await expect(page.getByText(/Apply one d10/)).toBeVisible();
    expect(confirmations).toBe(0);
    await page.getByRole('button', { name: 'Confirm action', exact: true }).click();
    await expect(page.getByRole('alert')).toContainText('Retry the same action');
    await page.reload();
    await page.getByRole('tab', { name: 'Character', exact: true }).click();
    await expect(page.getByRole('tabpanel', { name: 'Character', exact: true }).getByLabel('Saved Tactical Mind receipt')).toContainText('Second Wind: 2 → 1');
    expect(confirmations).toBe(1);
    expect(game.actions).toEqual([]);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  });
}
