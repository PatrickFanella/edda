import { test, expect, signIn } from './fixtures';
import { character } from '../fixtures/data';
import fixture from '../fixtures/srd-check-sheet.json' with { type: 'json' };

for (const width of [1440, 390]) for (const choice of ['use', 'decline', 'base-success']) {
  test(`select Easy Athletics then ${choice}, explicit confirmation and reconnect ${width}px`, async ({ page, game }) => {
    await page.setViewportSize({ width, height: 844 });
    let baseConfirmed = false, consumed = false, baseCalls = 0, tacticalCalls = 0;
    const saved = new Map<string, unknown>();
    const baseReceipt = { ...fixture.receipts[0], ability: 'strength', skill: 'athletics', mode: 'ordinary', dice: [{ sides: 20, value: choice === 'base-success' ? 5 : 4, kept: true }], total: choice === 'base-success' ? 10 : 9, dc: 10, outcome: choice === 'base-success' ? 'success' : 'failure' };
    await page.route('**/api/v1/campaigns/alpha/character', route => route.fulfill({ json: { ...character(), rules_version: 'srd-5.2.1', level: 2, sheet: { ...fixture.sheet, scope: 'fighter-2-chain-mail-defense-fixture' } } }));
    await page.route('**/api/v1/campaigns/alpha/character/action-preview?*', route => {
      const params = new URL(route.request().url()).searchParams;
      const isBase = params.get('kind') === 'ability-check';
      expect(params.get('entry_id')).toBe(isBase ? 'easy-strength-athletics-dc10' : choice === 'decline' ? 'decline' : 'use');
      const intent = { actor_id: 'hero', kind: isBase ? 'ability-check' : 'tactical-mind', entry_id: params.get('entry_id') };
      return route.fulfill({ json: { intent, rules_version: 'srd-5.2.1', expected_revision: isBase ? 'easy-v1' : 'tactical-v1', available: isBase || (baseConfirmed && !consumed && choice !== 'base-success'), reason: 'No eligible saved failed check.', targets: [], options: [isBase ? 'Edda bounded adjudication fixture: Easy Strength (Athletics), disclosed fixed DC 10; one ordinary d20.' : 'Apply one d10 to the saved failed check; spend Second Wind only on success. Decline closes the window without rolling or spending.'], costs: [], cost_status: isBase ? 'fixture-only' : 'conditional-on-success', source_ids: [] } });
    });
    await page.route('**/api/v1/campaigns/alpha/action?*', route => {
      const id = new URL(route.request().url()).searchParams.get('action_id') ?? '';
      return route.fulfill({ json: saved.has(id) ? { result: saved.get(id) } : null });
    });
    await page.route('**/api/v1/campaigns/alpha/action/confirm', async route => {
      const request = route.request().postDataJSON();
      const isBase = request.intent.kind === 'ability-check';
      if (isBase) {
        expect(request.intent).toEqual({ actor_id: 'hero', kind: 'ability-check', entry_id: 'easy-strength-athletics-dc10' });
        baseCalls++; baseConfirmed = true;
      } else {
        expect(baseConfirmed).toBe(true);
        expect(request.intent).toEqual({ actor_id: 'hero', kind: 'tactical-mind', entry_id: choice });
        tacticalCalls++; consumed = true;
      }
      const receipt = isBase ? baseReceipt : { base: baseReceipt, die: choice === 'use' ? 1 : 0, total: choice === 'use' ? 10 : 9, outcome: choice === 'use' ? 'success' : 'declined', second_wind_before: 2, second_wind_after: choice === 'use' ? 1 : 2, source_ids: [] };
      const result = { action_id: request.action_id, turn_number: isBase ? 1 : 2, narrative: isBase ? `Edda bounded adjudication fixture: DC 10; ${baseReceipt.outcome}. Inspect Tactical Mind only after failure.` : `Tactical Mind ${choice} saved. Original roll retained.`, combat_active: false, state_changes: [], resolution_events: [{ type: isBase ? 'srd_check' : 'srd_tactical_mind', label: isBase ? 'Easy Athletics' : 'Tactical Mind', outcome: receipt.outcome, details: { receipt } }] };
      saved.set(request.action_id, result);
      if (isBase) await route.fulfill({ json: result }); else await route.abort('failed');
    });
    await signIn(page); await page.goto('/play/alpha');
    await page.getByRole('tab', { name: 'Character', exact: true }).click();
    await page.getByLabel('Character action').selectOption('ability-check-fixture:easy-strength-athletics-dc10');
    await page.getByRole('button', { name: 'Preview action', exact: true }).click();
    await expect(page.getByText(/Edda bounded adjudication fixture: Easy Strength/)).toBeVisible();
    expect(baseCalls).toBe(0); expect(tacticalCalls).toBe(0);
    await page.getByRole('button', { name: 'Confirm action', exact: true }).click();
    await expect(page.getByLabel('Action preview').getByText(/Edda bounded adjudication fixture: DC 10/)).toBeVisible();
    expect(baseCalls).toBe(1); expect(tacticalCalls).toBe(0);
    await page.getByLabel('Character action').selectOption(`tactical-mind:${choice === 'decline' ? 'decline' : 'use'}`);
    await page.getByRole('button', { name: 'Preview action', exact: true }).click();
    if (choice === 'base-success') {
      await expect(page.getByText('No eligible saved failed check.')).toBeVisible();
      await expect(page.getByRole('button', { name: 'Confirm action', exact: true })).toHaveCount(0);
    } else {
      await expect(page.getByText(/Apply one d10/)).toBeVisible();
      expect(tacticalCalls).toBe(0);
      await page.getByRole('button', { name: 'Confirm action', exact: true }).click();
      await expect(page.getByRole('alert')).toContainText('Retry the same action');
      await page.reload(); await page.getByRole('tab', { name: 'Character', exact: true }).click();
      await expect(page.getByRole('tabpanel', { name: 'Character', exact: true }).getByLabel('Saved Tactical Mind receipt')).toContainText(choice === 'use' ? 'Second Wind: 2 → 1' : 'Second Wind: 2 → 2');
      expect(tacticalCalls).toBe(1);
    }
    expect(baseCalls).toBe(1); expect(game.actions).toEqual([]);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  });
}
