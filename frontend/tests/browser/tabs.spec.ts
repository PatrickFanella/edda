import { test, expect, signIn } from './fixtures';
const sections = ['Narrative', 'Character', 'Inventory', 'Quests', 'NPCs', 'World', 'Journal', 'Logs'];
test('all campaign tabs expose stable panels and support automatic keyboard selection', async ({ page, game }) => {
  await signIn(page); await page.goto('/play/alpha');
  await expect.poll(() => game.sockets.length).toBe(1);
  const tabs = page.getByRole('tablist', { name: 'Campaign play sections' });
  const first = tabs.getByRole('tab', { name: 'Narrative', exact: true });
  await first.focus();
  const ids = await tabs.getByRole('tab').evaluateAll(elements => elements.map(e => e.id));
  for (const name of sections) {
    const selected = tabs.getByRole('tab', { name, exact: true });
    await expect(selected).toBeFocused(); await expect(selected).toHaveAttribute('aria-selected', 'true');
    await expect(selected).toHaveAttribute('tabindex', '0');
    const panel = page.getByRole('tabpanel', { name, exact: true });
    await expect(panel).toBeVisible();
    expect(await panel.getAttribute('id')).toBe(await selected.getAttribute('aria-controls'));
    expect(await panel.getAttribute('aria-labelledby')).toBe(await selected.getAttribute('id'));
    await expect(page.getByRole('tabpanel')).toHaveCount(1);
    await expect(tabs.locator('[tabindex="0"]')).toHaveCount(1);
    await selected.press('ArrowRight');
  }
  await expect(first).toBeFocused();
  await first.press('ArrowLeft');
  const last = tabs.getByRole('tab', { name: 'Logs', exact: true });
  await expect(last).toBeFocused(); await last.press('Home'); await expect(first).toBeFocused();
  await first.press('End'); await expect(last).toBeFocused();
  // The existing HUD puts panels before the bottom tab bar in document order.
  await page.keyboard.press('Shift+Tab');
  await expect(page.getByRole('tabpanel', { name: 'Logs' })).toBeFocused();
  await page.keyboard.press('Tab'); await expect(last).toBeFocused();
  await tabs.getByRole('tab', { name: 'World', exact: true }).click();
  await expect(page.getByRole('tabpanel', { name: 'World', exact: true })).toBeVisible();
  expect(await tabs.getByRole('tab').evaluateAll(elements => elements.map(e => e.id))).toEqual(ids);
});
