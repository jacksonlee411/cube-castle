import { logger } from '@/shared/utils/logger';
import React, { useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { authManager } from '../api/auth';
import { authEvents, AUTH_UNAUTHORIZED } from './events';
import { AuthContext, type AuthContextValue } from './context';

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    const handler = () => {
      try { authManager.clearAuth(); } catch (error) {
        logger.warn('Failed to clear auth:', error);
      }
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

// Export hook in separate file to avoid react-refresh issues
// This component only exports AuthProvider
