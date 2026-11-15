import { test, expect, signIn } from './fixtures';
import { campaign } from '../fixtures/data';

for (const width of [390, 1440]) {
for (const build of [{class:'Fighter',background:'Soldier',summary:/Savage Attacker and Skilled · AC 19 · 12 HP/},{class:'Cleric',background:'Acolyte',summary:/AC 11 · 10 HP · Wisdom casting/}]) {
test(`guided ${build.class} creation at ${width} shows derived choices and retains them after failed generation`, async ({page,game})=>{
 await page.setViewportSize({width,height:1000});
 let attempts=0;
 const requests:unknown[]=[];
 await page.route('**/campaigns/start/world',route=>{
  attempts++;requests.push(route.request().postDataJSON());
  if(attempts===1) return route.fulfill({status:500,json:{error:'Fixture generation failed'}});
  return route.fulfill({json:{campaign:{...campaign(),rules_version:'srd-5.2.1',rules_mode:'crunch'},opening_scene:{narrative:'A lantern lights the road.',choices:[]}}});
 });
 await signIn(page);await page.goto('/new');
 await page.getByRole('button',{name:/Campaign interview/}).click();
 await page.getByRole('button',{name:'Continue',exact:true}).click();
 await page.getByLabel('Your campaign notes').fill('A hopeful frontier');
 await page.getByRole('button',{name:'Send reply'}).click();
 await page.getByRole('button',{name:/Supported SRD slice/}).click();
 await page.getByRole('button',{name:'Continue',exact:true}).click();
 await expect(page.getByRole('button',{name:/Character interview/})).toHaveCount(0);
 await page.getByRole('button',{name:/Guided character build/}).click();
 await page.getByRole('button',{name:'Continue',exact:true}).click();
 await page.getByLabel('Character name').fill('Mira');
 await page.getByLabel('Race',{exact:true}).selectOption('Human');
 await page.getByLabel('Class',{exact:true}).selectOption(build.class);
 await page.getByLabel('Background',{exact:true}).selectOption(build.background);
 await page.getByLabel('Alignment',{exact:true}).selectOption('Lawful Good');
 await page.getByRole('button',{name:'Continue to confirmation'}).click();
 await expect(page.getByText(build.summary)).toBeVisible();
 await page.getByRole('button',{name:'Begin adventure'}).click();
 await expect(page.getByText(/Fixture generation failed/)).toBeVisible();
 // Back returns to the controlled form without losing any guided choices.
 await page.getByRole('button',{name:'Back',exact:true}).click();
 await expect(page.getByLabel('Character name')).toHaveValue('Mira');
 await expect(page.getByLabel('Race',{exact:true})).toHaveValue('Human');
 await expect(page.getByLabel('Class',{exact:true})).toHaveValue(build.class);
 await expect(page.getByLabel('Background',{exact:true})).toHaveValue(build.background);
 await expect(page.getByLabel('Alignment',{exact:true})).toHaveValue('Lawful Good');
 await page.getByRole('button',{name:'Continue to confirmation'}).click();
 await page.getByRole('button',{name:'Begin adventure'}).click();
 await expect(page).toHaveURL(/play\/alpha$/);
 expect(requests).toHaveLength(2);
 expect(requests[1]).toEqual(requests[0]);
 expect(requests[0]).toMatchObject({rules_version:'srd-5.2.1',character_profile:{name:'Mira',concept:`human ${build.class.toLowerCase()}`,background:build.background}});
 expect(game.actions).toEqual([]);
});

}

}
