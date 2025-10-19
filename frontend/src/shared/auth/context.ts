import { createContext } from 'react';

export type AuthContextValue = {
  isAuthenticated: () => boolean;
  logout: () => void;
  userPermissions: string[];
  hasPermission: (permission: string) => boolean;
};

export const AuthContext = createContext<AuthContextValue | null>(null);
