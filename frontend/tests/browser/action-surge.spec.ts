import { test, expect, signIn } from './fixtures';
import { character } from '../fixtures/data';
import fixture from '../fixtures/srd-check-sheet.json' with { type: 'json' };
import type { TurnResponse } from '../../src/api/types';

for (const width of [1440, 390]) {
 test(`Action Surge preview confirmation lost response reload reconnect ${width}px`,async({page,game})=>{
  await page.setViewportSize({width,height:844});
  let commits=0; let receipt:TurnResponse|null=null;
  await page.route('**/campaigns/alpha/character',r=>r.fulfill({json:{...character(),level:2,rules_version:'srd-5.2.1',sheet:{...fixture.sheet,scope:'fighter-2-chain-mail-defense-fixture',resources:[{id:'action-surge',label:'Action Surge',current:commits?0:1,maximum:1,source_id:'fighter/action-surge'}],capabilities:[{id:'action-surge',label:'Action Surge',playable:true,source_id:'fighter/action-surge'}]}}}));
  await page.route('**/character/action-preview?*',r=>r.fulfill({json:{intent:{actor_id:'hero',kind:'action-surge'},rules_version:'srd-5.2.1',expected_revision:'surge-fixture',available:commits===0,reason:'Action Surge unavailable.',targets:[],options:['One additional Action this turn except Magic. Bonus Action, Reaction and movement unchanged.'],costs:[{resource_id:'action-surge',amount:1}],cost_status:'encounter',source_ids:['srd-5.2.1:content/fighter/action-surge']}}));
  await page.route('**/campaigns/alpha/action?*',r=>r.fulfill({json:receipt?{result:receipt}:null}));
  await page.route('**/campaigns/alpha/action/confirm',r=>{
   const request=r.request().postDataJSON();expect(request.intent).toEqual({actor_id:'hero',kind:'action-surge'});
   commits++;receipt={action_id:request.action_id,turn_number:3,narrative:'Action Surge saved.',combat_active:false,state_changes:[],resolution_events:[{type:'srd_action_surge',label:'Action Surge',outcome:'saved',details:{receipt:{rules_version:'srd-5.2.1',kind:'action-surge',uses_before:1,uses_after:0,additional_actions:1,excluded_action:'Magic',source_ids:['srd-5.2.1:content/fighter/action-surge']}}}]};
   return r.abort('failed');
  });
  await signIn(page);await page.goto('/play/alpha');
  await page.getByRole('tab',{name:'Character',exact:true}).click();
  await page.getByLabel('Character action').selectOption('action-surge');
  await page.getByRole('button',{name:'Preview action',exact:true}).click();
  await expect(page.getByText('Preview ready. No action has been performed.')).toBeVisible();expect(commits).toBe(0);
  await page.getByRole('button',{name:'Confirm action',exact:true}).click();
  await expect.poll(()=>commits).toBe(1);
  await page.reload();await page.getByRole('tab',{name:'Character',exact:true}).click();
  await expect(page.getByRole('tabpanel',{name:'Character',exact:true}).getByLabel('Action Surge receipt')).toContainText('excludes Magic');
  await expect(page.getByText('Action Surge: 0 / 1')).toBeVisible();expect(commits).toBe(1);
  await page.reload();await expect.poll(()=>game.sockets.length).toBeGreaterThan(1);
  await page.getByRole('tab',{name:'Character',exact:true}).click();
  await page.getByLabel('Character action').selectOption('action-surge');await page.getByRole('button',{name:'Preview action',exact:true}).click();
  await expect(page.getByText('Action Surge unavailable.',{exact:true})).toBeVisible();expect(commits).toBe(1);expect(game.actions).toEqual([]);
  expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
 });
}
