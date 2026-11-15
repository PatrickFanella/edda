import type { ResolutionEvent, StateChange } from '../api/types';
import type { SequencedEvent } from './useWebSocket';
import type { NarrativeChoice, TurnResponseWithChoices } from './useWebSocket';

export type NarrativeEntryKind = 'player' | 'gm' | 'system';

export interface NarrativeEntry {
  id: string;
  kind: NarrativeEntryKind;
  text: string;
  timestamp: string;
  speaker?: string;
  stateChanges?: StateChange[];
  resolutionEvents?: ResolutionEvent[];
  choices?: NarrativeChoice[];
  isStreaming?: boolean;
}

export type NarrativePhase = 'idle' | 'loading_history' | 'ready' | 'streaming' | 'processing';

export interface NarrativeState {
  phase: NarrativePhase;
  entries: NarrativeEntry[];
  streamingChunks: string[];
  pendingTimestamp: string | null;
  latestResult: TurnResponseWithChoices | null;
  suggestedChoices: NarrativeChoice[];
  error: string | null;
  acknowledgedSequence: number;
}

export const initialNarrativeState: NarrativeState = {
  phase: 'idle',
  entries: [],
  streamingChunks: [],
  pendingTimestamp: null,
  latestResult: null,
  suggestedChoices: [],
  error: null,
  acknowledgedSequence: 0,
};

export type NarrativeAction =
  | { type: 'RESET' }
  | { type: 'LIVE_EVENT'; event: SequencedEvent }
  | { type: 'HISTORY_LOADED'; entries: NarrativeEntry[] }
  | { type: 'HISTORY_FAILED'; error: string }
  | { type: 'ACTION_SENT'; input: string; timestamp: string; entryCount: number }
  | { type: 'CHUNK_RECEIVED'; text: string; timestamp: string }
  | { type: 'RESULT_RECEIVED'; payload: TurnResponseWithChoices; timestamp: string; entryCount: number }
  | { type: 'ERROR_RECEIVED'; error: string; timestamp: string; entryCount: number }
  | { type: 'STATUS_UPDATE' }
  | { type: 'CONNECTION_LOST' }
  | { type: 'SOCKET_ERROR'; error: string }
  | { type: 'SOCKET_ERROR_CLEARED' }
;

export function narrativeReducer(state: NarrativeState, action: NarrativeAction): NarrativeState {
  switch (action.type) {
    case 'LIVE_EVENT': {
      const event = action.event;
      if (event.sequence <= state.acknowledgedSequence) return state;
      const next = { ...state, acknowledgedSequence: event.sequence };
      const timestamp = event.envelope.timestamp;
      switch (event.kind) {
        case 'chunk': return narrativeReducer(next, { type: 'CHUNK_RECEIVED', text: event.payload.text, timestamp });
        case 'result': return narrativeReducer(next, { type: 'RESULT_RECEIVED', payload: event.payload, timestamp, entryCount: event.sequence });
        case 'error': return narrativeReducer(next, { type: 'ERROR_RECEIVED', error: event.payload.error, timestamp, entryCount: event.sequence });
        case 'status': return event.payload.stage === 'narrative_reset'
          ? { ...next, phase: 'processing', streamingChunks: [] } : next;
      }
      return next;
    }
    case 'RESET':
      return initialNarrativeState;

    case 'HISTORY_LOADED':
      return { ...state, phase: state.entries.length ? state.phase : 'ready', entries: [...action.entries.filter(entry => !state.entries.some(live => live.id === entry.id)), ...state.entries], error: state.error };

    case 'HISTORY_FAILED':
      return { ...state, phase: 'ready', error: action.error };

    case 'ACTION_SENT':
      return {
        ...state,
        phase: 'processing',
        entries: [...state.entries, {
          id: `${action.timestamp}-player-${action.entryCount}`,
          kind: 'player',
          text: action.input,
          timestamp: action.timestamp,
          speaker: 'You',
        }],
        streamingChunks: [],
        pendingTimestamp: action.timestamp,
        latestResult: null,
        suggestedChoices: [],
        error: null,
      };

    case 'CHUNK_RECEIVED':
      return {
        ...state,
        phase: 'streaming',
        streamingChunks: [...state.streamingChunks, action.text],
        pendingTimestamp: state.pendingTimestamp ?? action.timestamp,
        error: null,
      };

    case 'RESULT_RECEIVED': {
      const choices = action.payload.choices ?? [];
      const id = action.payload.turn_number ? `history-gm-${action.payload.turn_number}` : `${action.timestamp}-gm-${action.entryCount}`;
      if (state.entries.some(entry => entry.id === id)) return { ...state, phase: 'ready', streamingChunks: [], pendingTimestamp: null };
      return {
        ...state,
        phase: 'ready',
        entries: [...state.entries.map(entry => action.payload.turn_number && state.pendingTimestamp && entry.kind === 'player' && entry.timestamp === state.pendingTimestamp
          ? { ...entry, id: `history-player-${action.payload.turn_number}` } : entry), {
          id,
          kind: 'gm',
          text: action.payload.narrative,
          timestamp: action.timestamp,
          speaker: 'Game Master',
          stateChanges: action.payload.state_changes,
          resolutionEvents: action.payload.resolution_events,
          choices,
        }],
        streamingChunks: [],
        pendingTimestamp: null,
        latestResult: action.payload,
        suggestedChoices: choices,
        error: null,
      };
    }

    case 'ERROR_RECEIVED':
      return {
        ...state,
        phase: 'ready',
        streamingChunks: [],
        pendingTimestamp: null,
        suggestedChoices: [],
        error: action.error,
        entries: [...state.entries, {
          id: `${action.timestamp}-system-${action.entryCount}`,
          kind: 'system',
          text: action.error,
          timestamp: action.timestamp,
          speaker: 'System',
        }],
      };

    case 'STATUS_UPDATE':
      return state;

    case 'CONNECTION_LOST':
      return { ...state, phase: 'ready', streamingChunks: [], pendingTimestamp: null };

    case 'SOCKET_ERROR':
      return { ...state, phase: 'ready', streamingChunks: [], error: action.error, pendingTimestamp: null };

    case 'SOCKET_ERROR_CLEARED':
      return { ...state, error: null };


    default:
      return state;
  }
}
