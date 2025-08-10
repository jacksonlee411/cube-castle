import React from 'react';
import { Card } from '@workday/canvas-kit-react/card';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import type { ServiceStatus } from '../../../shared/types/monitoring';

interface ServiceCardProps {
  service: ServiceStatus;
}

export const ServiceCard: React.FC<ServiceCardProps> = ({ service }) => {
  const getStatusColor = (status: ServiceStatus['status']) => {
    switch (status) {
      case 'online': return 'green';
      case 'warning': return 'orange'; 
      case 'offline': return 'red';
      default: return 'gray';
    }
  };

  const getStatusIcon = (status: ServiceStatus['status']) => {
    switch (status) {
      case 'online': return '🟢';
      case 'warning': return '🟡';
      case 'offline': return '🔴';
      default: return '⚪';
    }
  };

  return (
    <Card padding="m" width="100%">
      <Flex alignItems="flex-start" justifyContent="space-between">
        <Box flex={1}>
          <Flex alignItems="center" gap="xs" marginBottom="xs">
            <Text fontSize={14}>{getStatusIcon(service.status)}</Text>
            <Text 
              fontWeight="bold"
              color={getStatusColor(service.status)}
            >
              {service.name}
            </Text>
          </Flex>
          
          <Text variant="hint" marginBottom="xxs">
            端口: {service.port}
          </Text>
          
          <Flex gap="m" marginTop="xs">
            <Box>
              <Text variant="hint" fontSize={12}>响应时间</Text>
              <Text fontSize={18} fontWeight="bold" color="blue">
                {service.responseTime}
              </Text>
            </Box>
            <Box>
              <Text variant="hint" fontSize={12}>请求/分钟</Text>
              <Text fontSize={18} fontWeight="bold" color="green">
                {service.requests}
              </Text>
            </Box>
          </Flex>
          
          {service.uptime && (
            <Box marginTop="xs">
              <Text variant="hint" fontSize={12}>
                可用率: <Text fontWeight="medium">{service.uptime}</Text>
              </Text>
            </Box>
          )}
        </Box>
      </Flex>
    </Card>
  );
};