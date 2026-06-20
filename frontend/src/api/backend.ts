import { getStoredToken } from './authSession';

const DEFAULT_API_BASE = '/api/v1';

export const API_BASE = normalizeBase(import.meta.env.VITE_API_BASE ?? DEFAULT_API_BASE);

export class APIError extends Error {
  readonly status: number;
  readonly body: unknown;

  constructor(status: number, message: string, body: unknown) {
    super(message);
    this.name = 'APIError';
    this.status = status;
    this.body = body;
  }
}

type APIRequestInit = Omit<RequestInit, 'body'> & {
  body?: BodyInit | unknown;
};

export type APIResponseMode = 'json' | 'void' | 'blob';

function normalizeBase(base: string): string {
  return base.endsWith('/') ? base.slice(0, -1) : base;
}

function resolvePath(path: string): string {
  return path.startsWith('/') ? `${API_BASE}${path}` : `${API_BASE}/${path}`;
}

function isBodyInit(body: unknown): body is BodyInit {
  return (
    typeof body === 'string' ||
    body instanceof FormData ||
    body instanceof URLSearchParams ||
    body instanceof Blob ||
    body instanceof ArrayBuffer ||
    ArrayBuffer.isView(body) ||
    body instanceof ReadableStream
  );
}

function buildHeaders(headers?: HeadersInit): Headers {
  const requestHeaders = new Headers(headers);
  requestHeaders.set('Accept', 'application/json');

  const token = getStoredToken();
  if (token && !requestHeaders.has('Authorization')) {
    requestHeaders.set('Authorization', `Bearer ${token}`);
  }

  return requestHeaders;
}

function buildRequestBody(body: APIRequestInit['body'], headers: Headers): BodyInit | undefined {
  if (body === undefined) return undefined;
  if (body === null || !isBodyInit(body)) {
    headers.set('Content-Type', 'application/json');
    return JSON.stringify(body);
  }
  return body;
}

async function readErrorBody(response: Response): Promise<unknown> {
  const raw = await response.text();
  return raw ? safeJsonParse(raw) : null;
}

function safeJsonParse(raw: string): unknown {
  try {
    return JSON.parse(raw) as unknown;
  } catch {
    return raw;
  }
}

function errorMessage(status: number, body: unknown): string {
  return body && typeof body === 'object' && 'error' in body && typeof body.error === 'string'
    ? body.error
    : `Request failed with status ${status}`;
}

export async function apiFetch<TResponse>(path: string, init: APIRequestInit = {}): Promise<TResponse> {
  return apiRequest<TResponse>(path, 'json', init);
}

export async function apiFetchVoid(path: string, init: APIRequestInit = {}): Promise<void> {
  await apiRequest(path, 'void', init);
}

export async function apiFetchBlob(path: string, init: APIRequestInit = {}): Promise<Blob> {
  return apiRequest<Blob>(path, 'blob', init);
}

async function apiRequest<TResponse>(path: string, mode: APIResponseMode, init: APIRequestInit): Promise<TResponse> {
  const { body, headers, ...rest } = init;
  const requestHeaders = buildHeaders(headers);
  const requestBody = buildRequestBody(body, requestHeaders);

  const response = await fetch(resolvePath(path), {
    ...rest,
    credentials: 'include',
    headers: requestHeaders,
    body: requestBody,
  });

  if (!response.ok) {
    const bodyValue = await readErrorBody(response);
    throw new APIError(response.status, errorMessage(response.status, bodyValue), bodyValue);
  }

  if (mode === 'void') return undefined as TResponse;
  if (mode === 'blob') return (await response.blob()) as TResponse;

  const raw = await response.text();
  return (raw ? (safeJsonParse(raw) as TResponse) : (null as TResponse));
}
