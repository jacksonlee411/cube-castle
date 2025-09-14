import React, { useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { authManager } from '../api/auth';

export const RequireAuth: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    if (!authManager.isAuthenticated()) {
      const redirect = encodeURIComponent(location.pathname + location.search);
      navigate(`/login?redirect=${redirect}`, { replace: true });
    }
  }, [navigate, location.pathname, location.search]);

  // 即使未认证也先渲染，401 会由全局拦截处理
  return <>{children}</>;
};

