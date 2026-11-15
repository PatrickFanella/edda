import { act, renderHook } from '@testing-library/react';
import { expect, it, vi } from 'vitest';
import { useWebSocket } from '../src/hooks/useWebSocket';
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

// Characterization, not a desired retention contract. Replace these baseline
// assertions with bounded assertions when the #43 follow-up implements acking.
it('measures lifetime transport retention independently of durable history', async () => {
  vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout', 'Date'] }); RetentionSocket.instances = []; vi.stubGlobal('WebSocket', RetentionSocket);
  const hook = renderHook(({ id }) => useWebSocket(id), { initialProps: { id: 'alpha' } });
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
      expect(events).toHaveLength(turn * 11);
      expect(events.filter(event => event.kind === 'result')).toHaveLength(turn);
      expect(hook.result.current.currentStatus).toBeNull();
      const now = performance.now();
      measurements.push({ turns: turn, frames: events.length, serializedBytes: Buffer.byteLength(JSON.stringify(events)), copiedReferences: events.length * (events.length - 1) / 2, intervalMs: Math.round(now - checkpoint) });
      checkpoint = now;
    }
  }
  console.info('RETENTION_BASELINE', JSON.stringify(measurements));
  act(() => socket.close());
  await act(() => vi.advanceTimersByTimeAsync(1000));
  await act(() => vi.advanceTimersByTimeAsync(0));
  act(() => RetentionSocket.instances[1].open());
  expect(hook.result.current.events).toHaveLength(11000);
  hook.rerender({ id: 'beta' });
  expect(hook.result.current.events).toHaveLength(0);
});
