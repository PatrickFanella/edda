import { expect, it } from 'vitest';
import { narrativeReducer, initialNarrativeState } from '../src/hooks/narrativeReducer';
import { decodeWebSocketEvent } from '../src/api/websocketParser';
import { result, timestamp } from './fixtures/data';
it('commits a streamed turn once and exposes its next choices', () => {
  let state = narrativeReducer(initialNarrativeState, { type: 'ACTION_SENT', input: 'Search', timestamp, entryCount: 0 });
  state = narrativeReducer(state, { type: 'CHUNK_RECEIVED', text: 'You find', timestamp });
  expect(state.phase).toBe('streaming');
  const decoded = decodeWebSocketEvent(JSON.stringify({ type: 'result', timestamp, payload: result }));
  if (decoded?.kind !== 'result') throw new Error('Invalid fixture');
  state = narrativeReducer(state, { type: 'RESULT_RECEIVED', payload: decoded.payload, timestamp, entryCount: 1 });
  expect(state.entries.map(e => e.text)).toEqual(['Search', result.narrative]);
  expect(state.streamingChunks).toEqual([]); expect(state.pendingTimestamp).toBeNull();
  expect(state.suggestedChoices).toEqual(result.choices);
});
it('rejects malformed live messages and resets narrative state', () => {
  expect(decodeWebSocketEvent('{')).toBeNull();
  expect(decodeWebSocketEvent(JSON.stringify({ type: 'result', timestamp, payload: { narrative: 'missing changes' } }))).toBeNull();
  expect(narrativeReducer({ ...initialNarrativeState, error: 'old campaign' }, { type: 'RESET' })).toEqual(initialNarrativeState);
});
