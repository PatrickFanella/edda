import { test, expect, signIn } from './fixtures';
import { campaign, history, result, timestamp } from '../fixtures/data';
import { expectDocumentFits, expectInternalScroll, expectReachable, recordLayout } from './layout';
import type { Page } from '@playwright/test';

const viewports = [[1366, 768], [1440, 900], [1920, 1080], [768, 1024], [360, 800], [390, 844]] as const;
const longChoices = Array.from({ length: 8 }, (_, i) => ({ id: `long-${i}`, text: `Option ${i + 1}: ${'Consider the ancient gate before entering the crowded chamber. '.repeat(20)}` }));
const longEntries = history().map(entry => ({ ...entry, llm_response: `${entry.llm_response} ${'The road continues past the old gate. '.repeat(200)}` }));

async function fixtureHistory(page: Page) {
  await page.route('**/campaigns/alpha/history', route => route.fulfill({ json: { entries: longEntries } }));
}

async function expectNavigation(page: Page) {
  for (const name of ['Play', 'Character', 'World', 'Journal']) {
    const tab = page.getByRole('tab', { name, exact: true });
    await expect(tab).toBeEnabled();
    // The narrow tab strip deliberately scrolls horizontally; prove each
    // destination can be exposed without moving the document.
    await tab.scrollIntoViewIfNeeded();
    await expectReachable(tab);
  }
}

