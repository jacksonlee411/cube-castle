import React, { useState, useEffect } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import { 
  chartIcon, 
  exclamationTriangleIcon, 
  clockIcon,
  homeIcon,
  activityStreamIcon
} from '@workday/canvas-system-icons-web';
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
        <SystemIcon icon={exclamationTriangleIcon} size={48} color="red" />
        <Box marginTop="s">
          <Text as="div" color="red">
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
        <Flex alignItems="center" style={{gap: '8px'}}>
          <SystemIcon icon={chartIcon} size={24} />
          <Text as="h1">
            系统监控
          </Text>
        </Flex>
        <Text as="p" marginTop="xs">
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
          <SystemIcon icon={clockIcon} size={48} />
          <Box marginTop="s">
            <Text as="div">正在加载监控数据...</Text>
          </Box>
        </Box>
      ) : (
        <>
          {/* 服务状态网格 */}
          <Box marginBottom="xl">
            <Flex alignItems="center" style={{gap: '8px'}} marginBottom="m">
              <SystemIcon icon={homeIcon} size={20} />
              <Text as="h3" fontWeight="bold">
                服务状态概览
              </Text>
            </Flex>
            <ServiceStatusGrid services={metrics?.services} />
          </Box>

          {/* 性能指标图表 */}
          <Box>
            <Flex alignItems="center" style={{gap: '8px'}} marginBottom="m">
              <SystemIcon icon={activityStreamIcon} size={20} />
              <Text as="h3" fontWeight="bold">
                性能指标趋势
              </Text>
            </Flex>
            <MetricsCharts data={metrics?.charts} />
          </Box>
        </>
      )}
    </Box>
  );
};