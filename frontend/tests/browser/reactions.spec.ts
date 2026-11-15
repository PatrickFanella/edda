import { test, expect, signIn } from './fixtures';
import { campaign, character } from '../fixtures/data';
import { readFileSync } from 'node:fs';
import type { Encounter } from '../../src/api/types';
import { expectDocumentFits, expectReachable } from './layout';

const sheetFixture = JSON.parse(readFileSync(new URL('../fixtures/srd-check-sheet.json', import.meta.url), 'utf8'));

for (const [width,height] of [[1366,768],[360,800]]) {
 for (const response of ['attack','decline']) {
  test(`saved reaction ${response} offers consent and reloads at ${width}x${height}`,async({page,game})=>{
   await page.setViewportSize({width,height});
   const encounter: Encounter = {id:'encounter',status:'active',round:1,turn:1,actors:[{id:'hero',name:'Mira',player:true,hp:11,max_hp:12,initiative:20,initiative_die:19},{id:'foe',name:'Practice foe',player:false,hp:20,max_hp:20,initiative:19,initiative_die:19}],budget:{action:false,bonus_action:true,movement:30},space:{distance_feet:5,line_of_sight:true,cover:'none',revision:'scene'},reaction:{available:true,pending:{id:'trigger',target_id:'foe',weapon_id:'sword',space_revision:'scene'}}};
   const confirmations: {intent:{kind:string};target_id?:string;action_id:string}[]=[];
   await page.route('**/campaigns/alpha',route=>route.fulfill({json:{...campaign(),rules_version:'srd-5.2.1',rules_mode:'crunch'}}));
   await page.route('**/campaigns/alpha/character',route=>route.fulfill({json:{...character(),id:'hero',rules_version:'srd-5.2.1',sheet:sheetFixture.sheet}}));
   await page.route('**/campaigns/alpha/character/action-preview?**',route=>{
    const kind=new URL(route.request().url()).searchParams.get('kind');
    return route.fulfill({json:{intent:{actor_id:'hero',kind,...(kind==='combat-react'?{entry_id:'sword'}:{})},rules_version:'srd-5.2.1',expected_revision:'trigger-revision',available:kind==='combat-react'||kind==='combat-decline-reaction',encounter,targets:kind==='combat-react'?[{id:'foe',label:'Practice foe',distance_feet:5,reach_feet:5}]:[],options:['Opportunity attack before leaving reach.'],costs:kind==='combat-react'?[{resource_id:'reaction',amount:1}]:[],cost_status:'encounter',source_ids:[]}});
   });
   await page.route('**/campaigns/alpha/action?**',route=>route.fulfill({json:null}));
   await page.route('**/campaigns/alpha/action/confirm',route=>{
    confirmations.push(route.request().postDataJSON());
    encounter.reaction={available:true};encounter.turn=0;encounter.round=2;encounter.space!.distance_feet=10;
    return route.fulfill({json:{action_id:confirmations[0].action_id,turn_number:4,narrative:response==='attack'?'One reaction spent; recovered at the start of your next turn.':'Opportunity attack declined. No reaction spent.',state_changes:[],resolution_events:[],combat_active:false}});
   });
   await signIn(page);await page.goto('/play/alpha');
   const choice=page.getByRole('region',{name:'Reaction choice'});
   await expect(choice).toContainText('Practice foe is about to move out of your 5-foot reach');
   await expectReachable(page.getByRole('button',{name:'Preview opportunity attack'}));
   await expectReachable(page.getByRole('button',{name:'Preview decline',exact:true}));
   await expectDocumentFits(page);
   expect(confirmations).toEqual([]);
   await page.reload();
   await expect(choice).toBeVisible();
   await page.getByRole('button',{name:response==='attack'?'Preview opportunity attack':'Preview decline',exact:true}).click();
   const preview=page.getByRole('button',{name:'Preview action',exact:true});
   await preview.scrollIntoViewIfNeeded();await preview.click();
   await expect(page.getByText('Preview ready. No action has been performed.')).toBeVisible();
   expect(confirmations).toEqual([]);
   if(response==='attack')await expect(page.getByText(/reaction: 1/)).toBeVisible();
   const confirm=page.getByRole('button',{name:'Confirm action',exact:true});
   await confirm.scrollIntoViewIfNeeded();await expectReachable(confirm);await expectDocumentFits(page);
   await confirm.click();
   await expect(choice).toHaveCount(0);
   await expect(page.getByRole('region',{name:'Encounter turns'})).toContainText('Distance: 10 ft');
   expect(confirmations).toHaveLength(1);
   expect(confirmations[0].intent.kind).toBe(response==='attack'?'combat-react':'combat-decline-reaction');
   if(response==='attack')expect(confirmations[0].target_id).toBe('foe');
   expect(game.actions).toEqual([]);
  });
 }
}
