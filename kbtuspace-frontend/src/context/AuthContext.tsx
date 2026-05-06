import { createContext, useContext, useState, useEffect, type ReactNode } from 'react';
import type { User, AuthTokenPayload } from '../types';
import { getProfile } from '../api/auth';

interface AuthContextValue { user: User | null; token: string | null; isLoading: boolean; login: (token: string) => Promise<void>; logout: () => void; }
const AuthContext = createContext<AuthContextValue | null>(null);

function decodeToken(token: string): AuthTokenPayload | null {
  try { return JSON.parse(atob(token.split('.')[1])); } catch { return null; }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(() => localStorage.getItem('token'));
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const stored = localStorage.getItem('token');
    if (!stored) { setIsLoading(false); return; }
    const payload = decodeToken(stored);
    if (!payload || payload.exp * 1000 < Date.now()) { localStorage.removeItem('token'); setToken(null); setIsLoading(false); return; }
    getProfile().then(setUser).catch(() => { localStorage.removeItem('token'); setToken(null); }).finally(() => setIsLoading(false));
  }, []);

  const login = async (newToken: string) => { localStorage.setItem('token', newToken); setToken(newToken); const profile = await getProfile(); setUser(profile); };
  const logout = () => { localStorage.removeItem('token'); setToken(null); setUser(null); };

  return <AuthContext.Provider value={{ user, token, isLoading, login, logout }}>{children}</AuthContext.Provider>;
}

export function useAuth() { const ctx = useContext(AuthContext); if (!ctx) throw new Error('useAuth must be used within AuthProvider'); return ctx; }
