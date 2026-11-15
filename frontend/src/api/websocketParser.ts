import type {
  StateChange,
  ResolutionEvent,
  TurnResponse,
  WebSocketChunkPayload,
  WebSocketErrorPayload,
  WebSocketMessageEnvelope,
  WebSocketStatusPayload,
} from './types';

export interface NarrativeChoice {
  id: string;
  text: string;
}

export interface TurnResponseWithChoices extends TurnResponse {
  choices: NarrativeChoice[];
}

export type ParsedWebSocketEvent =
  | {
      kind: 'chunk';
      envelope: WebSocketMessageEnvelope<WebSocketChunkPayload>;
      payload: WebSocketChunkPayload;
    }
  | {
      kind: 'result';
      envelope: WebSocketMessageEnvelope<TurnResponseWithChoices>;
      payload: TurnResponseWithChoices;
    }
  | {
      kind: 'error';
      envelope: WebSocketMessageEnvelope<WebSocketErrorPayload>;
      payload: WebSocketErrorPayload;
    }
  | {
      kind: 'status';
      envelope: WebSocketMessageEnvelope<WebSocketStatusPayload>;
      payload: WebSocketStatusPayload;
    };

interface WebSocketEnvelope {
  type: string;
  payload: unknown;
  timestamp: string;
}

export function decodeWebSocketEvent(raw: string): ParsedWebSocketEvent | null {
  let parsed: unknown;
  try {
    parsed = JSON.parse(raw) as unknown;
  } catch {
    return null;
  }

  if (!isEnvelope(parsed)) {
    return null;
  }

  switch (parsed.type) {
    case 'chunk':
      if (!isChunkPayload(parsed.payload)) {
        return null;
      }

      return {
        kind: 'chunk',
        envelope: {
          ...parsed,
          payload: parsed.payload,
        },
        payload: parsed.payload,
      };
    case 'result': {
      const resultPayload = normalizeTurnResponse(parsed.payload);
      if (resultPayload === null) {
        return null;
      }

      return {
        kind: 'result',
        envelope: {
          ...parsed,
          payload: resultPayload,
        },
        payload: resultPayload,
      };
    }
    case 'status':
      if (!isStatusPayload(parsed.payload)) {
        return null;
      }

      return {
        kind: 'status',
        envelope: {
          ...parsed,
          payload: parsed.payload,
        },
        payload: parsed.payload,
      };
    case 'error':
      if (!isErrorPayload(parsed.payload)) {
        return null;
      }

      return {
        kind: 'error',
        envelope: {
          ...parsed,
          payload: parsed.payload,
        },
        payload: parsed.payload,
      };
    default:
      return null;
  }
}

function isEnvelope(value: unknown): value is WebSocketEnvelope {
  return isRecord(value) && typeof value.type === 'string' && 'payload' in value && typeof value.timestamp === 'string';
}

function isChunkPayload(value: unknown): value is WebSocketChunkPayload {
  return isRecord(value) && typeof value.text === 'string';
}

function isStatusPayload(value: unknown): value is WebSocketStatusPayload {
  return isRecord(value) && typeof value.stage === 'string' && typeof value.description === 'string';
}

function isErrorPayload(value: unknown): value is WebSocketErrorPayload {
  return isRecord(value) && typeof value.error === 'string' && (value.action_id === undefined || typeof value.action_id === 'string');
}

function normalizeTurnResponse(value: unknown): TurnResponseWithChoices | null {
  if (!isRecord(value) || typeof value.narrative !== 'string' || !Array.isArray(value.state_changes)) {
    return null;
  }

  const stateChanges = value.state_changes;
  if (!stateChanges.every(isStateChange)) {
    return null;
  }

  const resolutionEvents = Array.isArray(value.resolution_events) ? value.resolution_events : [];
  if (!resolutionEvents.every(isResolutionEvent)) {
    return null;
  }

  return {
    action_id: typeof value.action_id === 'string' ? value.action_id : undefined,
    recovered: value.recovered === true,
    turn_number: typeof value.turn_number === 'number' ? value.turn_number : undefined,
    narrative: value.narrative,
    state_changes: stateChanges,
    resolution_events: resolutionEvents,
    combat_active: typeof value.combat_active === 'boolean' ? value.combat_active : false,
    choices: normalizeChoices(value.choices),
  };
}

function normalizeChoices(value: unknown): NarrativeChoice[] {
  if (!Array.isArray(value)) {
    return [];
  }

  return value.filter(isNarrativeChoice);
}

function isNarrativeChoice(value: unknown): value is NarrativeChoice {
  return isRecord(value) && typeof value.id === 'string' && typeof value.text === 'string';
}

function isStateChange(value: unknown): value is StateChange {
  return (
    isRecord(value) &&
    typeof value.entity_type === 'string' &&
    typeof value.entity_id === 'string' &&
    typeof value.change_type === 'string' &&
    isRecord(value.details)
  );
}

function isResolutionEvent(value: unknown): value is ResolutionEvent {
  return (
    isRecord(value) &&
    typeof value.type === 'string' &&
    typeof value.label === 'string' &&
    typeof value.outcome === 'string' &&
    isRecord(value.details)
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}
