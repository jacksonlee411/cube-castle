import { logger } from '@/shared/utils/logger';
import React, { useCallback, useEffect, useMemo } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { authManager } from '../api/auth';
import { authEvents, AUTH_UNAUTHORIZED } from './events';
import { AuthContext, type AuthContextValue } from './context';
import { useScopes } from '@/shared/hooks/useScopes';

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const { scopes, has } = useScopes();

  const permissions = useMemo(() => Array.from(scopes).sort(), [scopes]);

  const redirectToLogin = useCallback(() => {
    const redirect = encodeURIComponent(location.pathname + location.search);
    navigate(`/login?redirect=${redirect}`, { replace: true });
  }, [navigate, location.pathname, location.search]);

  useEffect(() => {
    const handler = () => {
      try { authManager.clearAuth(); } catch (error) {
        logger.warn('Failed to clear auth:', error);
      }
      redirectToLogin();
    };
    authEvents.addEventListener(AUTH_UNAUTHORIZED, handler);
    return () => authEvents.removeEventListener(AUTH_UNAUTHORIZED, handler);
  }, [redirectToLogin]);

  const isAuthenticated = useCallback(() => authManager.isAuthenticated(), []);

  const hasPermission = useCallback((permission: string) => {
    if (!permission) {
      return true;
    }
    return has(permission);
  }, [has]);

  const logout = useCallback(() => {
    authManager.clearAuth();
    redirectToLogin();
  }, [redirectToLogin]);

  const value: AuthContextValue = useMemo(() => ({
    isAuthenticated,
    logout,
    userPermissions: permissions,
    hasPermission,
  }), [isAuthenticated, logout, permissions, hasPermission]);

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};

// Export hook in separate file to avoid react-refresh issues
// This component only exports AuthProvider
