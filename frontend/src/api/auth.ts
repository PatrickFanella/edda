import { apiFetch, apiFetchVoid } from './backend';
import { buildAuthHeaders, clearStoredSession, getStoredToken, type AuthSession, type AuthUser } from './authSession';

export function register(name: string, email: string, password: string): Promise<AuthSession> {
  return apiFetch<AuthSession>('/auth/register', {
    method: 'POST',
    body: { name, email, password },
  });
}

export function login(email: string, password: string): Promise<AuthSession> {
  return apiFetch<AuthSession>('/auth/login', {
    method: 'POST',
    body: { email, password },
  });
}

export async function getMe(token: string): Promise<{ user: AuthUser }> {
  return apiFetch<{ user: AuthUser }>('/auth/me', {
    method: 'GET',
    credentials: 'include',
    headers: (() => {
      const headers = buildAuthHeaders(token);
      return headers;
    })(),
  });
}

export async function logout(): Promise<void> {
  try {
    await apiFetchVoid('/auth/logout', {
      method: 'POST',
      headers: buildAuthHeaders(getStoredToken()),
    });
  } finally {
    clearStoredSession();
  }
}
