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

for (const width of [1440,390]) {
 test(`casts Cure Wounds with explicit slot and target at ${width}px`,async({page,game})=>{
  await page.setViewportSize({width,height:844});
  let slots=2,hp=1,confirmations=0;
  const spellcasting={ability:'wisdom',modifier:{value:3,derivation:[]},spells:[{id:'cure-wounds',label:'Cure Wounds',known:true,prepared:true,components:['verbal','somatic'],components_available:true,slot_levels:[1],source_id:'spell'}]};
  await page.route('**/api/v1/campaigns/alpha/character',route=>route.fulfill({json:{...character(),hp,max_hp:10,rules_version:'srd-5.2.1',sheet:{...fixture.sheet,scope:'cleric-1-cure-wounds-self-fixture',spellcasting,hp:{value:hp,derivation:[]},max_hp:{value:10,derivation:[]},resources:[{id:'spell-slot-1',label:'Level 1 spell slots',current:slots,maximum:2,source_id:'spell'}],capabilities:[]}}}));
  await page.route('**/api/v1/campaigns/alpha/character/action-preview?*',route=>{
   expect(new URL(route.request().url()).searchParams.get('slot_level')).toBe('1');
   return route.fulfill({json:{intent:{actor_id:'hero',kind:'spell-cast',entry_id:'cure-wounds',slot_level:1},rules_version:'srd-5.2.1',expected_revision:'casting-fixture',available:true,targets:[{id:'hero',label:'Mira (self; touch)'}],options:['2d8 + Wisdom; verbal and somatic; touch (self only)'],costs:[{resource_id:'spell-slot-1',amount:1},{resource_id:'magic-action',amount:1}],cost_status:'fixture-only',source_ids:[]}});
  });
  await page.route('**/api/v1/campaigns/alpha/action?*',route=>route.fulfill({json:null}));
  await page.route('**/api/v1/campaigns/alpha/action/confirm',route=>{confirmations++;expect(route.request().postDataJSON()).toMatchObject({target_id:'hero',intent:{slot_level:1,entry_id:'cure-wounds'}});slots=1;hp=8;return route.fulfill({json:{action_id:'saved',turn_number:3,narrative:'Mira casts Cure Wounds on self: regains 7 HP.',combat_active:false,state_changes:[],resolution_events:[{type:'srd_spell',label:'Cure Wounds',outcome:'healed',details:{receipt:{rules_version:'srd-5.2.1',kind:'spell-cast',spell:'Cure Wounds',target:'Mira',target_id:'hero',slot_level:1,slots_before:2,slots_after:1,dice:[2,2],modifier:{value:3,derivation:[]},rolled_healing:7,healing:7,hp_before:1,hp_after:8,source_ids:[]}}}]}})});
  await signIn(page);await page.goto('/play/alpha');await page.getByRole('tab',{name:'Character',exact:true}).click();
  await expect(page.getByLabel('Spellcasting', {exact:true})).toContainText('Known · Prepared');
  await page.getByLabel('Character action').selectOption('spell-cast:cure-wounds');
  await expect(page.getByLabel('Spell slot level')).toHaveValue('1');
  await page.getByRole('button',{name:'Preview action',exact:true}).click();
  await expect(page.getByRole('button',{name:'Confirm action',exact:true})).toBeDisabled();
  await page.getByLabel('Spell target').selectOption('hero');expect(confirmations).toBe(0);expect(game.actions).toEqual([]);
  await page.getByRole('button',{name:'Confirm action',exact:true}).click();await expect(page.getByText('Action saved.')).toBeVisible();expect(confirmations).toBe(1);
  await expect(page.getByText('Level 1 spell slots: 1 / 2')).toBeVisible();
  await page.getByRole('tab',{name:'Play',exact:true}).click();await expect(page.getByText('Mira casts Cure Wounds on self: regains 7 HP.',{exact:true})).toBeVisible();
  await page.getByRole('article',{name:'Game Master entry'}).last().getByText('Turn details').click();
  await page.getByRole('tabpanel',{name:'Play',exact:true}).getByText('Cure Wounds: Mira regains 7 HP').click();await expect(page.getByText('Level 1 slots 2 → 1')).toBeVisible();
  expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
 });
}

