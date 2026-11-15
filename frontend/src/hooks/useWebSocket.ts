import { useCallback, useEffect, useEffectEvent, useMemo, useRef, useState } from 'react';

import type { ActionRequest, WebSocketStatusPayload } from '../api/types';
import { buildCampaignWebSocketURL, decodeWebSocketEvent, type ParsedWebSocketEvent } from '../api/websocket';
export type { NarrativeChoice, TurnResponseWithChoices } from '../api/websocketParser';

export type ConnectionStatus = 'idle' | 'connecting' | 'open' | 'closed' | 'error';

export type SequencedEvent = ParsedWebSocketEvent & { sequence: number };
// 256 frames is over 23 turns at the #43 baseline (11 frames/turn).
export const MAX_PENDING_EVENTS = 256;

export interface UseWebSocketResult {
  connectionStatus: ConnectionStatus;
  isConnected: boolean;
  isLoading: boolean;
  error: string | null;
  events: SequencedEvent[];
  acknowledge: (sequence: number) => void;
  currentStatus: WebSocketStatusPayload | null;
  combatActive: boolean;
  sendAction: (input: string) => boolean;
}

type OutboundActionEnvelope = {
  type: 'action';
  payload: ActionRequest;
};

const SOCKET_UNAVAILABLE_MESSAGE = 'Live connection unavailable.';
const MALFORMED_MESSAGE_ERROR = 'Received a malformed live update from the server.';
const RECONNECT_BASE_DELAY_MS = 1000;
const RECONNECT_MAX_DELAY_MS = 30000;
const RECONNECT_MAX_ATTEMPTS = 10;

