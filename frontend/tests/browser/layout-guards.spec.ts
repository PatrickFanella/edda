import { test, expect } from '@playwright/test';
import { controlGeometry } from './layout';

test('layout probe rejects a clipped or covered control even in a bounded document', async ({ page }) => {
  await page.setContent(`<style>body{margin:0}section{height:60px;overflow:hidden}button{display:block;width:100px;height:44px;margin-top:80px}aside{position:absolute;inset:0;background:black}</style><section><button>Send</button></section>`);
  expect(await page.evaluate(() => document.documentElement.scrollHeight)).toBe(await page.evaluate(() => innerHeight));
  const button = page.getByRole('button', { name: 'Send' });
  expect((await controlGeometry(button)).visibleHeight).toBe(0);
  await button.evaluate(el => { (el as HTMLElement).style.marginTop = '0'; });
  expect((await controlGeometry(button)).hittable).toBe(true);
  await page.evaluate(() => document.body.appendChild(document.createElement('aside')));
  const covered = await controlGeometry(button);
  expect(covered.visibleHeight).toBe(44);
  expect(covered.hittable).toBe(false);
});
