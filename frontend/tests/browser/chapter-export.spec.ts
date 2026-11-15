import { readFile } from 'node:fs/promises';
import { test, expect, signIn } from './fixtures';
import { chapterExportFixture } from '../fixtures/chapterExport';
import { expectDocumentFits, expectReachable } from './layout';

for (const width of [390, 1440]) {
  test(`downloads the selected saved reader revision at ${width}`, async ({ page, game }, testInfo) => {
    await page.setViewportSize({ width, height: 844 });
    const chapter = chapterExportFixture(1), other = chapterExportFixture(2);
    const requests: { method: string; ids: string[]; format: string | null }[] = [];
    await page.route('**/api/v1/campaigns/alpha/journal/chapters', route => route.fulfill({ json: [other, chapter] }));
    await page.route('**/api/v1/campaigns/alpha/export/chapters?**', route => {
      const url = new URL(route.request().url());
      requests.push({ method: route.request().method(), ids: url.searchParams.getAll('revision_id'), format: url.searchParams.get('format') });
      return route.fulfill({ contentType: 'application/json', json: { campaign_id: 'alpha', chapters: [{ ...chapter, revisions: [chapter.revisions[0]] }], attribution: 'Fixture attribution', changes: 'Saved prose; no generation.' } });
    });
    await signIn(page); await page.goto('/play/alpha');
    await page.getByRole('tab', { name: 'Journal', exact: true }).click();
    await page.getByRole('combobox', { name: 'Saved revision', exact: true }).selectOption(chapter.revisions[0].id);
    await expect(page.getByRole('region', { name: 'Chapter prose' })).toContainText(chapter.revisions[0].content.split('\n')[0]);
    await page.getByRole('button', { name: 'Close details' }).click();
    await page.locator('.game-app-header').getByRole('button').last().click();
    await page.getByRole('button', { name: 'Export', exact: true }).click();
    const dialog = page.getByRole('dialog');
    await dialog.getByText('Saved Chapters', { exact: true }).click();
    const download = dialog.getByRole('button', { name: 'Download selected chapters' });
    await expect(download).toBeDisabled();
    await dialog.getByRole('checkbox', { name: 'Chapter 1 · Turns 1–1', exact: true }).check();
    await expect(dialog.getByRole('combobox', { name: /Revision for Chapter 1/ })).toHaveValue(chapter.revisions[0].id);
    await expect(dialog.getByRole('combobox', { name: /Revision for Chapter 1/ }).getByRole('option')).toHaveCount(2);
    await dialog.getByLabel('Chapter file format').selectOption('json');
    await download.scrollIntoViewIfNeeded();
    await expectReachable(download);
    const waiting = page.waitForEvent('download');
    await download.click();
    const file = await waiting;
    expect(file.suggestedFilename()).toBe('campaign-chapters.json');
    const path = testInfo.outputPath('saved-chapters.json');
    await file.saveAs(path);
    const payload = JSON.parse(await readFile(path, 'utf8'));
    expect(payload.chapters[0].revisions[0].content).toBe(chapter.revisions[0].content);
    expect(requests).toEqual([{ method: 'GET', ids: [chapter.revisions[0].id], format: 'json' }]);
    expect(game.actions).toEqual([]);
    expect(game.requests.filter(request => request.startsWith('POST /campaigns/alpha/'))).toEqual([]);
    await dialog.getByRole('button', { name: 'Close', exact: true }).click();
    await expectDocumentFits(page);
  });
}

test('failed download keeps selection available for retry', async ({ page, game }) => {
  const chapter = chapterExportFixture(1);
  await page.route('**/api/v1/campaigns/alpha/journal/chapters', route => route.fulfill({ json: [chapter] }));
  let attempts = 0;
  await page.route('**/api/v1/campaigns/alpha/export/chapters?**', route => {
    attempts++;
    return attempts === 1 ? route.fulfill({ status: 404, json: { error: 'selected saved revision not available' } }) : route.fulfill({ contentType: 'text/markdown', body: chapter.revisions[1].content });
  });
  await signIn(page); await page.goto('/play/alpha');
  await page.locator('.game-app-header').getByRole('button').last().click();
  await page.getByRole('button', { name: 'Export', exact: true }).click();
  const dialog = page.getByRole('dialog');
  await dialog.getByText('Saved Chapters', { exact: true }).click();
  await dialog.getByRole('checkbox').check();
  const download = dialog.getByRole('button', { name: 'Download selected chapters' });
  await download.click();
  await expect(dialog.getByRole('alert')).toContainText('selected saved revision not available');
  await expect(dialog.getByRole('checkbox')).toBeChecked();
  const waiting = page.waitForEvent('download');
  await download.click();
  expect((await waiting).suggestedFilename()).toBe('campaign-chapters.md');
  expect(attempts).toBe(2);
  expect(game.actions).toEqual([]);
});
