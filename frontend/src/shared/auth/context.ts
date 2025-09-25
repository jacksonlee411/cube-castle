import { createContext } from 'react';

export interface AuthContextValue {
  isAuthenticated: () => boolean;
  logout: () => void;
}

export const AuthContext = createContext<AuthContextValue | null>(null);

