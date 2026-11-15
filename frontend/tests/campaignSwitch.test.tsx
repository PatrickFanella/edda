import { act, renderHook } from '@testing-library/react';
import { expect, it, vi } from 'vitest';
import { getSessionHistory } from '../src/api/campaigns';
import type { SessionHistoryResponse } from '../src/api/types';
import { useNarrative } from '../src/hooks/useNarrative';
import { useCampaign } from '../src/hooks/useCampaign';
import { history, result, timestamp } from './fixtures/data';

vi.mock('../src/api/campaigns', () => ({ getSessionHistory: vi.fn() }));
vi.mock('../src/hooks/useCampaign', () => ({ useCampaign: vi.fn() }));

class Socket extends EventTarget {
  static OPEN = 1;
  static CONNECTING = 0;
  static instances: Socket[] = [];
  readyState = 0;
  constructor() { super(); Socket.instances.push(this); }
  open() { this.readyState = 1; this.dispatchEvent(new Event('open')); }
  close() { this.readyState = 3; this.dispatchEvent(new Event('close')); }
  send() {}
  receive(type: string, payload: unknown) {
    this.dispatchEvent(new MessageEvent('message', { data: JSON.stringify({ type, payload, timestamp }) }));
  }
}

it.each(['resolve', 'reject'] as const)('isolates a direct campaign switch from late history (%s) and retained live state', async (outcome) => {
  vi.useFakeTimers(); Socket.instances = []; vi.stubGlobal('WebSocket', Socket);
  let resolveAlpha!: (response: SessionHistoryResponse) => void;
  let rejectAlpha!: (error: Error) => void;
  vi.mocked(getSessionHistory).mockImplementation((id) => id === 'alpha'
    ? new Promise((resolve, reject) => { resolveAlpha = resolve; rejectAlpha = reject; })
    : Promise.resolve({ entries: history(id) }));
  vi.mocked(useCampaign).mockReturnValue({ campaignId: 'alpha' } as ReturnType<typeof useCampaign>);
  const hook = renderHook(() => useNarrative());
  await act(() => vi.advanceTimersByTimeAsync(0));
  const alpha = Socket.instances[0]; act(() => alpha.open());
  act(() => alpha.receive('result', result));
  act(() => alpha.receive('status', { stage: 'combat_start', description: 'Alpha combat' }));
  act(() => { hook.result.current.sendAction('Alpha draft'); });
  expect(hook.result.current.suggestedChoices).toEqual([]);
  expect(hook.result.current.combatActive).toBe(true);
  expect(hook.result.current.isLoading).toBe(true);

  vi.mocked(useCampaign).mockReturnValue({ campaignId: 'beta' } as ReturnType<typeof useCampaign>);
  hook.rerender();
  await act(() => vi.advanceTimersByTimeAsync(0));
  act(() => Socket.instances[1].open());
  const betaEntries = hook.result.current.entries;
  expect(betaEntries.map((entry) => entry.text).join(' ')).toContain('beta');
  expect(betaEntries.map((entry) => entry.text).join(' ')).not.toContain(result.narrative);
  expect(hook.result.current.suggestedChoices).toEqual([]);
  expect(hook.result.current.latestResult).toBeNull();
  expect(hook.result.current.currentStatus).toBeNull();
  expect(hook.result.current.combatActive).toBe(false);
  expect(hook.result.current.isLoading).toBe(false);
  expect(alpha.readyState).toBe(3);
  await act(async () => {
    if (outcome === 'resolve') resolveAlpha({ entries: history('alpha') });
    else rejectAlpha(new Error('Late alpha failure'));
    alpha.receive('result', result);
  });
  expect(hook.result.current.entries).toEqual(betaEntries);
  expect(hook.result.current.error).toBeNull();

  hook.unmount();
  const remounted = renderHook(() => useNarrative());
  await act(() => vi.advanceTimersByTimeAsync(0));
  act(() => Socket.instances[2].open());
  expect(remounted.result.current.entries).toEqual(betaEntries);
  expect(remounted.result.current.combatActive).toBe(false);
});

it('consumes a burst beyond the bound in order and keeps live completion during history loading', async () => {
  vi.useFakeTimers(); Socket.instances = []; vi.stubGlobal('WebSocket', Socket);
  let resolveHistory!: (response: SessionHistoryResponse) => void;
  vi.mocked(getSessionHistory).mockReturnValue(new Promise(resolve => { resolveHistory = resolve; }));
  vi.mocked(useCampaign).mockReturnValue({ campaignId: 'alpha' } as ReturnType<typeof useCampaign>);
  const hook = renderHook(() => useNarrative());
  await act(() => vi.advanceTimersByTimeAsync(0));
  const socket = Socket.instances[0]; act(() => socket.open());
  act(() => {
    for (let n = 1; n <= 600; n++) {
      socket.receive('chunk', { text: `Chunk ${n}` });
      socket.receive('result', { ...result, narrative: `Result ${n}`, turn_number: n + 2 });
      socket.receive('error', { error: `Error ${n}` });
    }
  });
  expect(hook.result.current.entries).toHaveLength(1200);
  expect(new Set(hook.result.current.entries.map(entry => entry.id)).size).toBe(1200);
  await act(async () => resolveHistory({ entries: [...history(), { turn_number: 3, player_input: '', input_type: 'narrative', llm_response: 'Result 1', created_at: timestamp }] }));
  expect(hook.result.current.entries.filter(entry => entry.text === 'Result 1')).toHaveLength(1);
  expect(hook.result.current.entries).toHaveLength(1203);
  act(() => socket.close());
  await act(() => vi.advanceTimersByTimeAsync(1000)); await act(() => vi.advanceTimersByTimeAsync(0));
  act(() => { Socket.instances[1].open(); Socket.instances[1].receive('result', { ...result, narrative: 'After reconnect', turn_number: 603 }); });
  expect(hook.result.current.entries.at(-1)?.text).toBe('After reconnect');
  hook.unmount(); socket.receive('result', result);
});
