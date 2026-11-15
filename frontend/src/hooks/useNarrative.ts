import { useCallback, useEffect, useMemo, useState, useSyncExternalStore } from 'react';

import { getSessionHistory } from '../api/campaigns';
import type { TurnResponse, WebSocketStatusPayload } from '../api/types';
import { useCampaign } from './useCampaign';
import { useWebSocket, type ConnectionStatus, type NarrativeChoice, type TurnResponseWithChoices } from './useWebSocket';
import {
  narrativeReducer,
  initialNarrativeState,
  type NarrativeEntry,
  type NarrativeEntryKind,
  type NarrativeAction,
} from './narrativeReducer';

export type { NarrativeEntry, NarrativeEntryKind };

export interface UseNarrativeResult {
  campaignId: string | null;
  connectionStatus: ConnectionStatus;
  entries: NarrativeEntry[];
  streamingEntry: NarrativeEntry | null;
  latestResult: TurnResponseWithChoices | null;
  suggestedChoices: NarrativeChoice[];
  currentStatus: WebSocketStatusPayload | null;
  combatActive: boolean;
  isLoading: boolean;
  error: string | null;
  sendAction: (input: string) => boolean;
}

export function useNarrative(): UseNarrativeResult {
  const { campaignId } = useCampaign();
  const [store, setStore] = useState(() => createNarrativeStore(campaignId));
  if (store.campaignId !== campaignId) setStore(createNarrativeStore(campaignId));
  const state = useSyncExternalStore(store.subscribe, store.getSnapshot);
  const dispatch = store.dispatch;
  const { connectionStatus, error: socketError, isLoading, currentStatus, combatActive, sendAction: sendSocketAction } = useWebSocket(campaignId, (event) => {
    dispatch({ type: 'LIVE_EVENT', event });
  });

  useEffect(() => {
    const receive = (event: Event) => {
      const detail = (event as CustomEvent<{ campaignId: string; result: TurnResponse; input: string }>).detail;
      if (detail.campaignId !== campaignId) return;
      dispatch({ type: 'CONFIRMED_INPUT', input: detail.input, turn: detail.result.turn_number ?? 0, timestamp: new Date().toISOString() });
      dispatch({ type: 'RESULT_RECEIVED', payload: { ...detail.result, choices: [] }, timestamp: new Date().toISOString(), entryCount: 0 });
    };
    window.addEventListener('edda:confirmed-action', receive);
    return () => window.removeEventListener('edda:confirmed-action', receive);
  }, [campaignId, dispatch]);

  // Effect 1: campaign change + history load
  useEffect(() => {
    if (!campaignId) return;
    let active = true;

    getSessionHistory(campaignId)
      .then((resp) => {
        if (!active) return;
        const restored: NarrativeEntry[] = [];
        for (const entry of resp.entries) {
          if (entry.input_type !== 'narrative' && entry.player_input) {
            restored.push({
              id: `history-player-${entry.turn_number}`,
              kind: 'player',
              text: entry.player_input,
              timestamp: entry.created_at,
              speaker: 'You',
            });
          }
          if (entry.llm_response) {
            restored.push({
              id: `history-gm-${entry.turn_number}`,
              kind: 'gm',
              text: entry.llm_response,
              timestamp: entry.created_at,
              speaker: 'Game Master',
              resolutionEvents: entry.resolution_events ?? [],
              choices: entry.choices?.map((choice, index) => ({ id: `history-choice-${entry.turn_number}-${index + 1}`, text: choice })),
            });
          }
        }
        dispatch({ type: 'HISTORY_LOADED', entries: restored });
      })
      .catch((err) => {
        if (!active) return;
        console.warn('Failed to load session history:', err);
        dispatch({ type: 'HISTORY_FAILED', error: 'Failed to load history' });
      });
    return () => { active = false; };
  }, [campaignId, dispatch]);

  // Effect 2: propagate socket-level error
  useEffect(() => {
    if (socketError !== null) {
      dispatch({ type: 'SOCKET_ERROR', error: socketError });
    } else {
      dispatch({ type: 'SOCKET_ERROR_CLEARED' });
    }
  }, [socketError, dispatch]);

  // Effect 3: connection lost
  useEffect(() => {
    if (connectionStatus === 'closed' || connectionStatus === 'error') {
      dispatch({ type: 'CONNECTION_LOST' });
    }
  }, [connectionStatus, dispatch]);

  const sendAction = useCallback(
    (input: string): boolean => {
      const trimmedInput = input.trim();
      if (!trimmedInput) {
        return false;
      }

      const accepted = sendSocketAction(trimmedInput);
      if (!accepted) {
        return false;
      }

      const timestamp = new Date().toISOString();

      dispatch({
        type: 'ACTION_SENT',
        input: trimmedInput,
        timestamp,
        entryCount: state.entries.length,
      });

      return true;
    },
    [dispatch, sendSocketAction, state.entries.length],
  );

  const streamingEntry = useMemo<NarrativeEntry | null>(() => {
    if (!state.pendingTimestamp || !isLoading) {
      return null;
    }

    return {
      id: `${state.pendingTimestamp}-streaming`,
      kind: 'gm',
      text: state.streamingChunks.join(''),
      timestamp: state.pendingTimestamp,
      speaker: 'Game Master',
      isStreaming: true,
    };
  }, [isLoading, state.pendingTimestamp, state.streamingChunks]);

  return {
    campaignId,
    connectionStatus,
    entries: state.entries,
    streamingEntry,
    latestResult: state.latestResult,
    suggestedChoices: state.suggestedChoices,
    currentStatus,
    combatActive,
    isLoading,
    error: state.error,
    sendAction,
  };
}

// Reduce synchronously so even a burst larger than the pending-queue limit is
// consumed before React renders. React observes snapshots, never transport frames.
function createNarrativeStore(campaignId: string | null) {
  let state = initialNarrativeState;
  const listeners = new Set<() => void>();
  return {
    campaignId,
    getSnapshot: () => state,
    subscribe: (listener: () => void) => { listeners.add(listener); return () => { listeners.delete(listener); }; },
    dispatch: (action: NarrativeAction) => {
      state = narrativeReducer(state, action);
      listeners.forEach(listener => listener());
    },
  };
}
