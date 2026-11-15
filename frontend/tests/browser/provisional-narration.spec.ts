import { test, expect, signIn } from './fixtures';
import { result, timestamp } from '../fixtures/data';

for (const width of [390, 1366]) {
  test(`shows fragments before completion and replaces corrected text at ${width}`, async ({ page, game }) => {
    await page.setViewportSize({ width, height: 844 });
    await signIn(page); await page.goto('/play/alpha');
    await expect.poll(() => game.sockets.length).toBe(1);
    const socket = game.sockets[0];
    const actions: { input: string; action_id: string }[] = [];
    // Hold completion until the browser has rendered each provider fragment.
    socket.onMessage(raw => actions.push(JSON.parse(String(raw)).payload));
    const frame = (type: string, payload: unknown) => socket.send(JSON.stringify({ type, payload, timestamp }));
    const input = page.getByLabel('Your move');
    await input.fill('Search'); await input.press('Enter');
    await expect.poll(() => actions.length).toBe(1);
    const provisional = page.getByLabel('Provisional narration');
    frame('chunk', { text: 'You enter ' });
    await expect(provisional).toContainText('You enter');
    frame('chunk', { text: 'the castle.' });
    await expect(provisional).toContainText('You enter the castle.');
    await expect(provisional).toContainText('Provisional · not saved');
    await expect(provisional.getByText('Turn details')).toHaveCount(0);
    await expect(input).toBeDisabled();
    frame('status', { stage: 'narrative_reset', description: 'Correcting a tool outcome' });
    await expect(page.getByText('You enter the castle.')).toHaveCount(0);
    frame('chunk', { text: 'The gate stays shut.' });
    await expect(provisional).toContainText('The gate stays shut.');
    frame('error', { error: 'Narration interrupted. Retry the same action to recover.', action_id: actions[0].action_id });
    await expect(provisional).toHaveCount(0);
    await expect(input).toBeEnabled();
    await input.fill('Search'); await input.press('Enter');
    await expect.poll(() => actions.length).toBe(2);
    expect(actions[1]).toEqual(actions[0]);
    frame('chunk', { text: 'Recovered draft.' });
    await expect(provisional).toContainText('Recovered draft.');
    frame('result', { ...result, turn_number: 3, recovered: true, action_id: actions[0].action_id });
    await expect(provisional).toHaveCount(0);
    await expect(page.getByText(result.narrative, { exact: true })).toBeVisible();
    await expect(page.getByText('Recovered draft.')).toHaveCount(0);
    await expect(page.getByText('The gate stays shut.')).toHaveCount(0);
    await expect(page.getByText('Provisional · not saved')).toHaveCount(0);
    await expect(input).toBeEnabled();
  });
}
