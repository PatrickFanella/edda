import { useCallback, useEffect, useMemo, useReducer } from 'react';

import { getSessionHistory } from '../api/campaigns';
import type { WebSocketStatusPayload } from '../api/types';
import { useCampaign } from './useCampaign';
import { useWebSocket, type ConnectionStatus, type NarrativeChoice, type TurnResponseWithChoices } from './useWebSocket';
import {
  narrativeReducer,
  initialNarrativeState,
  type NarrativeEntry,
  type NarrativeEntryKind,
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
  const { connectionStatus, error: socketError, events, isLoading, currentStatus, combatActive, sendAction: sendSocketAction } = useWebSocket(campaignId);
  const [state, dispatch] = useReducer(narrativeReducer, initialNarrativeState);

  // Effect 1: campaign change + history load
  useEffect(() => {
    dispatch({ type: 'RESET' });

    if (!campaignId) return;

    getSessionHistory(campaignId)
      .then((resp) => {
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
            });
          }
        }
        dispatch({ type: 'HISTORY_LOADED', entries: restored });
      })
      .catch((err) => {
        console.warn('Failed to load session history:', err);
        dispatch({ type: 'HISTORY_FAILED', error: 'Failed to load history' });
      });
  }, [campaignId]);

  // Effect 2: propagate socket-level error
  useEffect(() => {
    if (socketError !== null) {
      dispatch({ type: 'SOCKET_ERROR', error: socketError });
    } else {
      dispatch({ type: 'SOCKET_ERROR_CLEARED' });
    }
  }, [socketError]);

  // Effect 3: connection lost
  useEffect(() => {
    if (connectionStatus === 'closed' || connectionStatus === 'error') {
      dispatch({ type: 'CONNECTION_LOST' });
    }
  }, [connectionStatus]);

  // Effect 4: process new events
  useEffect(() => {
    const newEvents = events.slice(state.processedEventCount);
    if (newEvents.length === 0) return;

    for (const event of newEvents) {
      switch (event.kind) {
        case 'chunk':
          dispatch({ type: 'CHUNK_RECEIVED', text: event.payload.text, timestamp: event.envelope.timestamp });
          break;
        case 'result':
          dispatch({ type: 'RESULT_RECEIVED', payload: event.payload, timestamp: event.envelope.timestamp, entryCount: state.entries.length });
          break;
        case 'error':
          dispatch({ type: 'ERROR_RECEIVED', error: event.payload.error, timestamp: event.envelope.timestamp, entryCount: state.entries.length });
          break;
        case 'status':
          break;
      }
    }
    dispatch({ type: 'EVENTS_PROCESSED', count: events.length });
  }, [events, state.processedEventCount, state.entries.length]);

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
    [sendSocketAction, state.entries.length],
  );

  const streamingEntry = useMemo<NarrativeEntry | null>(() => {
    if (!state.pendingTimestamp || !isLoading) {
      return null;
    }

    return {
      id: `${state.pendingTimestamp}-streaming`,
      kind: 'gm',
      text: state.streamingChunks.join('') || 'Game Master is thinking\u2026',
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
