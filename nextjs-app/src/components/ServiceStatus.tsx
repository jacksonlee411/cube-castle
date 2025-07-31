// src/components/ServiceStatus.tsx
import React, { useState, useEffect, useCallback } from 'react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tooltip, TooltipContent, TooltipTrigger, TooltipProvider } from '@/components/ui/tooltip';
import { toast } from 'react-hot-toast';
import { 
  CheckCircle, 
  XCircle, 
  Loader2,
  RefreshCw,
  AlertTriangle
} from 'lucide-react';
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
      case 'healthy': return <CheckCircle className="h-4 w-4" />;
      case 'unhealthy': return <XCircle className="h-4 w-4" />;
      case 'checking': return <Loader2 className="h-4 w-4 animate-spin" />;
      default: return <XCircle className="h-4 w-4" />;
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
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <div className={className} style={style}>
              <Badge 
                status={isAllHealthy ? 'success' : hasUnhealthyService ? 'error' : 'processing'} 
                text="服务状态"
              />
            </div>
          </TooltipTrigger>
          <TooltipContent>
            <p>GraphQL: {getStatusText(serviceHealth.graphql)}, REST: {getStatusText(serviceHealth.rest)}</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    );
  }

  return (
    <div className={className} style={style}>
      {hasUnhealthyService && (
        <div className="rounded-lg border border-yellow-200 bg-yellow-50 p-4 mb-4 shadow-sm">
          <div className="flex items-start">
            <div className="flex-shrink-0">
              <AlertTriangle className="h-5 w-5 text-yellow-400" aria-hidden="true" />
            </div>
            <div className="ml-3 flex-1">
              <h3 className="text-sm font-medium text-yellow-800">
                服务连接异常
              </h3>
              <div className="mt-1 text-sm text-yellow-700">
                部分服务不可用，系统将自动使用备用服务。
              </div>
              <div className="mt-3">
                <Button 
                  size="sm" 
                  variant="outline"
                  onClick={performHealthCheck}
                  disabled={isChecking}
                  className="bg-white text-yellow-800 border-yellow-300 hover:bg-yellow-50"
                >
                  {isChecking ? (
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  ) : (
                    <RefreshCw className="h-4 w-4 mr-2" />
                  )}
                  重新检查
                </Button>
              </div>
            </div>
          </div>
        </div>
      )}
      
      <div className="space-y-4">
        <div className="flex justify-between items-center">
          <span className="text-sm font-medium text-gray-900">服务状态：</span>
          <Button 
            size="sm" 
            variant="outline"
            onClick={performHealthCheck}
            disabled={isChecking}
          >
            {isChecking ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <RefreshCw className="h-4 w-4 mr-2" />
            )}
            刷新
          </Button>
        </div>
        
        <div className="space-y-2">
          <div className="flex items-center">
            <Badge 
              status={getStatusColor(serviceHealth.graphql)} 
              text={
                <span className="flex items-center">
                  GraphQL: {getStatusText(serviceHealth.graphql)}
                  {serviceHealth.graphql === 'healthy' && (
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <span className="ml-1 text-xs text-gray-500">(完整功能)</span>
                        </TooltipTrigger>
                        <TooltipContent>
                          <p>实时数据同步可用</p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  )}
                </span>
              }
            />
          </div>
          
          <div className="flex items-center">
            <Badge 
              status={getStatusColor(serviceHealth.rest)} 
              text={
                <span className="flex items-center">
                  REST API: {getStatusText(serviceHealth.rest)}
                  {serviceHealth.rest === 'healthy' && serviceHealth.graphql === 'unhealthy' && (
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <span className="ml-1 text-xs text-gray-500">(备用服务)</span>
                        </TooltipTrigger>
                        <TooltipContent>
                          <p>基础功能可用，但缺少实时更新</p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  )}
                </span>
              }
            />
          </div>
        </div>
        
        <div className="text-xs text-gray-500">
          上次检查: {serviceHealth.lastCheck.toLocaleTimeString()}
        </div>
      </div>
    </div>
  );
};

export default ServiceStatus;