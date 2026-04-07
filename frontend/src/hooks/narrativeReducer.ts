import type { StateChange } from '../api/types';
import type { NarrativeChoice, TurnResponseWithChoices } from './useWebSocket';

export type NarrativeEntryKind = 'player' | 'gm' | 'system';

export interface NarrativeEntry {
  id: string;
  kind: NarrativeEntryKind;
  text: string;
  timestamp: string;
  speaker?: string;
  stateChanges?: StateChange[];
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
  processedEventCount: number;
}

export const initialNarrativeState: NarrativeState = {
  phase: 'idle',
  entries: [],
  streamingChunks: [],
  pendingTimestamp: null,
  latestResult: null,
  suggestedChoices: [],
  error: null,
  processedEventCount: 0,
};

export type NarrativeAction =
  | { type: 'RESET' }
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
  | { type: 'EVENTS_PROCESSED'; count: number };

export function narrativeReducer(state: NarrativeState, action: NarrativeAction): NarrativeState {
  switch (action.type) {
    case 'RESET':
      return initialNarrativeState;

    case 'HISTORY_LOADED':
      return { ...state, phase: 'ready', entries: action.entries, error: null };

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
      return {
        ...state,
        phase: 'ready',
        entries: [...state.entries, {
          id: `${action.timestamp}-gm-${action.entryCount}`,
          kind: 'gm',
          text: action.payload.narrative,
          timestamp: action.timestamp,
          speaker: 'Game Master',
          stateChanges: action.payload.state_changes,
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
      return { ...state, pendingTimestamp: null };

    case 'SOCKET_ERROR':
      return { ...state, error: action.error, pendingTimestamp: null };

    case 'SOCKET_ERROR_CLEARED':
      return { ...state, error: null };

    case 'EVENTS_PROCESSED':
      return { ...state, processedEventCount: action.count };

    default:
      return state;
  }
}