for (const width of [1440,390]) {
 for (const recovery of ['confirm','retry','reload']) {
 test(`casts Cure Wounds on ally with ${recovery} at ${width}px`,async({page,game})=>{
  await page.setViewportSize({width,height:844});
  let slots=2,confirmations=0, reconcile=false;
  const hp=1;
  let requestId='';
  const spellcasting={ability:'wisdom',modifier:{value:3,derivation:[]},spells:[{id:'cure-wounds',label:'Cure Wounds',known:true,prepared:true,components:['verbal','somatic'],components_available:true,slot_levels:[1],source_id:'spell'}]};
  await page.route('**/api/v1/campaigns/alpha/character',route=>route.fulfill({json:{...character(),hp,max_hp:10,rules_version:'srd-5.2.1',sheet:{...fixture.sheet,scope:'cleric-1-cure-wounds-self-fixture',spellcasting,hp:{value:hp,derivation:[]},max_hp:{value:10,derivation:[]},resources:[{id:'spell-slot-1',label:'Level 1 spell slots',current:slots,maximum:2,source_id:'spell'}],capabilities:[]}}}));
  await page.route('**/api/v1/campaigns/alpha/character/action-preview?*',route=>{
   expect(new URL(route.request().url()).searchParams.get('slot_level')).toBe('1');
   return route.fulfill({json:{intent:{actor_id:'hero',kind:'spell-cast',entry_id:'cure-wounds',slot_level:1},rules_version:'srd-5.2.1',expected_revision:'casting-fixture',available:true,targets:[{id:'hero',label:'Mira (self; touch)'},{id:'ally',label:'Rowan (willing conscious ally; touch)',distance_feet:5,reach_feet:5}],options:['2d8 + Wisdom; verbal and somatic; touch; willing conscious ally'],costs:[{resource_id:'spell-slot-1',amount:1},{resource_id:'magic-action',amount:1}],cost_status:'fixture-only',source_ids:[]}});
  });
  const saved={action_id:'saved',turn_number:3,narrative:'Mira casts Cure Wounds on Rowan: regains 7 HP.',combat_active:false,state_changes:[],resolution_events:[{type:'srd_spell',label:'Cure Wounds',outcome:'healed',details:{receipt:{rules_version:'srd-5.2.1',kind:'spell-cast',spell:'Cure Wounds',target:'Rowan',target_id:'ally',slot_level:1,slots_before:2,slots_after:1,dice:[2,2],modifier:{value:3,derivation:[]},rolled_healing:7,healing:7,hp_before:1,hp_after:8,source_ids:[]}}}]};
  await page.route('**/api/v1/campaigns/alpha/action?*',route=>{if(reconcile) expect(new URL(route.request().url()).searchParams.get('action_id')).toBe(requestId);return route.fulfill({json:reconcile?{result:saved}:null})});
  await page.route('**/api/v1/campaigns/alpha/action/confirm',route=>{
   confirmations++;
   const body=route.request().postDataJSON();
   expect(body).toMatchObject({target_id:'ally',intent:{slot_level:1,entry_id:'cure-wounds'}});
   if(requestId) expect(body.action_id).toBe(requestId);
   requestId=body.action_id;saved.action_id=requestId;slots=1;
   if(recovery!=='confirm' && confirmations===1) return route.abort('failed');
   return route.fulfill({json:saved});
  });
  await signIn(page);await page.goto('/play/alpha');await page.getByRole('tab',{name:'Character',exact:true}).click();
  await expect(page.getByLabel('Spellcasting', {exact:true})).toContainText('Known · Prepared');
  await page.getByLabel('Character action').selectOption('spell-cast:cure-wounds');
  await expect(page.getByLabel('Spell slot level')).toHaveValue('1');
  await page.getByRole('button',{name:'Preview action',exact:true}).click();
  await expect(page.getByRole('button',{name:'Confirm action',exact:true})).toBeDisabled();
  await page.getByLabel('Spell target').selectOption('ally');expect(confirmations).toBe(0);expect(game.actions).toEqual([]);
  await page.getByRole('button',{name:'Confirm action',exact:true}).click();
  if(recovery!=='confirm') {
   await expect(page.getByRole('button',{name:'Retry same action',exact:true})).toBeVisible();
   if(recovery==='reload') {reconcile=true;await page.reload();await page.getByRole('tab',{name:'Character',exact:true}).click();}
   else await page.getByRole('button',{name:'Retry same action',exact:true}).click();
  }
  await expect(page.getByText('Action saved.')).toBeVisible();expect(confirmations).toBe(recovery==='retry'?2:1);
  await expect(page.getByText('Level 1 spell slots: 1 / 2')).toBeVisible();
  await page.getByRole('tab',{name:'Play',exact:true}).click();await expect(page.getByText('Mira casts Cure Wounds on Rowan: regains 7 HP.',{exact:true})).toBeVisible();
  await page.getByRole('article',{name:'Game Master entry'}).last().getByText('Turn details').click();
  await page.getByRole('tabpanel',{name:'Play',exact:true}).getByText('Cure Wounds: Rowan regains 7 HP').click();await expect(page.getByText('Level 1 slots 2 → 1')).toBeVisible();
  expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
 });
 }
}
