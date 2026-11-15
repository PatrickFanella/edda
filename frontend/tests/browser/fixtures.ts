import { test as base, expect, type WebSocketRoute } from '@playwright/test';
import { campaign, character, history, heroProfile, profile, result, timestamp, user } from '../fixtures/data';

export const test = base.extend<{ game: { sockets: WebSocketRoute[]; actions: string[]; requests: string[] } }>({
  game: async ({ page }, provide) => {
    const sockets: WebSocketRoute[] = [], actions: string[] = [], requests: string[] = [];
    let created = false;
    const unmapped: string[] = [];
    // Fail closed: no API request, socket or external resource reaches a real service.
    await page.route('**/*', async route => {
      const url = new URL(route.request().url());
      if (url.origin !== 'http://127.0.0.1:4175') return route.abort('blockedbyclient');
      if (!url.pathname.startsWith('/api/')) return route.continue();
      const path = url.pathname.replace('/api/v1', '');
      requests.push(`${route.request().method()} ${path}`);
      let body: unknown;
      if (path === '/auth/me') body = { user };
      else if (path === '/auth/login' || path === '/auth/register') body = { token: 'fixture-token', user };
      else if (path === '/auth/logout') body = {};
      else if (path === '/campaigns') body = { campaigns: [campaign(), campaign('beta')] };
      else if (path === '/campaigns/start/name') body = { name: 'Adventure alpha' };
      else if (path === '/campaigns/start/world') {
        created = true;
        body = { campaign: campaign(), opening_scene: { narrative: 'A lantern lights the road.', choices: ['Follow the light'] } };
      }
      else if (path.startsWith('/campaigns/start/')) {
        const isHero = path.includes('character-interview');
        const done = path.endsWith('/interview');
        body = { session_id: 'interview', message: done ? 'Ready to begin.' : 'What would you like?', done, ...(done ? { profile: isHero ? heroProfile : profile } : {}) };
      } else {
        const match = path.match(/^\/campaigns\/([^/]+)(.*)$/);
        const id = match?.[1] ?? '', resource = match?.[2];
        if (resource === '') body = campaign(id);
        else if (resource === '/history') body = { entries: created ? [] : history(id) };
        else if (resource === '/character') body = character(id);
        else if (resource === '/map') body = { locations: [], connections: [] };
        else if (resource === '/time') body = { day: 1, hour: 12, minute: 0 };
        else if (['/quests', '/character/inventory', '/character/abilities', '/character/feats', '/character/skills', '/npcs', '/npcs/encountered', '/journal/summaries', '/journal/entries', '/locations', '/facts', '/journal', '/encountered', '/relationships'].includes(resource ?? '')) body = [];
        else {
          unmapped.push(path);
          return route.fulfill({ status: 501, json: { error: `Unmapped fixture: ${path}` } });
        }
      }
      await route.fulfill({ json: body });
    });
    await page.routeWebSocket('**/*', async socket => {
      if (!new URL(socket.url()).pathname.match(/^\/api\/v1\/campaigns\/[^/]+\/ws$/)) { await socket.close(); return; }
      sockets.push(socket);
      socket.onMessage(raw => {
        const action = JSON.parse(String(raw));
        actions.push(action.payload.input);
        socket.send(JSON.stringify({ type: 'result', timestamp, payload: result }));
      });
    });
    await provide({ sockets, actions, requests });
    expect(unmapped, 'Add explicit fixtures for new API resources').toEqual([]);
  },
});
export { expect, user };
export async function signIn(page: import('@playwright/test').Page) {
  await page.goto('/login');
  await page.getByLabel('Email').fill(user.email);
  await page.getByLabel('Password', { exact: true }).fill('fixture-password');
  await page.getByRole('button', { name: /^Sign in$/i }).click();
  await expect(page.getByRole('heading', { name: 'Campaigns', exact: true })).toBeVisible();
}