for (const [width, height] of viewports) {
  test(`viewport qualification: story, choices, pending and validation ${width}x${height}`, async ({ page, game }, testInfo) => {
    await page.setViewportSize({ width, height });
    await fixtureHistory(page);
    await signIn(page); await page.goto('/play/alpha');
    await expect.poll(() => game.sockets.length).toBe(1);
    const log = page.getByRole('log');
    await expect(log.getByText(/The alpha gate opens/)).toBeVisible();
    await expectInternalScroll(page, log);
    game.sockets[0].send(JSON.stringify({ type: 'result', timestamp, payload: { ...result, narrative: 'The keeper offers eight long choices.', choices: longChoices } }));
    const choices = page.getByRole('region', { name: 'Suggested choices' });
    await expect(choices.getByRole('button')).toHaveCount(8);
    const input = page.getByLabel('Your move');
    await input.fill('My unsent action\nA second line. ' + 'A long draft. '.repeat(80));
    const send = page.getByRole('button', { name: 'Send action' });
    await expectReachable(input); await expectReachable(send); await expect(send).toBeEnabled();
    await expectNavigation(page);
    expect(await input.evaluate(el => parseFloat(getComputedStyle(el).fontSize))).toBeGreaterThanOrEqual(17);
    await expectInternalScroll(page, choices);
    await choices.getByRole('button').last().scrollIntoViewIfNeeded();
    await expectReachable(choices.getByRole('button').last(), { partial: true });
    await expectReachable(send);
    await recordLayout(page, testInfo, 'long-story-choices', { input, send, lastChoice: choices.getByRole('button').last() });

    // Hold the response after dispatch so pending controls/status are measured.
    game.sockets[0].onMessage(raw => {
      const input = JSON.parse(String(raw)).payload.input;
      if (typeof input === 'string') game.actions.push(input);
    });
    await send.click();
    await expect.poll(() => game.actions.length).toBe(1);
    await expect(input).toBeDisabled();
    await expect(page.getByRole('button', { name: 'Sending…' })).toBeDisabled();
    await expectReachable(page.getByRole('button', { name: 'Sending…' }));
    await expectNavigation(page);
    await recordLayout(page, testInfo, 'pending', { input, send: page.getByRole('button', { name: 'Sending…' }) });
    game.sockets[0].send(JSON.stringify({ type: 'error', timestamp, payload: { error: `Validation: ${'Choose a different action. '.repeat(80)}` } }));
    const feedback = page.locator('.game-composer-feedback');
    await expect(feedback).toContainText('Validation:');
    await expect(input).toBeEnabled();
    await input.fill('A corrected action');
    await expectReachable(send); await expectReachable(input);
    await expectInternalScroll(page, feedback);
    await recordLayout(page, testInfo, 'validation', { input, send });
    game.sockets[0].onMessage(() => game.sockets[0].send(JSON.stringify({ type: 'result', timestamp, payload: result })));
    await send.click();
    await expect(page.getByText(result.narrative, { exact: true })).toBeVisible();
  });

  test(`viewport qualification: context, saves and export ${width}x${height}`, async ({ page, game }, testInfo) => {
    await page.setViewportSize({ width, height });
    await fixtureHistory(page);
    await page.route('**/campaigns/alpha/saves', route => route.fulfill({ json: Array.from({ length: 20 }, (_, i) => ({ id: `save-${i}`, campaign_id: 'alpha', name: `Checkpoint ${i + 1}`, turn_number: i + 1, is_auto: true, created_at: timestamp })) }));
    await page.route('**/campaigns/alpha/export/json', route => route.fulfill({ json: { campaign: campaign(), entries: longEntries } }));
    await signIn(page); await page.goto('/play/alpha');
    await expect.poll(() => game.sockets.length).toBe(1);
    const log = page.getByRole('log');
    await expect(log.getByText(/The alpha gate opens/)).toBeVisible();
    await expectInternalScroll(page, log);
    const readingPosition = await log.evaluate(el => el.scrollTop);
    const input = page.getByLabel('Your move');
    await input.fill('Retain my draft\nand selection');
    await input.evaluate((el: HTMLTextAreaElement) => el.setSelectionRange(3, 12));
    await input.focus();
    for (const name of ['Character', 'World', 'Journal']) {
      await page.getByRole('tab', { name, exact: true }).click();
      const panel = page.getByRole('tabpanel', { name, exact: true });
      await expect(panel).toBeVisible();
      if (name === 'Journal') {
        await panel.locator('summary').filter({ hasText: 'Logs' }).click();
        await expectInternalScroll(page, panel.locator('.overflow-y-auto').first());
      }
      await expectNavigation(page);
      const close = page.getByRole('button', { name: 'Close details' });
      await expectReachable(close);
      await recordLayout(page, testInfo, `details-${name}`, { close });
      await close.click();
      await expect(input).toBeFocused();
      await expect(input).toHaveValue('Retain my draft\nand selection');
      expect(await input.evaluate((el: HTMLTextAreaElement) => [el.selectionStart, el.selectionEnd])).toEqual([3, 12]);
      expect(await log.evaluate(el => el.scrollTop)).toBe(readingPosition);
    }
    await page.getByRole('button', { name: 'Saves', exact: true }).click();
    await expect(page.getByRole('heading', { name: 'Save Points (20)', exact: true })).toBeVisible();
    const saves = page.getByRole('heading', { name: 'Save Points (20)', exact: true }).locator('../..');
    await expectInternalScroll(page, saves);
    await expectReachable(page.getByRole('button', { name: 'Hide Saves', exact: true }));
    // At short heights expanded saves use the existing internal reflow escape.
    await page.getByRole('button', { name: 'Send action' }).scrollIntoViewIfNeeded();
    await expectReachable(page.getByRole('button', { name: 'Send action' }));
    await expectDocumentFits(page);
    await page.getByRole('button', { name: 'Hide Saves', exact: true }).click();
    await page.locator('.game-app-header').getByRole('button').last().click();
    const exportButton = page.getByRole('button', { name: 'Export', exact: true });
    await expectReachable(exportButton); await exportButton.click();
    const dialog = page.getByRole('dialog');
    for (const name of ['JSON Data', 'Session Transcript', 'Character Sheet', 'Close']) {
      const button = dialog.getByRole('button', { name: new RegExp(`^${name}`) });
      await button.scrollIntoViewIfNeeded();
      await expectReachable(button);
    }
    await recordLayout(page, testInfo, 'export', { close: dialog.getByRole('button', { name: 'Close', exact: true }) });
    const downloaded = page.waitForEvent('download');
    await dialog.getByRole('button', { name: /^JSON Data/ }).click();
    const download = await downloaded;
    expect(download.suggestedFilename()).toBe('campaign-json.json');
    await download.saveAs(testInfo.outputPath('fixture-export.json'));
    await dialog.getByRole('button', { name: 'Close', exact: true }).click();
    await expect(dialog).toBeHidden();
    await expectDocumentFits(page);
  });

  test(`viewport qualification: combat controls ${width}x${height}`, async ({ page, game }, testInfo) => {
    await page.setViewportSize({ width, height });
    await fixtureHistory(page);
    await page.route('**/campaigns/alpha', route => route.fulfill({ json: { ...campaign(), rules_mode: 'light' } }));
    await signIn(page); await page.goto('/play/alpha');
    await expect.poll(() => game.sockets.length).toBe(1);
    game.sockets[0].send(JSON.stringify({ type: 'status', timestamp, payload: { stage: 'combat_start', description: 'Fixture combat' } }));
    await expect(page.getByRole('heading', { name: 'Combat Actions', exact: true })).toBeVisible();
    for (const name of ['Attack', 'Defend', 'Cast Spell', 'Use Item', 'Flee']) {
      const button = page.getByRole('button', { name, exact: true });
      await button.scrollIntoViewIfNeeded();
      await expectReachable(button);
    }
    await page.getByRole('button', { name: 'Cast Spell', exact: true }).click();
    const spell = page.getByPlaceholder('Enter spell name...');
    await spell.fill('Light');
    await spell.scrollIntoViewIfNeeded();
    await expectReachable(spell);
    const send = page.getByRole('button', { name: 'Send', exact: true });
    await send.scrollIntoViewIfNeeded();
    await expectReachable(send);
    await expectNavigation(page);
    await recordLayout(page, testInfo, 'combat', { spell, send });
    await send.click();
    await expect.poll(() => game.actions).toEqual(['I cast Light']);
    await expectDocumentFits(page);
  });
}

