import React, { createContext, useContext, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { authManager } from '../api/auth';
import { authEvents, AUTH_UNAUTHORIZED } from './events';

interface AuthContextValue {
  isAuthenticated: () => boolean;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    const handler = () => {
      try { authManager.clearAuth(); } catch {}
      const redirect = encodeURIComponent(location.pathname + location.search);
      navigate(`/login?redirect=${redirect}`, { replace: true });
    };
    authEvents.addEventListener(AUTH_UNAUTHORIZED, handler);
    return () => authEvents.removeEventListener(AUTH_UNAUTHORIZED, handler);
  }, [navigate, location.pathname, location.search]);

  const value: AuthContextValue = {
    isAuthenticated: () => authManager.isAuthenticated(),
    logout: () => {
      authManager.clearAuth();
      const redirect = encodeURIComponent(location.pathname + location.search);
      navigate(`/login?redirect=${redirect}`, { replace: true });
    }
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
};

