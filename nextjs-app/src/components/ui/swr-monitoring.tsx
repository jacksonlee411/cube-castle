import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { logger } from '@/lib/logger';

interface SWRMonitoringProps {
  className?: string;
}

const SWRMonitoring: React.FC<SWRMonitoringProps> = ({ className = '' }) => {
  const [insights, setInsights] = useState<any>(null);
  const [overallMetrics, setOverallMetrics] = useState<any>(null);
  const [isVisible, setIsVisible] = useState(false);

  const refreshMetrics = () => {
    const performanceInsights = logger.getPerformanceInsights();
    const swrMetrics = logger.getSWRMetrics();
    
    setInsights(performanceInsights);
    setOverallMetrics(swrMetrics);
  };

  useEffect(() => {
    refreshMetrics();
    
    // Auto-refresh metrics every 30 seconds
    const interval = setInterval(refreshMetrics, 30000);
    
    return () => clearInterval(interval);
  }, []);

  if (!isVisible) {
    return (
      <div className={`fixed bottom-4 right-4 z-50 ${className}`}>
        <Button
          onClick={() => setIsVisible(true)}
          variant="outline"
          size="sm"
          className="bg-blue-50 hover:bg-blue-100 border-blue-200"
        >
          📊 SWR监控
        </Button>
      </div>
    );
  }

  return (
    <div className={`fixed bottom-4 right-4 z-50 w-80 max-h-96 overflow-y-auto ${className}`}>
      <Card className="shadow-lg border-blue-200">
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm">SWR性能监控</CardTitle>
            <div className="flex gap-2">
              <Button onClick={refreshMetrics} size="sm" variant="ghost" className="h-6 w-6 p-0">
                🔄
              </Button>
              <Button onClick={() => setIsVisible(false)} size="sm" variant="ghost" className="h-6 w-6 p-0">
                ✕
              </Button>
            </div>
          </div>
          <CardDescription className="text-xs">
            实时数据获取性能指标
          </CardDescription>
        </CardHeader>
        
        <CardContent className="text-xs space-y-3">
          {/* Overall Health */}
          {insights?.overallHealth && (
            <div className="space-y-2">
              <h4 className="font-medium text-sm">整体健康状况</h4>
              <div className="grid grid-cols-2 gap-2">
                <div className="bg-green-50 p-2 rounded">
                  <div className="text-green-700 font-medium">
                    成功率: {(insights.overallHealth.successRate * 100).toFixed(1)}%
                  </div>
                </div>
                <div className="bg-blue-50 p-2 rounded">
                  <div className="text-blue-700 font-medium">
                    平均响应: {insights.overallHealth.avgResponseTime.toFixed(0)}ms
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Recent Metrics */}
          {overallMetrics?.recentMetrics && overallMetrics.recentMetrics.length > 0 && (
            <div className="space-y-2">
              <h4 className="font-medium text-sm">最近请求</h4>
              <div className="space-y-1 max-h-20 overflow-y-auto">
                {overallMetrics.recentMetrics.slice(-3).map((metric: any, index: number) => (
                  <div key={index} className="flex items-center justify-between text-xs">
                    <span className="truncate flex-1 mr-2">
                      {metric.action.replace('/api/employees', 'employees')}
                    </span>
                    <div className="flex items-center gap-1">
                      <Badge variant={metric.success ? 'default' : 'destructive'} className="h-4 text-xs">
                        {metric.duration}ms
                      </Badge>
                      <span>{metric.success ? '✅' : '❌'}</span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Slowest Endpoints */}
          {insights?.slowestEndpoints && insights.slowestEndpoints.length > 0 && (
            <div className="space-y-2">
              <h4 className="font-medium text-sm">最慢端点</h4>
              <div className="space-y-1">
                {insights.slowestEndpoints.slice(0, 3).map((endpoint: any, index: number) => (
                  <div key={index} className="flex items-center justify-between text-xs">
                    <span className="truncate flex-1 mr-2">
                      {endpoint.key.replace('/api/employees', 'employees')}
                    </span>
                    <Badge variant="secondary" className="h-4 text-xs">
                      {endpoint.avgTime.toFixed(0)}ms
                    </Badge>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Error Prone Endpoints */}
          {insights?.errorProneEndpoints && insights.errorProneEndpoints.length > 0 && (
            <div className="space-y-2">
              <h4 className="font-medium text-sm text-red-600">错误率高的端点</h4>
              <div className="space-y-1">
                {insights.errorProneEndpoints.slice(0, 3).map((endpoint: any, index: number) => (
                  <div key={index} className="flex items-center justify-between text-xs">
                    <span className="truncate flex-1 mr-2">
                      {endpoint.key.replace('/api/employees', 'employees')}
                    </span>
                    <Badge variant="destructive" className="h-4 text-xs">
                      {(endpoint.errorRate * 100).toFixed(1)}%
                    </Badge>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Clear Metrics Button */}
          <div className="pt-2 border-t">
            <Button
              onClick={() => {
                logger.clearMetrics();
                refreshMetrics();
              }}
              size="sm"
              variant="outline"
              className="w-full h-6 text-xs"
            >
              清除监控数据
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export { SWRMonitoring };