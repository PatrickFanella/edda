import { test, expect, signIn } from './fixtures';
import { character } from '../fixtures/data';
import fixture from '../fixtures/srd-check-sheet.json' with { type: 'json' };

for (const width of [1440, 390]) {
  test(`previews concentration and rest costs before confirming at ${width}px`, async ({ page, game }) => {
    await page.setViewportSize({ width, height: 844 });
    let slots = 2, confirmations = 0;
    let concentration: { id: string; spell: string; started_at: number; expires_at: number } | undefined;
    const spellcasting = { ability: 'wisdom', modifier: { value: 3, derivation: [] }, spells: [{ id: 'shield-of-faith', label: 'Shield of Faith', known: true, prepared: true, playable: true, components: ['verbal', 'somatic', 'Holy Symbol focus'], components_available: true, slot_levels: [1], source_id: 'spell' }] };
    await page.route('**/api/v1/campaigns/alpha/character', route => route.fulfill({ json: { ...character(), hp: 10, max_hp: 10, rules_version: 'srd-5.2.1', sheet: { ...fixture.sheet, spellcasting, concentration, resources: [{ id: 'spell-slot-1', label: 'Level 1 spell slots', current: slots, maximum: 2, source_id: 'spell' }], capabilities: [] } } }));
    await page.route('**/api/v1/campaigns/alpha/character/action-preview?*', route => {
      const params = new URL(route.request().url()).searchParams;
      const kind = params.get('kind');
      const cast = kind === 'spell-cast';
      if (cast) expect(params.get('slot_level')).toBe('1');
      return route.fulfill({ json: { intent: { actor_id: 'hero', kind, ...(cast ? { entry_id: 'shield-of-faith', slot_level: 1 } : {}) }, rules_version: 'srd-5.2.1', expected_revision: 'lifecycle-fixture', available: true, targets: cast ? [{ id: 'hero', label: 'Mira (self)' }] : [], options: [cast ? '+2 AC; concentration for 10 minutes; replaces current effect' : 'Eight uninterrupted hours including six hours of sleep; recover level 1 slots'], costs: cast ? [{ resource_id: 'spell-slot-1', amount: 1 }, { resource_id: 'bonus-action', amount: 1 }] : [], cost_status: 'fixture-only', source_ids: [] } });
    });
    await page.route('**/api/v1/campaigns/alpha/action?*', route => route.fulfill({ json: null }));
    await page.route('**/api/v1/campaigns/alpha/action/confirm', route => {
      const request = route.request().postDataJSON();
      confirmations++;
      const cast = request.intent.kind === 'spell-cast';
      const before = concentration;
      if (cast) { expect(request.target_id).toBe('hero'); slots = 1; concentration = { id: request.action_id, spell: 'shield-of-faith', started_at: 28800, expires_at: 29400 }; }
      else { expect(request.intent.kind).toBe('rest-long'); slots = 2; concentration = undefined; }
      const reason = cast ? 'Shield of Faith saved.' : 'Long Rest saved; concentration ended.';
      return route.fulfill({ json: { action_id: request.action_id, turn_number: confirmations + 2, narrative: reason, combat_active: false, state_changes: [], resolution_events: [{ type: 'srd_effect', label: 'Caster effects and recovery', outcome: reason, details: { receipt: { rules_version: 'srd-5.2.1', kind: request.intent.kind, reason, before, after: concentration, time_before: 28800, time_after: cast ? 28800 : 57600, slots_before: cast ? 2 : 1, slots_after: slots, hp_before: 10, hp_after: 10, hit_dice_before: 1, hit_dice_after: 1, source_ids: [] } } }] } });
    });
    await signIn(page); await page.goto('/play/alpha');
    await page.getByRole('tab', { name: 'Character', exact: true }).click();
    await page.getByLabel('Character action').selectOption('spell-cast:shield-of-faith');
    await page.getByRole('button', { name: 'Preview action', exact: true }).click();
    await expect(page.getByRole('button', { name: 'Confirm action', exact: true })).toBeDisabled();
    await page.getByLabel('Spell target').selectOption('hero');
    expect(confirmations).toBe(0); expect(game.actions).toEqual([]);
    await page.getByRole('button', { name: 'Confirm action', exact: true }).click();
    await expect(page.getByText('Level 1 spell slots: 1 / 2')).toBeVisible();
    await expect(page.getByLabel('Concentration and recovery', { exact: true })).toContainText('expires Day 1, 08:10:00');
    await page.reload(); await page.getByRole('tab', { name: 'Character', exact: true }).click();
    await expect(page.getByLabel('Concentration and recovery', { exact: true })).toContainText('expires Day 1, 08:10:00');
    await page.getByLabel('Character action').selectOption('rest-long');
    await page.getByRole('button', { name: 'Preview action', exact: true }).click();
    await expect(page.getByText(/Eight uninterrupted hours including six hours of sleep/)).toBeVisible();
    expect(confirmations).toBe(1);
    await page.getByRole('button', { name: 'Confirm action', exact: true }).click();
    await expect(page.getByText('Level 1 spell slots: 2 / 2')).toBeVisible();
    await expect(page.getByLabel('Concentration and recovery', { exact: true })).toContainText('No concentration effect active.');
    await expect(page.getByRole('tabpanel', { name: 'Character', exact: true }).getByLabel('Saved caster effect receipt')).toContainText('Day 1, 08:00:00 → Day 1, 16:00:00');
    expect(confirmations).toBe(2);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  });
}
