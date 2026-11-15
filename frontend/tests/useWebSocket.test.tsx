import { act, renderHook } from '@testing-library/react';
import { expect, it, vi } from 'vitest';
import { useWebSocket } from '../src/hooks/useWebSocket';
import { result, timestamp } from './fixtures/data';

class FixtureSocket extends EventTarget {
  static OPEN = 1;
  static CONNECTING = 0;
  static instances: FixtureSocket[] = [];
  readyState = 0;
  sent: string[] = [];
  url: string;
  constructor(url: string) { super(); this.url = url; FixtureSocket.instances.push(this); }
  send(message: string) { this.sent.push(message); }
  close() { this.readyState = 3; this.dispatchEvent(new Event('close')); }
  open() { this.readyState = 1; this.dispatchEvent(new Event('open')); }
  receive(payload: unknown) { this.dispatchEvent(new MessageEvent('message', { data: JSON.stringify({ type: 'result', timestamp, payload }) })); }
}
it('completes turns, reconnects, and closes the old campaign socket on switch/unmount', async () => {
  vi.useFakeTimers(); FixtureSocket.instances = []; vi.stubGlobal('WebSocket', FixtureSocket);
  const { result: hook, rerender, unmount } = renderHook(({ id }) => useWebSocket(id), { initialProps: { id: 'alpha' } });
  await act(() => vi.advanceTimersByTimeAsync(0));
  const first = FixtureSocket.instances[0]; act(() => first.open());
  act(() => { expect(hook.current.sendAction(' Look ')).toBe(true); });
  expect(JSON.parse(first.sent[0])).toEqual({ type: 'action', payload: { input: 'Look', action_id: expect.any(String) } });
  expect(hook.current.isLoading).toBe(true);
  act(() => first.receive(result)); expect(hook.current.isLoading).toBe(false);
  expect(hook.current.events).toHaveLength(1);
  act(() => first.close()); expect(hook.current.connectionStatus).toBe('closed');
  await act(() => vi.advanceTimersByTimeAsync(1000));
  await act(() => vi.advanceTimersByTimeAsync(0));
  const second = FixtureSocket.instances[1]; act(() => second.open());
  expect(hook.current.isConnected).toBe(true);
  rerender({ id: 'beta' }); await act(() => vi.advanceTimersByTimeAsync(0));
  expect(second.readyState).toBe(3);
  const third = FixtureSocket.instances[2]; expect(third.url).toContain('/beta/ws');
  act(() => third.open()); unmount(); expect(third.readyState).toBe(3);
  await act(() => vi.advanceTimersByTimeAsync(30000)); expect(FixtureSocket.instances).toHaveLength(3);
});

it('reuses the action ID for recovery and keeps a new action distinct', async () => {
 vi.useFakeTimers(); FixtureSocket.instances = []; vi.stubGlobal('WebSocket', FixtureSocket);
 const hook = renderHook(() => useWebSocket('alpha'));
 await act(() => vi.advanceTimersByTimeAsync(0));
 const socket = FixtureSocket.instances[0]; act(() => socket.open());
 act(() => { hook.result.current.sendAction('Look'); });
 const first = JSON.parse(socket.sent[0]);
 act(() => socket.dispatchEvent(new MessageEvent('message', { data: JSON.stringify({ type: 'error', timestamp, payload: { error: 'Recover recorded outcomes', action_id: first.payload.action_id } }) })));
 act(() => { hook.result.current.sendAction('Look'); });
 expect(JSON.parse(socket.sent[1]).payload.action_id).toBe(first.payload.action_id);
 act(() => socket.receive(result));
 act(() => { hook.result.current.sendAction('Look'); });
 expect(JSON.parse(socket.sent[2]).payload.action_id).not.toBe(first.payload.action_id);
});

it('adopts a durable recovery ID returned after a fresh hook sends an action', async () => {
 vi.useFakeTimers(); FixtureSocket.instances = []; vi.stubGlobal('WebSocket', FixtureSocket);
 const hook = renderHook(() => useWebSocket('alpha'));
 await act(() => vi.advanceTimersByTimeAsync(0));
 const socket = FixtureSocket.instances[0]; act(() => socket.open());
 act(() => { hook.result.current.sendAction('Original action'); });
 const canonical = '12345678-1234-4321-8123-123456789012';
 act(() => socket.dispatchEvent(new MessageEvent('message', { data: JSON.stringify({ type: 'error', timestamp, payload: { error: 'Recover recorded outcomes', action_id: canonical } }) })));
 act(() => { hook.result.current.sendAction('Original action'); });
 expect(JSON.parse(socket.sent[1]).payload.action_id).toBe(canonical);
});
