import { test, expect, signIn } from './fixtures';
import { history, timestamp } from '../fixtures/data';
import type { Page } from '@playwright/test';

async function expectDocumentFits(page: Page) {
  const size = await page.evaluate(() => ({ width: innerWidth, height: innerHeight, docWidth: document.documentElement.scrollWidth, docHeight: document.documentElement.scrollHeight }));
  expect(size.docWidth).toBeLessThanOrEqual(size.width);
  expect(size.docHeight).toBeLessThanOrEqual(size.height);
}

async function longHistory(page: Page) {
  await page.route('**/campaigns/alpha/history', route => route.fulfill({ json: { entries: history().map(entry => ({ ...entry, llm_response: `${entry.llm_response} ${'The road continues past the old gate. '.repeat(200)}` })) } }));
}

for (const [width, height] of [[1366, 768], [1440, 900], [1920, 1080], [768, 1024], [360, 800], [390, 844]]) {
  test(`gameplay owns its scrolling at ${width}x${height}`, async ({ page, game }, testInfo) => {
    await page.setViewportSize({ width, height });
    await longHistory(page);
    await signIn(page); await page.goto('/play/alpha');
    await expect.poll(() => game.sockets.length).toBe(1);
    const log = page.getByRole('log');
    await expect(log.getByText(/The alpha gate opens/)).toBeVisible();
    await expectDocumentFits(page);
    const input = page.getByLabel('Your move');
    await input.fill('A multiline draft\nthat remains reachable.');
    const send = page.getByRole('button', { name: 'Send action' });
    await expect(send).toBeInViewport();
    expect(await send.evaluate(el => el.getBoundingClientRect().height)).toBeGreaterThanOrEqual(44);
    expect(await log.evaluate(el => el.scrollHeight)).toBeGreaterThan(await log.evaluate(el => el.clientHeight));
    await log.evaluate(el => { el.scrollTop = el.scrollHeight; });
    expect(await page.evaluate(() => scrollY)).toBe(0);
    await page.screenshot({ path: testInfo.outputPath('viewport.png') });
    await page.getByRole('tab', { name: 'Logs', exact: true }).click();
    const details = page.getByRole('tabpanel', { name: 'Logs' }).locator(':scope > div');
    await expect(details).toBeInViewport();
    expect(await details.evaluate(el => el.scrollHeight)).toBeGreaterThan(await details.evaluate(el => el.clientHeight));
    expect(await details.evaluate(el => el.scrollWidth)).toBeLessThanOrEqual(await details.evaluate(el => el.clientWidth));
    await expectDocumentFits(page);
  });
}

for (const [width, height] of [[360, 800], [390, 844]]) {
  test(`long choices, details and reduced viewport keep controls reachable at ${width}`, async ({ page, game }) => {
    await page.setViewportSize({ width, height });
    await longHistory(page);
    await signIn(page); await page.goto('/play/alpha');
    await expect.poll(() => game.sockets.length).toBe(1);
    const choices = Array.from({ length: 8 }, (_, i) => ({ id: `long-${i}`, text: `Option ${i + 1}: ${'Consider the old gate carefully. '.repeat(20)}` }));
    game.sockets[0].send(JSON.stringify({ type: 'result', timestamp, payload: { narrative: 'The keeper waits.', state_changes: [], choices } }));
    const region = page.getByRole('region', { name: 'Suggested choices' });
    await expect(region.getByRole('button')).toHaveCount(8);
    await expectDocumentFits(page);
    expect(await page.getByRole('log').evaluate(el => el.clientHeight)).toBeGreaterThanOrEqual(80);
    await expect(page.getByRole('button', { name: 'Send action' })).toBeInViewport();
    await region.getByRole('button').last().scrollIntoViewIfNeeded();
    await expect(region.getByRole('button').last()).toBeInViewport();
    // Deterministic resize, not evidence of an actual software keyboard.
    await page.setViewportSize({ width, height: 480 });
    await page.getByLabel('Your move').fill('Keyboard-sized viewport\nMultiline draft');
    await page.getByRole('button', { name: 'Send action' }).scrollIntoViewIfNeeded();
    await expect(page.getByRole('button', { name: 'Send action' })).toBeInViewport();
    await expectDocumentFits(page);
    await page.setViewportSize({ width, height });
    await page.evaluate(() => { document.documentElement.style.fontSize = '200%'; });
    for (const name of ['Logs', 'Character', 'Narrative']) {
      await page.getByRole('tab', { name, exact: true }).click();
      await expect(page.getByRole('tabpanel', { name, exact: true })).toBeVisible();
      await expectDocumentFits(page);
    }
    await page.getByRole('button', { name: 'Send action' }).scrollIntoViewIfNeeded();
    await expect(page.getByRole('button', { name: 'Send action' })).toBeInViewport();
  });
}

test('visual viewport resize budgets the credit and removes its override on navigation', async ({ page, game }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await signIn(page); await page.goto('/play/alpha');
  await expect.poll(() => game.sockets.length).toBe(1);
  await page.getByRole('button', { name: 'Toggle audio controls' }).click();
  const audio = page.getByRole('heading', { name: 'Audio', exact: true }).locator('..');
  const bounds = await audio.boundingBox();
  expect(bounds!.x).toBeGreaterThanOrEqual(0);
  expect(bounds!.x + bounds!.width).toBeLessThanOrEqual(390);
  await page.getByRole('button', { name: 'Toggle audio controls' }).click();
  await page.evaluate(() => {
    Object.defineProperty(visualViewport, 'height', { configurable: true, value: 420 });
    visualViewport!.dispatchEvent(new Event('resize'));
  });
  await expect.poll(() => page.evaluate(() => document.body.getBoundingClientRect().height)).toBe(420);
  await page.getByLabel('Your move').fill('A draft in the reduced viewport');
  const send = page.getByRole('button', { name: 'Send action' });
  await send.scrollIntoViewIfNeeded();
  expect((await send.boundingBox())!.y + (await send.boundingBox())!.height).toBeLessThanOrEqual(420);
  await expect(page.getByRole('contentinfo', { name: 'Software identity' })).toBeInViewport();
  await page.locator('.game-app-header').getByRole('button').last().click();
  await page.getByRole('link', { name: 'Back to campaigns' }).click();
  await expect(page).toHaveURL('/');
  // The URL changes before React necessarily commits the route and cleans up
  // its effect. Check the completed transition, including listener removal.
  await expect(page.getByRole('heading', { name: 'Campaigns', exact: true })).toBeVisible();
  await expect.poll(() => page.evaluate(() => document.documentElement.style.getPropertyValue('--game-viewport-height'))).toBe('');
  await page.evaluate(() => visualViewport!.dispatchEvent(new Event('resize')));
  expect(await page.evaluate(() => document.documentElement.style.getPropertyValue('--game-viewport-height'))).toBe('');
});
