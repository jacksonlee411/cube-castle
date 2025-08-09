import React, { useState, useEffect } from 'react';
import { Box } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import type { SystemMetrics } from '../../shared/types/monitoring';
import { MonitoringService } from '../../shared/api/monitoring';
import { ServiceStatusGrid } from './components/ServiceStatusGrid';
import { MetricsCharts } from './components/MetricsCharts';
import { ControlPanel } from './components/ControlPanel';

export const MonitoringDashboard: React.FC = () => {
  const [metrics, setMetrics] = useState<SystemMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchMetrics = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await MonitoringService.getMetrics();
      setMetrics(data);
    } catch (err) {
      setError('获取监控数据失败');
      console.error('[MonitoringDashboard] Error fetching metrics:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = () => {
    fetchMetrics();
  };

  // 初始化加载数据
  useEffect(() => {
    fetchMetrics();
  }, []);

  // 自动刷新 (30秒)
  useEffect(() => {
    const interval = setInterval(() => {
      // 只有在页面可见且没有正在加载时才自动刷新
      if (document.visibilityState === 'visible' && !loading) {
        console.log('[MonitoringDashboard] 自动刷新监控数据');
        fetchMetrics();
      }
    }, 30000); // 30秒

    return () => clearInterval(interval);
  }, [loading]);

  if (error) {
    return (
      <Box padding="l" textAlign="center">
        <Box as="span" fontSize="48px">⚠️</Box>
        <Box marginTop="s">
          <Text variant="subtext" color="red">
            {error}
          </Text>
        </Box>
        <Box marginTop="m">
          <button onClick={handleRefresh}>重试</button>
        </Box>
      </Box>
    );
  }

  return (
    <Box>
      {/* 页面标题 */}
      <Box marginBottom="l">
        <Text as="h1" variant="heading" fontSize={28}>
          📊 系统监控
        </Text>
        <Text variant="subtext" marginTop="xs">
          实时监控系统状态和性能指标
        </Text>
      </Box>

      {/* 控制面板 */}
      <ControlPanel 
        lastUpdated={metrics?.lastUpdated}
        loading={loading}
        onRefresh={handleRefresh}
      />

      {loading && !metrics ? (
        <Box padding="xl" textAlign="center">
          <Box as="span" fontSize="48px">⏳</Box>
          <Box marginTop="s">
            <Text variant="subtext">正在加载监控数据...</Text>
          </Box>
        </Box>
      ) : (
        <>
          {/* 服务状态网格 */}
          <Box marginBottom="xl">
            <Text variant="subtext" fontWeight="bold" marginBottom="m">
              🖥️ 服务状态概览
            </Text>
            <ServiceStatusGrid services={metrics?.services} />
          </Box>

          {/* 性能指标图表 */}
          <Box>
            <Text variant="subtext" fontWeight="bold" marginBottom="m">
              📈 性能指标趋势
            </Text>
            <MetricsCharts data={metrics?.charts} />
          </Box>
        </>
      )}
    </Box>
  );
};