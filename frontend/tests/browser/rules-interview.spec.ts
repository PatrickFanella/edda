import { test, expect, signIn } from './fixtures';
import { heroProfile, campaign } from '../fixtures/data';

for (const width of [390, 1440]) {
  test(`SRD interview corrects unsupported choices and resumes after provider failure at ${width}px`, async ({ page, game }) => {
    await page.setViewportSize({ width, height: 1000 });
    let starts = 0, steps = 0, builds = 0;
    const selected = { ...heroProfile, name: 'Mira', concept: 'Human Cleric', background: 'Acolyte', rules_choices: { origin: 'Human', class: 'Cleric', background: 'Acolyte' } };
    await page.route('**/campaigns/start/character-interview', route => {
      expect(route.request().postDataJSON().rules_version).toBe('srd-5.2.1'); starts++;
      return route.fulfill({ json: { session_id: 'rules-interview', message: 'Choose Human Soldier Fighter or Human Acolyte Cleric.', done: false } });
    });
    await page.route('**/campaigns/start/character-interview/rules-interview', route => {
      steps++;
      if (steps === 1) return route.fulfill({ json: { session_id: 'rules-interview', message: 'Rules choices need correction: Wizard is unsupported; no character has been created.', done: false } });
      if (steps === 2) return route.fulfill({ status: 500, json: { error: 'Fixture provider unavailable' } });
      return route.fulfill({ json: { session_id: 'rules-interview', message: 'Human Acolyte Cleric choices validated. Review before generation.', done: true, profile: selected } });
    });
    await page.route('**/campaigns/start/world', route => {
      builds++; expect(route.request().postDataJSON()).toMatchObject({ rules_version: 'srd-5.2.1', character_profile: selected });
      return route.fulfill({ json: { campaign: { ...campaign(), rules_version: 'srd-5.2.1' }, opening_scene: { narrative: 'A lantern lights the road.', choices: [] } } });
    });
    await page.route('**/api/v1/campaigns/alpha/character/action-preview?*', route => route.fulfill({ json: { intent: { actor_id: 'hero', kind: 'combat-end-turn' }, available: false, targets: [], options: [], costs: [], source_ids: [] } }));
    await signIn(page); await page.goto('/new');
    await page.getByRole('button', { name: /Campaign interview/ }).click(); await page.getByRole('button', { name: 'Continue', exact: true }).click();
    await page.getByLabel('Your campaign notes').fill('A hopeful frontier'); await page.getByRole('button', { name: 'Send reply' }).click();
    await page.getByRole('button', { name: /Supported SRD slice/ }).click(); await page.getByRole('button', { name: 'Continue', exact: true }).click();
    await page.getByRole('button', { name: /Character interview/ }).click(); await page.getByRole('button', { name: 'Continue', exact: true }).click();
    await page.getByLabel('Your character notes').fill('An invented Wizard'); await page.getByRole('button', { name: 'Send reply' }).click();
    await expect(page.getByText(/Rules choices need correction: Wizard/)).toBeVisible(); expect(builds).toBe(0);
    await page.getByLabel('Your character notes').fill('Use Human Acolyte Cleric'); await page.getByRole('button', { name: 'Send reply' }).click();
    await expect(page.getByText('Fixture provider unavailable')).toBeVisible();
    await page.getByRole('button', { name: 'Back', exact: true }).click(); await page.getByRole('button', { name: 'Continue', exact: true }).click();
    await expect(page.getByText(/Rules choices need correction: Wizard/)).toBeVisible(); expect(starts).toBe(1);
    await page.getByLabel('Your character notes').fill('Use Human Acolyte Cleric'); await page.getByRole('button', { name: 'Send reply' }).click();
    await expect(page.getByText(/AC 11 · 10 HP · Wisdom casting/)).toBeVisible(); expect(builds).toBe(0);
    await page.getByRole('button', { name: 'Begin adventure' }).click(); await expect(page).toHaveURL(/play\/alpha$/);
    expect(builds).toBe(1); expect(game.actions).toEqual([]);
  });
}
