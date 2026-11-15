import { expect, type Locator, type Page, type TestInfo } from '@playwright/test';

// Do not scroll before measuring: actionability auto-scroll can conceal an
// off-screen composer. Intersect every clipping ancestor and the visual viewport.
export async function controlGeometry(control: Locator) {
  return control.evaluate(el => {
    const rect = el.getBoundingClientRect();
    const viewport = window.visualViewport;
    const visible = {
      left: viewport?.offsetLeft ?? 0, top: viewport?.offsetTop ?? 0,
      right: (viewport?.offsetLeft ?? 0) + (viewport?.width ?? innerWidth),
      bottom: (viewport?.offsetTop ?? 0) + (viewport?.height ?? innerHeight),
    };
    const clips: string[] = [];
    for (let parent = el.parentElement; parent; parent = parent.parentElement) {
      const style = getComputedStyle(parent);
      const box = parent.getBoundingClientRect();
      if (/(auto|scroll|hidden|clip)/.test(style.overflowX)) {
        visible.left = Math.max(visible.left, box.left + parent.clientLeft);
        visible.right = Math.min(visible.right, box.left + parent.clientLeft + parent.clientWidth);
        clips.push(`${parent.tagName}.${parent.className}:x`);
      }
      if (/(auto|scroll|hidden|clip)/.test(style.overflowY)) {
        visible.top = Math.max(visible.top, box.top + parent.clientTop);
        visible.bottom = Math.min(visible.bottom, box.top + parent.clientTop + parent.clientHeight);
        clips.push(`${parent.tagName}.${parent.className}:y`);
      }
    }
    visible.left = Math.max(visible.left, rect.left);
    visible.right = Math.min(visible.right, rect.right);
    visible.top = Math.max(visible.top, rect.top);
    visible.bottom = Math.min(visible.bottom, rect.bottom);
    const hit = document.elementFromPoint((visible.left + visible.right) / 2, (visible.top + visible.bottom) / 2);
    return {
      rect: { x: rect.x, y: rect.y, width: rect.width, height: rect.height },
      visibleWidth: Math.max(0, visible.right - visible.left),
      visibleHeight: Math.max(0, visible.bottom - visible.top),
      hittable: visible.right > visible.left && visible.bottom > visible.top && hit !== null && (hit === el || el.contains(hit)), clips,
      fontSize: parseFloat(getComputedStyle(el).fontSize),
    };
  });
}

export async function expectReachable(control: Locator, { partial = false, touch = true } = {}) {
  await expect(control).toBeVisible();
  const geometry = await controlGeometry(control);
  const message = JSON.stringify(geometry);
  // Oversized choice text may scroll inside its owner. Fixed controls must fit
  // completely, with one CSS pixel tolerance for fractional grid/scroll rounding.
  expect(geometry.visibleWidth, message).toBeGreaterThanOrEqual(partial ? 44 : geometry.rect.width - 1);
  expect(geometry.visibleHeight, message).toBeGreaterThanOrEqual(partial ? 44 : geometry.rect.height - 1);
  expect(geometry.hittable, message).toBe(true);
  if (touch) {
    expect(geometry.rect.width, message).toBeGreaterThanOrEqual(44);
    expect(geometry.rect.height, message).toBeGreaterThanOrEqual(44);
  }
  return geometry;
}

export async function expectDocumentFits(page: Page) {
  const size = await page.evaluate(() => ({
    width: innerWidth, height: innerHeight,
    documentWidth: document.documentElement.scrollWidth, documentHeight: document.documentElement.scrollHeight,
    scrollX, scrollY,
  }));
  expect(size.documentWidth).toBeLessThanOrEqual(size.width + 1);
  expect(size.documentHeight).toBeLessThanOrEqual(size.height + 1);
  expect([size.scrollX, size.scrollY]).toEqual([0, 0]);
  return size;
}

// Wheel the actual scroll owner; record outer scroll positions as well as the
// document so scrolling an intermediate workspace cannot produce a false pass.
export async function expectInternalScroll(page: Page, owner: Locator) {
  const ancestors = () => owner.evaluate(el => {
    const positions = [];
    for (let parent = el.parentElement; parent; parent = parent.parentElement) positions.push([parent.scrollLeft, parent.scrollTop]);
    return positions;
  });
  await owner.evaluate(el => { el.scrollTop = 0; el.dispatchEvent(new Event('scroll')); });
  await expect.poll(() => owner.evaluate(el => el.scrollTop)).toBe(0);
  expect(await owner.evaluate(el => el.scrollHeight - el.clientHeight)).toBeGreaterThan(100);
  const before = await ancestors();
  const box = await owner.boundingBox();
  await page.mouse.move(box!.x + box!.width / 2, box!.y + Math.min(20, box!.height / 2));
  await page.mouse.wheel(0, 180);
  await expect.poll(() => owner.evaluate(el => el.scrollTop)).toBeGreaterThan(0);
  expect(await ancestors()).toEqual(before);
  await expectDocumentFits(page);
}

export async function recordLayout(page: Page, testInfo: TestInfo, name: string, controls: Record<string, Locator>) {
  const measured = Object.fromEntries(await Promise.all(Object.entries(controls).map(async ([key, control]) => [key, await controlGeometry(control)])));
  await testInfo.attach(`${name}.json`, { body: JSON.stringify({ viewport: page.viewportSize(), document: await expectDocumentFits(page), controls: measured }, null, 2), contentType: 'application/json' });
  await page.screenshot({ path: testInfo.outputPath(`${name}.png`) });
}
