import { test, expect, signIn } from './fixtures';
import { character } from '../fixtures/data';
import fixture from '../fixtures/srd-check-sheet.json' with { type: 'json' };
import { expectDocumentFits, expectReachable } from './layout';

for (const width of [1440, 390]) {
 for (const weapon of ['Scimitar', 'Light Crossbow']) {
  test(`Bandit ${weapon} confirmation and lost-response reconnect at ${width}px`, async ({ page, game }) => {
   await page.setViewportSize({ width, height: 844 });
   let confirmations = 0;
   let saved: Record<string, unknown> | null = null;
   await page.route('**/api/v1/campaigns/alpha/character', route => route.fulfill({ json: { ...character(), id: 'hero', rules_version: 'srd-5.2.1', sheet: fixture.sheet } }));
   await page.route('**/api/v1/campaigns/alpha/character/action-preview?*', route => route.fulfill({ json: { intent: { actor_id: 'hero', kind: 'combat-end-turn' }, rules_version: 'srd-5.2.1', expected_revision: 'bandit-fixture', available: true, targets: [], options: [`End turn resolves one Bandit ${weapon} attack +3.`], costs: [], cost_status: 'encounter', source_ids: ['srd-5.2.1:content/monsters/bandit'] } }));
   await page.route('**/api/v1/campaigns/alpha/action?*', route => route.fulfill({ json: saved ? { result: saved } : null }));
   await page.route('**/api/v1/campaigns/alpha/action/confirm', async route => {
    confirmations++;
    expect(route.request().postDataJSON().intent).toEqual({ actor_id: 'hero', kind: 'combat-end-turn' });
    saved = { action_id: route.request().postDataJSON().action_id, turn_number: 3, narrative: `${weapon} hits for 4 damage.`, combat_active: false, state_changes: [], resolution_events: [{ type: 'srd_attack', label: `NPC ${weapon}`, outcome: 'hit', details: { receipt: { rules_version: 'srd-5.2.1', kind: 'martial-attack', weapon, target: 'Mira', target_id: 'hero', die: 19, modifier: { value: 3, derivation: [{ label: 'Bandit weapon attack', value: 3, source_id: 'srd-5.2.1:content/monsters/bandit' }] }, total: 22, outcome: 'hit', damage_dice: [3], damage_modifier: 1, damage: 4, damage_type: weapon === 'Scimitar' ? 'slashing' : 'piercing', source_ids: ['srd-5.2.1:content/monsters/bandit'] } } }] };
    await route.abort('failed');
   });
   await signIn(page); await page.goto('/play/alpha');
   await page.getByRole('tab', { name: 'Character', exact: true }).click();
   await page.getByLabel('Character action').selectOption('combat-end-turn');
   await page.getByRole('button', { name: 'Preview action', exact: true }).click();
   await expect(page.getByText(/Roll options: End turn resolves one Bandit/)).toBeVisible();
   expect(confirmations).toBe(0);
   const confirm = page.getByRole('button', { name: 'Confirm action', exact: true });
   await confirm.scrollIntoViewIfNeeded(); await expectReachable(confirm); await confirm.click();
   await expect(page.getByRole('alert')).toContainText('Retry the same action');
   await page.reload(); await page.getByRole('tab', { name: 'Character', exact: true }).click();
   const receipt = page.getByRole('region', { name: 'Saved action result' });
   const summary = receipt.getByText(`${weapon} against Mira: hit · 4`, { exact: false });
   await expect(summary).toBeVisible(); await summary.click();
   await expect(receipt).toContainText('Damage: 3 + 1 = 4');
   await expect(receipt).not.toContainText('d8 3');
   await expect(receipt).not.toContainText('hp');
   await expectDocumentFits(page);
   await expect.poll(() => game.sockets.length).toBe(2);
   await game.sockets[1].close({ code: 1012, reason: 'Bandit fixture reconnect' });
   await expect.poll(() => game.sockets.length).toBe(3);
   await expect(receipt).toContainText('Damage: 3 + 1 = 4');
   expect(confirmations).toBe(1); expect(game.actions).toEqual([]);
  });
 }
}
