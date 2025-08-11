import React from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import { 
  chartIcon, 
  activityStreamIcon
} from '@workday/canvas-system-icons-web';
import { LineChart, Line, XAxis, YAxis, ResponsiveContainer, CartesianGrid, Tooltip } from 'recharts';
import type { ChartData } from '../../../shared/types/monitoring';

interface MetricsChartsProps {
  data?: ChartData | undefined;
}

export const MetricsCharts: React.FC<MetricsChartsProps> = ({ data }) => {
  if (!data) {
    return (
      <Box padding="l" textAlign="center">
        <SystemIcon icon={chartIcon} size={48} />
        <Box marginTop="s">暂无图表数据</Box>
      </Box>
    );
  }

  return (
    <Box>
      {/* 响应时间图表 */}
      <Box marginBottom="xl">
        <Flex alignItems="center" style={{gap: '8px'}} marginBottom="m">
          <SystemIcon icon={chartIcon} size={16} />
          <Text fontWeight="bold">
            平均响应时间 (毫秒)
          </Text>
        </Flex>
        <Box height="200px" width="100%">
          <ResponsiveContainer>
            <LineChart data={data.responseTime}>
              <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
              <XAxis 
                dataKey="timestamp" 
                fontSize={12}
                tick={{ fill: '#666' }}
              />
              <YAxis 
                fontSize={12}
                tick={{ fill: '#666' }}
              />
              <Tooltip 
                contentStyle={{
                  backgroundColor: '#fff',
                  border: '1px solid #e0e0e0',
                  borderRadius: '4px'
                }}
              />
              <Line 
                type="monotone" 
                dataKey="value" 
                stroke="#3498db" 
                strokeWidth={2}
                dot={{ r: 4, fill: '#3498db' }}
                activeDot={{ r: 6 }}
              />
            </LineChart>
          </ResponsiveContainer>
        </Box>
      </Box>

      {/* 错误率图表 */}
      <Box marginBottom="xl">
        <Text fontWeight="bold" marginBottom="m">
          🚨 错误率 (%)
        </Text>
        <Box height="200px" width="100%">
          <ResponsiveContainer>
            <LineChart data={data.errorRate}>
              <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
              <XAxis 
                dataKey="timestamp" 
                fontSize={12}
                tick={{ fill: '#666' }}
              />
              <YAxis 
                fontSize={12}
                tick={{ fill: '#666' }}
              />
              <Tooltip 
                contentStyle={{
                  backgroundColor: '#fff',
                  border: '1px solid #e0e0e0',
                  borderRadius: '4px'
                }}
              />
              <Line 
                type="monotone" 
                dataKey="value" 
                stroke="#e74c3c" 
                strokeWidth={2}
                dot={{ r: 4, fill: '#e74c3c' }}
                activeDot={{ r: 6 }}
              />
            </LineChart>
          </ResponsiveContainer>
        </Box>
      </Box>

      {/* 请求量图表 */}
      <Box>
        <Flex alignItems="center" style={{gap: '8px'}} marginBottom="m">
          <SystemIcon icon={activityStreamIcon} size={16} />
          <Text fontWeight="bold">
            请求量 (次/分钟)
          </Text>
        </Flex>
        <Box height="200px" width="100%">
          <ResponsiveContainer>
            <LineChart data={data.requestVolume}>
              <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
              <XAxis 
                dataKey="timestamp" 
                fontSize={12}
                tick={{ fill: '#666' }}
              />
              <YAxis 
                fontSize={12}
                tick={{ fill: '#666' }}
              />
              <Tooltip 
                contentStyle={{
                  backgroundColor: '#fff',
                  border: '1px solid #e0e0e0',
                  borderRadius: '4px'
                }}
              />
              <Line 
                type="monotone" 
                dataKey="value" 
                stroke="#27ae60" 
                strokeWidth={2}
                dot={{ r: 4, fill: '#27ae60' }}
                activeDot={{ r: 6 }}
              />
            </LineChart>
          </ResponsiveContainer>
        </Box>
      </Box>
    </Box>
  );
};