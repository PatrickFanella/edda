import { act, renderHook } from '@testing-library/react';
import { expect, it, vi } from 'vitest';
import { MAX_PENDING_EVENTS, useWebSocket } from '../src/hooks/useWebSocket';
import { timestamp } from './fixtures/data';

class RetentionSocket extends EventTarget {
  static instances: RetentionSocket[] = [];
  static OPEN = 1;
  static CONNECTING = 0;
  readyState = 0;
  constructor() { super(); RetentionSocket.instances.push(this); }
  open() { this.readyState = 1; this.dispatchEvent(new Event('open')); }
  close() { this.readyState = 3; this.dispatchEvent(new Event('close')); }
  frame(type: string, payload: unknown) {
    this.dispatchEvent(new MessageEvent('message', { data: JSON.stringify({ type, timestamp, payload }) }));
  }
}


it('releases consumed transport frames independently of completed turns', async () => {
  vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout', 'Date'] }); RetentionSocket.instances = []; vi.stubGlobal('WebSocket', RetentionSocket);
  let consumed = 0;
  const hook = renderHook(({ id }) => useWebSocket(id, () => { consumed++; }), { initialProps: { id: 'alpha' } });
  await act(() => vi.advanceTimersByTimeAsync(0));
  const socket = RetentionSocket.instances[0]; act(() => socket.open());
  const measurements: unknown[] = [];
  let checkpoint = performance.now();
  for (let turn = 1; turn <= 1000; turn++) {
    for (let status = 0; status < 10; status++) {
      act(() => socket.frame('status', { stage: 'tool_execution', description: 'Inspecting the gate', tool: 'probe' }));
    }
    act(() => socket.frame('result', { narrative: `Turn ${turn}: ${'A quiet road. '.repeat(40)}`, state_changes: [], choices: [] }));
    if (turn === 100 || turn === 1000) {
      const events = hook.result.current.events;
      expect(events).toHaveLength(0);
      expect(consumed).toBe(turn * 11);
      
      expect(hook.result.current.currentStatus).toBeNull();
      const now = performance.now();
      measurements.push({ turns: turn, frames: events.length, serializedBytes: Buffer.byteLength(JSON.stringify(events)), copiedReferences: events.length * (events.length - 1) / 2, intervalMs: Math.round(now - checkpoint) });
      checkpoint = now;
    }
  }
  console.info('RETENTION_BOUNDED', JSON.stringify(measurements));
  act(() => socket.close());
  await act(() => vi.advanceTimersByTimeAsync(1000));
  await act(() => vi.advanceTimersByTimeAsync(0));
  act(() => RetentionSocket.instances[1].open());
  expect(hook.result.current.events).toHaveLength(0);
  hook.rerender({ id: 'beta' });
  expect(hook.result.current.events).toHaveLength(0);
});

it('acknowledges pending sequences and stops an abandoned consumer at the bound', async () => {
  vi.useFakeTimers(); RetentionSocket.instances = []; vi.stubGlobal('WebSocket', RetentionSocket);
  const hook = renderHook(() => useWebSocket('alpha'));
  await act(() => vi.advanceTimersByTimeAsync(0));
  const socket = RetentionSocket.instances[0]; act(() => socket.open());
  act(() => { for (let n = 0; n < MAX_PENDING_EVENTS; n++) socket.frame('error', { error: `Error ${n}` }); });
  expect(hook.result.current.events).toHaveLength(MAX_PENDING_EVENTS);
  act(() => hook.result.current.acknowledge(128));
  expect(hook.result.current.events[0].sequence).toBe(129);
  act(() => { for (let n = 0; n <= MAX_PENDING_EVENTS; n++) socket.frame('error', { error: `Next ${n}` }); });
  expect(hook.result.current.events).toHaveLength(MAX_PENDING_EVENTS);
  expect(hook.result.current.error).toContain('refresh');
  expect(hook.result.current.sendAction('Retry')).toBe(false);
  hook.unmount();
});
