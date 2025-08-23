import React from 'react';
import { Box } from '@workday/canvas-kit-react/layout';
import { Card } from '@workday/canvas-kit-react/card';
import { Text } from '@workday/canvas-kit-react/text';
import type { StatCardProps } from './StatsTypes';

export const StatCard: React.FC<StatCardProps> = ({ 
  title, 
  stats, 
  variant = 'default' 
}) => {
  // 防护检查：确保 stats 是一个对象且不为空
  if (!stats || typeof stats !== 'object') {
    return (
      <Card height="100%" data-testid={`stat-card-${title.replace(/\s+/g, '-').toLowerCase()}`}>
        <Card.Heading>{title}</Card.Heading>
        <Card.Body>
          <div 
            className={`stat-card-content ${variant}`}
            style={{ 
              display: 'flex', 
              flexDirection: 'column', 
              justifyContent: 'center', 
              height: '100%' 
            }}
          >
            <Box paddingY="xs">
              <Text>暂无数据</Text>
            </Box>
          </div>
        </Card.Body>
      </Card>
    );
  }

  return (
    <Card height="100%" data-testid={`stat-card-${title.replace(/\s+/g, '-').toLowerCase()}`}>
      <Card.Heading>{title}</Card.Heading>
      <Card.Body>
        <div 
          className={`stat-card-content ${variant}`}
          style={{ 
            display: 'flex', 
            flexDirection: 'column', 
            justifyContent: 'center', 
            height: '100%' 
          }}
        >
          {Object.entries(stats).map(([key, value], index) => (
            <Box key={`${title}-${key}-${index}`} paddingY="xs">
              <Text>{key}: {value}</Text>
            </Box>
          ))}
        </div>
      </Card.Body>
    </Card>
  );
};