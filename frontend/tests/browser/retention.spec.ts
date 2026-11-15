import { test, expect, signIn } from './fixtures';
import { timestamp } from '../fixtures/data';

test('retention baseline preserves history and completed turns across reconnect', async ({ page, game }) => {
  // Hosted 2-core runners need a bounded 60s budget for 1,100 live frames.
  test.setTimeout(60_000);
  await signIn(page); await page.goto('/play/alpha');
  await expect(page.getByText('The alpha gate opens.', { exact: true })).toBeVisible();
  await expect.poll(() => game.sockets.length).toBe(1);
  for (let turn = 1; turn <= 100; turn++) {
    for (let status = 0; status < 10; status++) {
      game.sockets[0].send(JSON.stringify({ type: 'status', timestamp, payload: { stage: 'thinking', description: 'Fixture only' } }));
    }
    game.sockets[0].send(JSON.stringify({ type: 'result', timestamp: new Date(Date.parse(timestamp) + turn * 1000).toISOString(), payload: { narrative: `Retained turn ${turn}`, state_changes: [], choices: [] } }));
  }
  await expect(page.getByText('Retained turn 100', { exact: true })).toBeVisible();
  await expect(page.getByText(/^Retained turn \d+$/)).toHaveCount(100);
  await game.sockets[0].close({ code: 1012, reason: 'retention fixture reconnect' });
  await expect.poll(() => game.sockets.length).toBe(2);
  await expect(page.getByText(/^Retained turn \d+$/)).toHaveCount(100);
  await expect(page.getByText('The alpha gate opens.', { exact: true })).toHaveCount(1);
  await page.getByLabel('Your move').fill('Look again');
  await page.getByLabel('Your move').press('Enter');
  await expect(page.getByText('You find a silver key.', { exact: true })).toBeVisible();
  await expect(page.getByText(/^Retained turn \d+$/)).toHaveCount(100);
});
