import React, { useCallback, useMemo, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Card } from '@workday/canvas-kit-react/card';
import { Heading, Text } from '@workday/canvas-kit-react/text';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { LoadingDots } from '@workday/canvas-kit-react/loading-dots';
import { borderRadius, colors } from '@workday/canvas-kit-react/tokens';
import { authManager } from '../shared/api/auth';
import { env } from '../shared/config/environment';

export const LoginPage: React.FC = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const redirect = useMemo(() => {
    const params = new URLSearchParams(location.search);
    return params.get('redirect') || '/';
  }, [location.search]);

  const handleDevReauth = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      await authManager.forceRefresh();
      // 使用硬跳转避免路由守卫与认证状态的竞态
      window.location.replace(decodeURIComponent(redirect));
    } catch (e) {
      const msg = e instanceof Error ? e.message : '开发令牌获取失败';
      setError(msg);
    } finally {
      setLoading(false);
    }
  }, [redirect]);

  const handleEnterpriseLogin = useCallback(() => {
    const target = `/auth/login?redirect=${encodeURIComponent(redirect)}`;
    window.location.assign(target);
  }, [redirect]);

  return (
    <Flex justifyContent="center" alignItems="center" style={{ minHeight: '80vh' }}>
      <Card padding="l" style={{ minWidth: 420, borderRadius: borderRadius.l }}>
        <Heading size="large">登录</Heading>
        <Text typeLevel="body.small" color="hint" style={{ marginTop: 8 }}>
          会话已过期或尚未认证。请重新获取开发令牌继续使用。
        </Text>

        <Box marginTop="m" padding="s" style={{ background: colors.soap100, borderRadius: borderRadius.m }}>
          <Text typeLevel="subtext.small">
            当前租户：<b>{env.defaultTenantId}</b>
          </Text>
        </Box>

        {error && (
          <Box marginTop="s" padding="s" style={{ background: colors.cinnamon100, border: `1px solid ${colors.cinnamon600}`, borderRadius: borderRadius.s }}>
            <Text color={colors.cinnamon600}>⚠️ {error}</Text>
          </Box>
        )}

        <Flex gap="s" marginTop="l">
          <PrimaryButton onClick={handleDevReauth} disabled={loading}>
            {loading ? (<><LoadingDots /> 获取中...</>) : '重新获取开发令牌并继续'}
          </PrimaryButton>
          <SecondaryButton onClick={() => navigate('/', { replace: true })} disabled={loading}>
            返回首页
          </SecondaryButton>
        </Flex>

        {env.auth.mode === 'oidc' && (
          <Box marginTop="s">
            <PrimaryButton onClick={handleEnterpriseLogin}>
              前往企业登录（生产）
            </PrimaryButton>
          </Box>
        )}

        <Text typeLevel="subtext.small" color="hint" style={{ marginTop: 12 }}>
          成功后将跳转至：{decodeURIComponent(redirect)}
        </Text>
      </Card>
    </Flex>
  );
};

export default LoginPage;
