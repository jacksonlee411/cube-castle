import React from 'react';
import { Box, Grid } from '@workday/canvas-kit-react/layout';
import { space } from '@workday/canvas-kit-react/tokens';
import type { ServiceStatus } from '../../../shared/types/monitoring';
import { ServiceCard } from './ServiceCard';

interface ServiceStatusGridProps {
  services?: ServiceStatus[] | undefined;
}

export const ServiceStatusGrid: React.FC<ServiceStatusGridProps> = ({ services = [] }) => {
  if (services.length === 0) {
    return (
      <Box padding="l" textAlign="center">
        <Box as="span" fontSize="48px">📊</Box>
        <Box marginTop="s">暂无服务数据</Box>
      </Box>
    );
  }

  return (
    <Grid
      gridTemplateColumns="repeat(3, 1fr)"
      gridGap={space.m}
    >
      {services.map((service, index) => (
        <ServiceCard key={service.name + index} service={service} />
      ))}
    </Grid>
  );
};