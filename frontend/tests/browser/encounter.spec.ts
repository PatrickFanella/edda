import { test, expect, signIn } from './fixtures';
import type { Encounter } from '../../src/api/types';
import { campaign, character } from '../fixtures/data';
import { readFileSync } from 'node:fs';
const sheetFixture = JSON.parse(readFileSync(new URL('../fixtures/srd-check-sheet.json', import.meta.url), 'utf8'));
import { expectDocumentFits, expectReachable } from './layout';

for (const [width,height] of [[1366,768],[360,800],[390,844]]) {
 test(`durable encounter summary fits and reloads at ${width}x${height}`, async ({page,game}) => {
  await page.setViewportSize({width,height});
  const encounter = {id:'fixture',status:'active',round:3,turn:0,actors:[{id:'hero',name:'Mira',player:true,hp:11,max_hp:12,initiative:20,initiative_die:19},{id:'foe',name:'Practice foe',player:false,hp:18,max_hp:20,initiative:19,initiative_die:19}],budget:{action:false,bonus_action:true,movement:15}};
  const space = {distance_feet:10,line_of_sight:true,cover:'none',revision:'declared'};
  Object.assign(encounter,{space});
  await page.route('**/campaigns/alpha',route=>route.fulfill({json:{...campaign(),rules_version:'srd-5.2.1',rules_mode:'crunch'}}));
  await page.route('**/campaigns/alpha/character',route=>route.fulfill({json:{...character(),id:'hero',rules_version:'srd-5.2.1',sheet:sheetFixture.sheet}}));
  await page.route('**/campaigns/alpha/character/action-preview?**',route=>route.fulfill({json:{intent:{actor_id:'hero',kind:'combat-end-turn'},rules_version:'srd-5.2.1',expected_revision:'fixture',available:true,targets:[],options:[],costs:[],cost_status:'encounter',source_ids:[],encounter}}));
  await signIn(page);await page.goto('/play/alpha');
  await expect.poll(()=>game.sockets.length).toBe(1);
  const summary=page.getByRole('region',{name:'Encounter turns'});
  await expect(summary).toContainText("Round 3 · Mira's turn");
  await expect(summary).toContainText('Action: spent · Bonus action: available · Movement: 15 ft');
  await expect(summary).toContainText('Distance: 10 ft · Sight: clear · Cover: none');
  await expectDocumentFits(page);
  await expectReachable(page.getByRole('button',{name:'Choose combat action'}));
  await page.reload();
  await expect(summary).toContainText("Round 3 · Mira's turn");
  await expect(page.getByRole('list',{name:'Initiative order'})).toContainText('Practice foe: initiative 19');
  await expectDocumentFits(page);
  await page.getByRole('button',{name:'Choose combat action'}).click();
  const action=page.getByLabel('Character action');
  await action.scrollIntoViewIfNeeded();
  await action.selectOption('combat-end-turn');
  const preview=page.getByRole('button',{name:'Preview action'});
  await preview.scrollIntoViewIfNeeded();await expectReachable(preview);await preview.click();
  const confirm=page.getByRole('button',{name:'Confirm action'});
  await confirm.scrollIntoViewIfNeeded();await expectReachable(confirm);await expectDocumentFits(page);
 });
}


test('starting and completing a fixture enters and exits the durable combat surface',async({page,game})=>{
 await page.setViewportSize({width:390,height:844});
 let encounter: Encounter|undefined;
 let confirmations=0;
 await page.route('**/campaigns/alpha',route=>route.fulfill({json:{...campaign(),rules_version:'srd-5.2.1',rules_mode:'crunch'}}));
 await page.route('**/campaigns/alpha/character',route=>route.fulfill({json:{...character(),id:'hero',rules_version:'srd-5.2.1',sheet:sheetFixture.sheet}}));
 await page.route('**/campaigns/alpha/character/inventory',route=>route.fulfill({json:[{id:'sword',campaign_id:'alpha',name:'Longsword',description:'Fixture sword',rarity:'common',item_type:'weapon',properties:{srd521_id:'longsword'},equipped:true,quantity:1}]}));
 await page.route('**/campaigns/alpha/character/action-preview?**',route=>{
  const kind=new URL(route.request().url()).searchParams.get('kind');
  return route.fulfill({json:{intent:{actor_id:'hero',kind, ...(kind==='combat-start'?{entry_id:'sword'}:{})},rules_version:'srd-5.2.1',expected_revision:'fixture',available:!!encounter||kind==='combat-start',targets:kind==='combat-start'?[{id:'foe',label:'Practice foe'}]:[],options:[],costs:[],cost_status:'encounter',source_ids:[],encounter}});
 });
 await page.route('**/campaigns/alpha/action?**',route=>route.fulfill({json:null}));
 await page.route('**/campaigns/alpha/action/confirm',route=>{
  confirmations++;
  encounter={id:'fixture',status:confirmations===1?'active':'completed',round:1,turn:0,actors:[{id:'hero',name:'Mira',player:true,hp:12,max_hp:12,initiative:20,initiative_die:19},{id:'foe',name:'Practice foe',player:false,hp:confirmations===1?8:0,max_hp:8,initiative:19,initiative_die:19}],budget:{action:true,bonus_action:true,movement:30}};
  return route.fulfill({json:{action_id:route.request().postDataJSON().action_id,turn_number:confirmations+2,narrative:confirmations===1?'Initiative recorded.':'Encounter complete.',combat_active:false,state_changes:[],resolution_events:[]}});
 });
 await signIn(page);await page.goto('/play/alpha');
 await expect(page.getByRole('region',{name:'Narrative combat',exact:true})).toHaveCount(0);
 await page.getByRole('tab',{name:'Character',exact:true}).click();
 await page.getByLabel('Character action').selectOption('combat-start:sword');
 await page.getByRole('button',{name:'Preview action',exact:true}).click();
 await page.getByLabel('Attack target').selectOption('foe');
 await page.getByRole('button',{name:'Confirm action',exact:true}).click();
 await expect(page.getByText('Action saved.')).toBeVisible();
 await page.getByRole('tab',{name:'Play',exact:true}).click();
 await expectReachable(page.getByRole('button',{name:'Choose combat action'}));
 await page.getByRole('button',{name:'Choose combat action'}).click();
 await page.getByLabel('Character action').selectOption('combat-end-turn');
 await page.getByRole('button',{name:'Preview action',exact:true}).click();
 await page.getByRole('button',{name:'Confirm action',exact:true}).click();
 await expect(page.getByRole('region',{name:'Narrative combat',exact:true})).toHaveCount(0);
 await expect(page.getByRole('region',{name:'Encounter turns'})).toContainText('Encounter complete');
 await page.reload();
 await expect(page.getByRole('region',{name:'Encounter turns'})).toContainText('Encounter complete');
 await expect(page.getByRole('region',{name:'Narrative combat',exact:true})).toHaveCount(0);
 expect(game.actions).toEqual([]);
});