test.describe('touch emulation (not a software-keyboard session)', () => {
  test.use({ hasTouch: true, isMobile: true });
  for (const width of [360, 390]) {
    test(`tap action and restore context at ${width}`, async ({ page, game }) => {
      await page.setViewportSize({ width, height: width === 360 ? 800 : 844 });
      await signIn(page); await page.goto('/play/alpha');
      await expect.poll(() => game.sockets.length).toBe(1);
      const input = page.getByLabel('Your move');
      await input.tap(); await input.fill('Touch draft\nSecond line');
      await page.getByRole('tab', { name: 'World', exact: true }).tap();
      await expectReachable(page.getByRole('button', { name: 'Close details' }));
      await page.getByRole('button', { name: 'Close details' }).tap();
      await expect(input).toHaveValue('Touch draft\nSecond line'); await expect(input).toBeFocused();
      await expectReachable(page.getByRole('button', { name: 'Send action' }));
      await page.getByRole('button', { name: 'Send action' }).tap();
      await expect.poll(() => game.actions).toEqual(['Touch draft\nSecond line']);
      await page.getByRole('region', { name: 'Suggested choices' }).getByRole('button', { name: /Inspect the key/ }).tap();
      await expect.poll(() => game.actions).toEqual(['Touch draft\nSecond line', 'Inspect the key']);
      await expectDocumentFits(page);
    });
  }
});

for (const [width, height] of [[360, 800], [390, 844], [683, 384]]) {
  test(`enlarged text reflows through internal owners at ${width}x${height}`, async ({ page, game }, testInfo) => {
    await page.setViewportSize({ width, height });
    await fixtureHistory(page);
    await signIn(page); await page.goto('/play/alpha');
    await expect.poll(() => game.sockets.length).toBe(1);
    // 200% root text is distinct from a visible keyboard or browser pinch zoom.
    await page.evaluate(() => { document.documentElement.style.fontSize = '200%'; });
    const input = page.getByLabel('Your move');
    await input.fill('Enlarged-text action\nwith a second line');
    const send = page.getByRole('button', { name: 'Send action' });
    for (const control of [input, send]) {
      await control.scrollIntoViewIfNeeded(); await expectReachable(control, { partial: control === input });
    }
    expect(await input.evaluate(el => parseFloat(getComputedStyle(el).fontSize))).toBeGreaterThanOrEqual(32);
    for (const name of ['Journal', 'Character', 'World', 'Play']) {
      const tab = page.getByRole('tab', { name, exact: true });
      await tab.scrollIntoViewIfNeeded(); await expectReachable(tab); await tab.click();
      await expectDocumentFits(page);
    }
    await send.scrollIntoViewIfNeeded(); await expectReachable(send);
    await recordLayout(page, testInfo, 'enlarged-text', { send });
  });
}
