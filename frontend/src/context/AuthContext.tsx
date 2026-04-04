import { createContext, useContext, useMemo } from 'react';
import type { PropsWithChildren } from 'react';

const DEFAULT_USER_ID = '00000000-0000-0000-0000-000000000001';
const DEFAULT_USERNAME = 'Adventurer';

export interface AuthUser {
  readonly id: string;
  readonly username: string;
}

export interface AuthContextValue {
  readonly user: AuthUser;
  readonly isAuthenticated: true;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: PropsWithChildren) {
  const value = useMemo<AuthContextValue>(
    () => ({
      user: { id: DEFAULT_USER_ID, username: DEFAULT_USERNAME },
      isAuthenticated: true,
    }),
    [],
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
