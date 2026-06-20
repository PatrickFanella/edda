import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

import type { ActionRequest, WebSocketStatusPayload } from '../api/types';
import { buildCampaignWebSocketURL, decodeWebSocketEvent, type ParsedWebSocketEvent } from '../api/websocket';
export type { NarrativeChoice, TurnResponseWithChoices } from '../api/websocketParser';

export type ConnectionStatus = 'idle' | 'connecting' | 'open' | 'closed' | 'error';

export interface UseWebSocketResult {
  connectionStatus: ConnectionStatus;
  isConnected: boolean;
  isLoading: boolean;
  error: string | null;
  events: ParsedWebSocketEvent[];
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

export function useWebSocket(campaignId: string | null | undefined): UseWebSocketResult {
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>('idle');
  const [error, setError] = useState<string | null>(null);
  const [events, setEvents] = useState<ParsedWebSocketEvent[]>([]);
  const [isAwaitingResponse, setIsAwaitingResponse] = useState(false);
  const [currentStatus, setCurrentStatus] = useState<WebSocketStatusPayload | null>(null);
  const [combatActive, setCombatActive] = useState(false);
  const [reconnectNonce, setReconnectNonce] = useState(0);
  const socketRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnectAttemptsRef = useRef(0);

  useEffect(() => {
    const activeCampaignId = campaignId?.trim() ?? '';

    if (reconnectTimerRef.current) {
      clearTimeout(reconnectTimerRef.current);
      reconnectTimerRef.current = null;
    }

    if (!activeCampaignId) {
      setEvents([]);
      setError(null);
      setIsAwaitingResponse(false);
      setCurrentStatus(null);
      setConnectionStatus('idle');
      socketRef.current = null;
      return undefined;
    }

    const socket = new WebSocket(buildCampaignWebSocketURL(activeCampaignId));
    socketRef.current = socket;
    setConnectionStatus('connecting');

    let allowReconnect = true;
    const isCurrentSocket = () => socketRef.current === socket;

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
      setError(null);
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

      setEvents((current) => [...current, decodedEvent]);

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

    socket.addEventListener('open', handleOpen);
    socket.addEventListener('close', handleClose);
    socket.addEventListener('error', handleSocketError);
    socket.addEventListener('message', handleMessage);

    return () => {
      allowReconnect = false;

      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current);
        reconnectTimerRef.current = null;
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
      currentStatus,
      combatActive,
      sendAction,
    }),
    [combatActive, connectionStatus, currentStatus, error, events, isConnected, isLoading, sendAction],
  );
}
