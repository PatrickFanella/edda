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

for (const width of [1440, 390]) {
 test(`selects and confirms the martial fixture at ${width}px`, async ({ page, game }) => {
  await page.setViewportSize({width,height:844});
  await page.route('**/api/v1/campaigns/alpha/character', route => route.fulfill({json:{...character(),rules_version:'srd-5.2.1',sheet:fixture.sheet}}));
  await page.route('**/api/v1/campaigns/alpha/character/inventory', route => route.fulfill({json:[{id:'sword',campaign_id:'alpha',description:'Fixture sword',rarity:'common',name:'Longsword',item_type:'weapon',properties:{srd521_id:'longsword'},equipped:true,quantity:1}]}));
  await page.route('**/api/v1/campaigns/alpha/character/action-preview?*', route => route.fulfill({json:{intent:{actor_id:'hero',kind:'martial-attack',entry_id:'sword'},rules_version:'srd-5.2.1',expected_revision:'martial-v1:fixture',available:true,modifier:fixture.sheet.skills.athletics,targets:[{id:'foe',label:'Practice foe'}],options:['ordinary, one hand'],costs:[{resource_id:'attack-action',amount:1}],cost_status:'fixture-only',source_ids:[]}}));
  let confirmations=0;
  await page.route('**/api/v1/campaigns/alpha/action?*', route => route.fulfill({json:null}));
  await page.route('**/api/v1/campaigns/alpha/action/confirm',route=>{confirmations++;expect(route.request().postDataJSON().target_id).toBe('foe');return route.fulfill({json:{action_id:'saved',turn_number:3,narrative:'Longsword hits for 8 slashing.',combat_active:false,state_changes:[],resolution_events:[]}})});
  await signIn(page);await page.goto('/play/alpha');await page.getByRole('tab',{name:'Character',exact:true}).click();
  await page.getByLabel('Character action').selectOption('martial-attack:sword');await page.getByRole('button',{name:'Preview action',exact:true}).click();
  await expect(page.getByRole('button',{name:'Confirm action',exact:true})).toBeDisabled();
  await page.getByLabel('Attack target').selectOption('foe');expect(confirmations).toBe(0);expect(game.actions).toEqual([]);
  await page.getByRole('button',{name:'Confirm action',exact:true}).click();await expect(page.getByText('Action saved.')).toBeVisible();expect(confirmations).toBe(1);
  await page.getByRole('tab',{name:'Play',exact:true}).click();await expect(page.getByText('Longsword hits for 8 slashing.',{exact:true})).toBeVisible();
  expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
 });
}

for (const width of [1440,390]) {
 test(`uses one potion and refreshes inventory at ${width}px`,async({page,game})=>{
  await page.setViewportSize({width,height:844});
  let quantity=2,hp=1,confirmations=0;
  await page.route('**/api/v1/campaigns/alpha/character',route=>route.fulfill({json:{...character(),hp,rules_version:'srd-5.2.1',sheet:{...fixture.sheet,hp:{...fixture.sheet.hp,value:hp}}}}));
  await page.route('**/api/v1/campaigns/alpha/character/inventory',route=>route.fulfill({json:[{id:'potion',campaign_id:'alpha',name:'Potion of Healing',description:'Fixture potion',rarity:'common',item_type:'consumable',properties:{srd521_id:'potion-of-healing'},equipped:false,quantity}]}));
  await page.route('**/api/v1/campaigns/alpha/character/action-preview?*',route=>route.fulfill({json:{intent:{actor_id:'hero',kind:'consumable-use',entry_id:'potion'},rules_version:'srd-5.2.1',expected_revision:'consumable-v1:fixture',available:true,targets:[{id:'hero',label:'Mira (self)'}],options:['Quantity: 2; regain 2d4 + 2 HP, capped at 12 (current 1); self-use only'],costs:[{resource_id:'item:potion',amount:1},{resource_id:'bonus-action',amount:1}],cost_status:'fixture-only',source_ids:[]}}));
  await page.route('**/api/v1/campaigns/alpha/action?*',route=>route.fulfill({json:null}));
  await page.route('**/api/v1/campaigns/alpha/action/confirm',route=>{confirmations++;expect(route.request().postDataJSON().target_id).toBe('hero');quantity=1;hp=7;return route.fulfill({json:{action_id:'saved',turn_number:3,narrative:'Mira drinks a Potion of Healing: regains 6 HP.',combat_active:false,state_changes:[],resolution_events:[]}})});
  await signIn(page);await page.goto('/play/alpha');await page.getByRole('tab',{name:'Character',exact:true}).click();
  await page.getByLabel('Character action').selectOption('consumable-use:potion');await page.getByRole('button',{name:'Preview action',exact:true}).click();
  await expect(page.getByText(/Roll options: Quantity: 2/)).toBeVisible();expect(confirmations).toBe(0);expect(game.actions).toEqual([]);
  await page.getByRole('button',{name:'Confirm action',exact:true}).click();await expect(page.getByText('Action saved.')).toBeVisible();expect(confirmations).toBe(1);
  await expect(page.getByRole('option',{name:'Potion of Healing (quantity 1, self)'})).toBeAttached();
  await page.getByRole('tab',{name:'Play',exact:true}).click();await expect(page.getByText('Mira drinks a Potion of Healing: regains 6 HP.',{exact:true})).toBeVisible();
  expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
 });
}