export function useWebSocket(campaignId: string | null | undefined, consume?: (event: SequencedEvent) => void): UseWebSocketResult {
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>('idle');
  const [error, setError] = useState<string | null>(null);
  const [events, setEvents] = useState<SequencedEvent[]>([]);
  const [isAwaitingResponse, setIsAwaitingResponse] = useState(false);
  const [currentStatus, setCurrentStatus] = useState<WebSocketStatusPayload | null>(null);
  const [combatActive, setCombatActive] = useState(false);
  const [reconnectNonce, setReconnectNonce] = useState(0);
  const [stateCampaignId, setStateCampaignId] = useState(campaignId);
  // Reset before committing a render for another campaign. Reconnects keep
  // their event cursor; a campaign switch starts a separate live session.
  if (stateCampaignId !== campaignId) {
    setStateCampaignId(campaignId);
    setEvents([]);
    setError(null);
    setIsAwaitingResponse(false);
    setCurrentStatus(null);
    setCombatActive(false);
    setConnectionStatus(campaignId ? 'connecting' : 'idle');
  }
  const cursor = useRef({ campaignId, next: 0, pending: [] as SequencedEvent[], blocked: false });
  const consumeEvent = useEffectEvent((event: SequencedEvent) => {
    if (!consume) return false;
    consume(event); // Synchronous reduction acknowledges the frame before releasing it.
    return true;
  });
  const acknowledge = useCallback((sequence: number) => {
    if (cursor.current.campaignId !== campaignId) return;
    cursor.current.pending = cursor.current.pending.filter(event => event.sequence > sequence);
    setEvents(cursor.current.pending);
  }, [campaignId]);
  const socketRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnectAttemptsRef = useRef(0);

  useEffect(() => {
    if (cursor.current.campaignId !== campaignId) {
      cursor.current = { campaignId, next: 0, pending: [], blocked: false };
    }
    const activeCampaignId = campaignId?.trim() ?? '';

    if (reconnectTimerRef.current) {
      clearTimeout(reconnectTimerRef.current);
      reconnectTimerRef.current = null;
    }

    if (!activeCampaignId) {
      socketRef.current = null;
      return undefined;
    }

    let allowReconnect = true;
    let socket: WebSocket | null = null;
    const isCurrentSocket = () => socket !== null && socketRef.current === socket;

    const scheduleReconnect = (nextStatus: ConnectionStatus, nextError: string | null) => {
      if (!isCurrentSocket()) {
        return;
      }

      setConnectionStatus(nextStatus);
      setError(nextError);
      setIsAwaitingResponse(false);

      if (!allowReconnect) {
        return;
      }

      reconnectAttemptsRef.current += 1;
      if (reconnectAttemptsRef.current > RECONNECT_MAX_ATTEMPTS) {
        setError('Connection lost. Refresh the page to reconnect.');
        return;
      }

      const delay = Math.min(
        RECONNECT_BASE_DELAY_MS * Math.pow(2, reconnectAttemptsRef.current - 1),
        RECONNECT_MAX_DELAY_MS,
      );

      reconnectTimerRef.current = setTimeout(() => {
        if (!allowReconnect) {
          return;
        }

        setReconnectNonce((current) => current + 1);
      }, delay);
    };

    const handleOpen = () => {
      if (!isCurrentSocket()) {
        return;
      }

      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current);
        reconnectTimerRef.current = null;
      }

      setConnectionStatus('open');
      if (!cursor.current.blocked) setError(null);
      reconnectAttemptsRef.current = 0;
    };

    const handleClose = () => {
      scheduleReconnect('closed', null);
    };

    const handleSocketError = () => {
      scheduleReconnect('error', SOCKET_UNAVAILABLE_MESSAGE);
    };

    const handleMessage = (event: MessageEvent<string>) => {
      if (!isCurrentSocket()) {
        return;
      }

      const decodedEvent = decodeWebSocketEvent(event.data);
      if (decodedEvent === null) {
        setError(MALFORMED_MESSAGE_ERROR);
        setIsAwaitingResponse(false);
        return;
      }

      if (cursor.current.blocked) return;
      const sequenced = { ...decodedEvent, sequence: ++cursor.current.next };
      if (!consumeEvent(sequenced) && decodedEvent.kind !== 'status') {
        if (cursor.current.pending.length >= MAX_PENDING_EVENTS) {
          // Never silently slice results/errors. Stop accepting actions and require
          // an explicit refresh to reload committed history after consumer failure.
          cursor.current.blocked = true;
          allowReconnect = false;
          socket?.close(1008, 'consumer_overflow');
          setError('Live updates paused: refresh to recover committed history. Do not resend your last action.');
          setIsAwaitingResponse(false);
          return;
        }
        cursor.current.pending = [...cursor.current.pending, sequenced];
        setEvents(cursor.current.pending);
      }

      if (decodedEvent.kind === 'status') {
        setCurrentStatus(decodedEvent.payload);
        if (decodedEvent.payload.stage === 'combat_start') {
          setCombatActive(true);
        } else if (decodedEvent.payload.stage === 'combat_end') {
          setCombatActive(false);
        }
        return;
      }

      if (decodedEvent.kind === 'chunk') {
        setError(null);
        return;
      }

      if (decodedEvent.kind === 'result') {
        setError(null);
        setCurrentStatus(null);
        setIsAwaitingResponse(false);
        if ('combat_active' in decodedEvent.payload) {
          setCombatActive(Boolean(decodedEvent.payload.combat_active));
        }
        return;
      }

      setError(decodedEvent.payload.error);
      setCurrentStatus(null);
      setIsAwaitingResponse(false);
    };

    const connectionTimer = setTimeout(() => {
      if (!allowReconnect) {
        return;
      }

      setConnectionStatus('connecting');
      socket = new WebSocket(buildCampaignWebSocketURL(activeCampaignId));
      socketRef.current = socket;

      socket.addEventListener('open', handleOpen);
      socket.addEventListener('close', handleClose);
      socket.addEventListener('error', handleSocketError);
      socket.addEventListener('message', handleMessage);
    }, 0);

    return () => {
      allowReconnect = false;
      clearTimeout(connectionTimer);

      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current);
        reconnectTimerRef.current = null;
      }

      if (socket === null) {
        return;
      }

      socket.removeEventListener('open', handleOpen);
      socket.removeEventListener('close', handleClose);
      socket.removeEventListener('error', handleSocketError);
      socket.removeEventListener('message', handleMessage);

      if (socketRef.current === socket) {
        socketRef.current = null;
      }

      if (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING) {
        socket.close(1000, 'campaign_changed');
      }
    };
  }, [campaignId, reconnectNonce]);

  const sendAction = useCallback(
    (input: string): boolean => {
      const trimmedInput = input.trim();
      if (!trimmedInput) {
        return false;
      }

      if (!(campaignId?.trim())) {
        setError(SOCKET_UNAVAILABLE_MESSAGE);
        return false;
      }

      const socket = socketRef.current;
      if (cursor.current.blocked) return false;
      if (socket === null || socket.readyState !== WebSocket.OPEN) {
        setError(SOCKET_UNAVAILABLE_MESSAGE);
        return false;
      }

      const message: OutboundActionEnvelope = {
        type: 'action',
        payload: { input: trimmedInput },
      };

      try {
        socket.send(JSON.stringify(message));
        setError(null);
        setIsAwaitingResponse(true);
        return true;
      } catch {
        setError(SOCKET_UNAVAILABLE_MESSAGE);
        setIsAwaitingResponse(false);
        return false;
      }
    },
    [campaignId],
  );

  const isConnected = connectionStatus === 'open';
  const isLoading = connectionStatus === 'connecting' || isAwaitingResponse;

  return useMemo(
    () => ({
      connectionStatus,
      isConnected,
      isLoading,
      error,
      events,
      acknowledge,
      currentStatus,
      combatActive,
      sendAction,
    }),
    [acknowledge, combatActive, connectionStatus, currentStatus, error, events, isConnected, isLoading, sendAction],
  );
}
