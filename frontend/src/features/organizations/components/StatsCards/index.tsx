import React from 'react';
import { Box } from '@workday/canvas-kit-react/layout';
import { Card } from '@workday/canvas-kit-react/card';
import { Text } from '@workday/canvas-kit-react/text';
import { StatCard } from './StatCard';
import type { StatsCardsProps } from './StatsTypes';

export const StatsCards: React.FC<StatsCardsProps> = ({ stats }) => {
  return (
    <div 
      style={{ 
        marginBottom: '16px', 
        display: 'flex', 
        alignItems: 'stretch', 
        gap: '16px' 
      }}
      data-testid="stats-cards-container"
    >
      <Box flex={1}>
        <StatCard 
          title="按类型统计" 
          stats={stats.by_type} 
        />
      </Box>
      <Box flex={1}>
        <StatCard 
          title="按状态统计" 
          stats={stats.by_status} 
        />
      </Box>
      <Box flex={1}>
        <Card height="100%">
          <Card.Heading>总体概况</Card.Heading>
          <Card.Body>
            <div 
              style={{ 
                textAlign: 'center', 
                display: 'flex', 
                flexDirection: 'column', 
                justifyContent: 'center', 
                height: '100%' 
              }}
            >
              <Text fontWeight="bold" style={{ fontSize: '2rem' }}>
                {stats.total_count}
              </Text>
              <Text>组织单元总数</Text>
            </div>
          </Card.Body>
        </Card>
      </Box>
    </div>
  );
};