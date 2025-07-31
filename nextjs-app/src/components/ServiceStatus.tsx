// src/components/ServiceStatus.tsx
import React, { useState, useEffect, useCallback } from 'react';
import { Alert, Badge, Button, Space, Tooltip } from 'antd';
import { 
  CheckCircleOutlined, 
  CloseCircleOutlined, 
  LoadingOutlined,
  SyncOutlined 
} from '@ant-design/icons';
import { restApiClient } from '@/lib/rest-api-client';
import { apolloClient } from '@/lib/graphql-client';
import { gql } from '@apollo/client';

interface ServiceStatusProps {
  showDetails?: boolean;
  className?: string;
  style?: React.CSSProperties;
}

interface ServiceHealth {
  graphql: 'healthy' | 'unhealthy' | 'checking';
  rest: 'healthy' | 'unhealthy' | 'checking';
  lastCheck: Date;
}

const ServiceStatus: React.FC<ServiceStatusProps> = ({ 
  showDetails = false, 
  className,
  style 
}) => {
  const [serviceHealth, setServiceHealth] = useState<ServiceHealth>({
    graphql: 'checking',
    rest: 'checking',
    lastCheck: new Date(),
  });
  const [isChecking, setIsChecking] = useState(false);

  const checkGraphQLHealth = async (): Promise<boolean> => {
    try {
      const result = await apolloClient.query({
        query: gql`
          query HealthCheck {
            __typename
          }
        `,
        fetchPolicy: 'network-only',
        errorPolicy: 'none',
      });
      return !!result.data;
    } catch (error) {
      // GraphQL health check failed
      return false;
    }
  };

  const checkRESTHealth = async (): Promise<boolean> => {
    try {
      const result = await restApiClient.healthCheck();
      return result.success;
    } catch (error) {
      // REST API health check failed
      return false;
    }
  };

  const performHealthCheck = useCallback(async () => {
    setIsChecking(true);
    
    const [graphqlHealthy, restHealthy] = await Promise.all([
      checkGraphQLHealth(),
      checkRESTHealth(),
    ]);

    setServiceHealth({
      graphql: graphqlHealthy ? 'healthy' : 'unhealthy',
      rest: restHealthy ? 'healthy' : 'unhealthy',
      lastCheck: new Date(),
    });
    
    setIsChecking(false);
  }, []);

  useEffect(() => {
    performHealthCheck();
    
    // Check health every 30 seconds
    const interval = setInterval(performHealthCheck, 30000);
    
    return () => clearInterval(interval);
  }, [performHealthCheck]);

  const getStatusColor = (status: ServiceHealth['graphql']) => {
    switch (status) {
      case 'healthy': return 'success';
      case 'unhealthy': return 'error';
      case 'checking': return 'processing';
      default: return 'default';
    }
  };

  const getStatusIcon = (status: ServiceHealth['graphql']) => {
    switch (status) {
      case 'healthy': return <CheckCircleOutlined />;
      case 'unhealthy': return <CloseCircleOutlined />;
      case 'checking': return <LoadingOutlined />;
      default: return <CloseCircleOutlined />;
    }
  };

  const getStatusText = (status: ServiceHealth['graphql']) => {
    switch (status) {
      case 'healthy': return '正常';
      case 'unhealthy': return '异常';
      case 'checking': return '检查中';
      default: return '未知';
    }
  };

  const isAllHealthy = serviceHealth.graphql === 'healthy' && serviceHealth.rest === 'healthy';
  const hasUnhealthyService = serviceHealth.graphql === 'unhealthy' || serviceHealth.rest === 'unhealthy';

  if (!showDetails) {
    // Simple badge display
    return (
      <Tooltip 
        title={`GraphQL: ${getStatusText(serviceHealth.graphql)}, REST: ${getStatusText(serviceHealth.rest)}`}
      >
        <Badge 
          status={isAllHealthy ? 'success' : hasUnhealthyService ? 'error' : 'processing'} 
          text="服务状态"
          className={className}
          style={style}
        />
      </Tooltip>
    );
  }

  return (
    <div className={className} style={style}>
      {hasUnhealthyService && (
        <Alert
          message="服务连接异常"
          description="部分服务不可用，系统将自动使用备用服务。"
          type="warning"
          showIcon
          style={{ marginBottom: '16px' }}
          action={
            <Button 
              size="small" 
              icon={<SyncOutlined />} 
              onClick={performHealthCheck}
              loading={isChecking}
            >
              重新检查
            </Button>
          }
        />
      )}
      
      <Space direction="vertical" style={{ width: '100%' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span>服务状态：</span>
          <Button 
            size="small" 
            icon={<SyncOutlined />} 
            onClick={performHealthCheck}
            loading={isChecking}
          >
            刷新
          </Button>
        </div>
        
        <Space>
          <Badge 
            status={getStatusColor(serviceHealth.graphql)} 
            text={
              <span>
                GraphQL: {getStatusText(serviceHealth.graphql)}
                {serviceHealth.graphql === 'healthy' && 
                  <Tooltip title="实时数据同步可用">
                    {' '}(完整功能)
                  </Tooltip>
                }
              </span>
            }
          />
        </Space>
        
        <Space>
          <Badge 
            status={getStatusColor(serviceHealth.rest)} 
            text={
              <span>
                REST API: {getStatusText(serviceHealth.rest)}
                {serviceHealth.rest === 'healthy' && serviceHealth.graphql === 'unhealthy' &&
                  <Tooltip title="基础功能可用，但缺少实时更新">
                    {' '}(备用服务)
                  </Tooltip>
                }
              </span>
            }
          />
        </Space>
        
        <div style={{ fontSize: '12px', color: '#666' }}>
          上次检查: {serviceHealth.lastCheck.toLocaleTimeString()}
        </div>
      </Space>
    </div>
  );
};

export default ServiceStatus;