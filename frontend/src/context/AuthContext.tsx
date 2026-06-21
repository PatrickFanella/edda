import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react';
import type { PropsWithChildren } from 'react';

import { getMe, logout as logoutRequest } from '../api/auth';
import { clearStoredSession, getStoredToken, getStoredUser, setStoredSession, type AuthUser } from '../api/authSession';

export interface AuthContextValue {
  readonly user: AuthUser | null;
  readonly token: string | null;
  readonly isAuthenticated: boolean;
  readonly isLoading: boolean;
  readonly setSession: (token: string, user: AuthUser) => void;
  readonly logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: PropsWithChildren) {
  const [token, setToken] = useState<string | null>(getStoredToken);
  const [user, setUser] = useState<AuthUser | null>(getStoredUser);
  const [isLoading, setIsLoading] = useState(true);

  const didValidate = useRef(false);

  useEffect(() => {
    if (didValidate.current) return;
    didValidate.current = true;

    if (!token) {
      setIsLoading(false);
      return;
    }

    // Validate token against the server on initial load.
    getMe(token)
      .then((res) => {
        setUser(res.user);
        setStoredSession(token, res.user);
      })
      .catch(() => {
        // Token expired or invalid — clear everything.
        clearStoredSession();
        setToken(null);
        setUser(null);
      })
      .finally(() => setIsLoading(false));
  }, [token]);

  const setSession = useCallback((newToken: string, newUser: AuthUser) => {
    setStoredSession(newToken, newUser);
    setToken(newToken);
    setUser(newUser);
  }, []);

  const logout = useCallback(async () => {
    setToken(null);
    setUser(null);

    try {
      await logoutRequest();
    } catch {
      // Local auth state must still clear even if the network request fails.
      clearStoredSession();
    }
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      token,
      isAuthenticated: !!token && !!user,
      isLoading,
      setSession,
      logout,
    }),
    [user, token, isLoading, setSession, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return ctx;
}
