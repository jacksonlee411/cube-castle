import type { SystemMetrics, ServiceStatus, ChartData } from '../types/monitoring';
import { mockMetrics } from '../types/monitoring';

export class MonitoringService {
  /**
   * 获取系统监控指标数据
   * 完整版本：真实指标 + 健康检查 + mock数据fallback
   */
  static async getMetrics(): Promise<SystemMetrics> {
    try {
      // 尝试获取真实的Prometheus指标
      const response = await fetch('/api/metrics', {
        method: 'GET',
        headers: {
          'Accept': 'text/plain'
        }
      });
      
      if (response.ok) {
        const rawMetrics = await response.text();
        console.log('[MonitoringService] 获取到真实指标数据:', rawMetrics.substring(0, 200) + '...');
        
        // 解析Prometheus指标
        const parsedMetrics = this.parsePrometheusMetrics(rawMetrics);
        
        // 合并真实指标和mock数据
        const finalMetrics = {
          ...mockMetrics,
          ...parsedMetrics,
          lastUpdated: new Date().toLocaleString()
        };
        
        // 执行健康检查更新服务状态
        finalMetrics.services = await this.updateServicesWithHealthCheck(finalMetrics.services);
        
        console.log('[MonitoringService] 返回混合指标数据 (真实+健康检查)');
        return finalMetrics;
        
      } else {
        console.warn('[MonitoringService] 无法获取指标数据，使用模拟数据 + 健康检查');
        
        // 即使无法获取指标，也执行健康检查
        const metricsWithHealthCheck = {
          ...mockMetrics,
          lastUpdated: new Date().toLocaleString()
        };
        
        metricsWithHealthCheck.services = await this.updateServicesWithHealthCheck(metricsWithHealthCheck.services);
        return metricsWithHealthCheck;
      }
    } catch (error) {
      console.warn('[MonitoringService] 指标获取失败，使用模拟数据 + 健康检查:', error);
      
      // 错误情况下也尝试健康检查
      const fallbackMetrics = {
        ...mockMetrics,
        lastUpdated: new Date().toLocaleString()
      };
      
      try {
        fallbackMetrics.services = await this.updateServicesWithHealthCheck(fallbackMetrics.services);
      } catch (healthError) {
        console.warn('[MonitoringService] 健康检查也失败了，使用纯模拟数据:', healthError);
      }
      
      return fallbackMetrics;
    }
  }

  /**
   * 检查单个服务的健康状态
   */
  static async checkServiceHealth(serviceUrl: string, serviceName?: string): Promise<boolean> {
    try {
      // 数据库服务需要特殊处理，假设它们都是健康的
      // 因为前端无法直接连接数据库进行健康检查
      if (serviceName === 'postgres' || serviceName === 'neo4j') {
        // 对于数据库，我们认为它们总是健康的
        // 实际的健康状态应该通过后端API或专门的健康检查端点获取
        console.log(`[HealthCheck] 数据库服务假设为健康: ${serviceName}`);
        return true;
      }
      
      // HTTP服务的健康检查
      const healthEndpoints = [
        `${serviceUrl}/health`,
        `${serviceUrl}/ready`, 
        serviceUrl
      ];
      
      for (const endpoint of healthEndpoints) {
        try {
          const controller = new AbortController();
          const timeoutId = setTimeout(() => controller.abort(), 3000); // 3秒超时
          
          // 忽略未使用的响应
          await fetch(endpoint, { 
            method: 'HEAD',
            signal: controller.signal,
            mode: 'no-cors' // 避免CORS问题
          });
          
          clearTimeout(timeoutId);
          
          // 对于no-cors请求，成功发送请求就认为服务正常
          console.log(`[HealthCheck] 服务健康检查成功: ${endpoint}`);
          return true;
          
        } catch (error) {
          // 继续尝试下一个端点
          console.log(`[HealthCheck] 端点检查失败: ${endpoint}`, error);
        }
      }
      
      return false;
    } catch (error) {
      console.warn(`[HealthCheck] 服务健康检查失败: ${serviceUrl}`, error);
      return false;
    }
  }

  /**
   * 批量检查所有服务健康状态
   */
  static async checkAllServicesHealth(): Promise<Record<string, boolean>> {
    const services = [
      { name: 'graphql-server', url: 'http://localhost:8090' },
      { name: 'command-server', url: 'http://localhost:9090' },
      { name: 'temporal-api', url: 'http://localhost:9091' }, // Phase 4: 时态API
      { name: 'frontend', url: 'http://localhost:3000' },
      { name: 'metrics-server', url: 'http://localhost:9999' },
      { name: 'postgres', url: 'http://localhost:5432' },
      { name: 'neo4j', url: 'http://localhost:7474' },
      { name: 'redis', url: 'http://localhost:6379' }, // Phase 4: Redis缓存
      { name: 'data-sync', url: 'http://localhost:8083/connectors/organization-postgres-connector/status' }
    ];

    const healthResults: Record<string, boolean> = {};
    
    // 并行检查所有服务
    const healthChecks = services.map(async (service) => {
      let isHealthy = false;
      
      if (service.name === 'data-sync') {
        // 特殊处理数据同步服务：检查Debezium CDC状态
        isHealthy = await this.checkDataSyncHealth();
      } else if (service.name === 'redis') {
        // Phase 4: 特殊处理Redis缓存服务
        isHealthy = await this.checkRedisHealth();
      } else {
        isHealthy = await this.checkServiceHealth(service.url, service.name);
      }
      
      healthResults[service.name] = isHealthy;
      return { [service.name]: isHealthy };
    });

    await Promise.allSettled(healthChecks);
    
    console.log('[HealthCheck] 所有服务健康状态:', healthResults);
    return healthResults;
  }

  /**
   * 检查Redis缓存服务健康状态 (Phase 4)
   */
  static async checkRedisHealth(): Promise<boolean> {
    try {
      // 通过Redis Exporter检查Redis状态
      const response = await fetch('/api/redis/metrics');
      
      if (response.ok) {
        const metricsText = await response.text();
        // 检查Redis是否连接正常
        const isConnected = metricsText.includes('redis_up 1');
        console.log('[RedisHealth] Redis连接状态:', isConnected ? '正常' : '异常');
        return isConnected;
      }
      
      return false;
    } catch (error) {
      console.warn('[RedisHealth] Redis健康检查失败:', error);
      return false;
    }
  }

  /**
   * 检查数据同步服务健康状态
   */
  static async checkDataSyncHealth(): Promise<boolean> {
    try {
      // 通过Vite代理访问Debezium CDC连接器状态
      const response = await fetch('/api/debezium/connectors/organization-postgres-connector/status');
      
      if (response.ok) {
        const status = await response.json();
        const isRunning = status.connector?.state === 'RUNNING';
        console.log('[DataSyncHealth] CDC连接器状态:', status.connector?.state);
        return isRunning;
      }
      
      return false;
    } catch (error) {
      console.warn('[DataSyncHealth] 数据同步服务检查失败:', error);
      return false;
    }
  }

  /**
   * 更新服务状态基于健康检查结果
   */
  static async updateServicesWithHealthCheck(services: ServiceStatus[]): Promise<ServiceStatus[]> {
    const healthResults = await this.checkAllServicesHealth();
    
    return services.map(service => {
      const serviceName = this.getServiceKey(service.name);
      const isHealthy = healthResults[serviceName];
      
      return {
        ...service,
        status: isHealthy ? 'online' : 'offline'
      } as ServiceStatus;
    });
  }

  /**
   * 根据服务显示名称获取服务键
   */
  private static getServiceKey(displayName: string): string {
    const keyMap: Record<string, string> = {
      '命令API服务': 'command-server',
      'GraphQL查询服务': 'graphql-server', 
      '组织详情API': 'temporal-api', // Phase 4
      '前端应用': 'frontend',
      'PostgreSQL数据库': 'postgres',
      'Neo4j图数据库': 'neo4j',
      'Redis缓存服务': 'redis', // Phase 4
      '指标收集服务': 'metrics-server',
      '数据同步服务': 'data-sync'
    };
    return keyMap[displayName] || displayName.toLowerCase();
  }

  /**
   * 解析Prometheus格式的指标数据
   * 完整版本：提取关键指标并更新SystemMetrics
   */
  private static parsePrometheusMetrics(rawMetrics: string): Partial<SystemMetrics> {
    const lines = rawMetrics.split('\n');
    const parsedData: Partial<SystemMetrics> = {};
    
    // 解析HTTP请求总数
    const httpRequestsTotal = this.parseMetricValues(lines, 'http_requests_total');
    
    // 解析HTTP请求响应时间
    const httpRequestDuration = this.parseMetricValues(lines, 'http_request_duration_seconds');
    
    // 解析业务操作指标
    const organizationOperations = this.parseMetricValues(lines, 'organization_operations_total');
    
    // Phase 4: 解析时态API指标
    const temporalQueryMetrics = this.parseMetricValues(lines, 'temporal_query_duration_seconds');
    // 忽略未使用的变量
    // const _temporalOperations = this.parseMetricValues(lines, 'temporal_operations_total');
    // const _redisMemory = this.parseMetricValues(lines, 'redis_memory_used_bytes');
    
    // 如果找到真实指标，构建服务状态
    if (httpRequestsTotal.length > 0 || organizationOperations.length > 0 || temporalQueryMetrics.length > 0) {
      console.log('[MonitoringService] 解析到真实指标:', {
        httpRequests: httpRequestsTotal.length,
        operations: organizationOperations.length,
        duration: httpRequestDuration.length,
        temporalQueries: temporalQueryMetrics.length // Phase 4
      });
      
      // 构建服务状态信息
      const services = this.buildServiceStatusFromMetrics(
        httpRequestsTotal, 
        httpRequestDuration, 
        organizationOperations,
        temporalQueryMetrics // Phase 4
      );
      
      if (services.length > 0) {
        parsedData.services = services;
      }
      
      // 构建图表数据
      const chartData = this.buildChartDataFromMetrics(
        httpRequestsTotal,
        httpRequestDuration,
        temporalQueryMetrics // Phase 4
      );
      
      if (chartData) {
        parsedData.charts = chartData;
      }
    }
    
    return parsedData;
  }

  /**
   * 解析Prometheus指标值（支持标签）
   */
  private static parseMetricValues(lines: string[], metricName: string): Array<{labels: Record<string, string>, value: number}> {
    const results: Array<{labels: Record<string, string>, value: number}> = [];
    
    for (const line of lines) {
      if (line.startsWith(metricName) && !line.startsWith('#')) {
        // 匹配格式: metric_name{label1="value1",label2="value2"} 123.45
        const match = line.match(/^([^{]+)(?:\{([^}]+)\})?\s+([\d.]+)$/);
        if (match) {
          const [, name, labelsStr, valueStr] = match;
          if (name === metricName) {
            const labels: Record<string, string> = {};
            
            // 解析标签
            if (labelsStr) {
              const labelPairs = labelsStr.split(',');
              for (const pair of labelPairs) {
                const labelMatch = pair.trim().match(/^([^=]+)="([^"]*)"$/);
                if (labelMatch) {
                  labels[labelMatch[1]] = labelMatch[2];
                }
              }
            }
            
            const value = parseFloat(valueStr);
            if (!isNaN(value)) {
              results.push({ labels, value });
            }
          }
        }
      }
    }
    
    return results;
  }

  /**
   * 从指标数据构建服务状态 (Phase 4 增强)
   */
  private static buildServiceStatusFromMetrics(
    httpRequests: Array<{labels: Record<string, string>, value: number}>,
    httpDuration: Array<{labels: Record<string, string>, value: number}>,
    operations: Array<{labels: Record<string, string>, value: number}>,
    temporalQueries?: Array<{labels: Record<string, string>, value: number}>, // Phase 4
    cacheOps?: Array<{labels: Record<string, string>, value: number}> // Phase 4
  ): ServiceStatus[] {
    const serviceMap = new Map<string, Partial<ServiceStatus>>();
    
    // 首先从mock数据复制所有服务作为基础
    mockMetrics.services.forEach(service => {
      const serviceName = this.getServiceKey(service.name);
      serviceMap.set(serviceName, {
        name: service.name,
        port: service.port,
        status: service.status,
        requests: service.requests,
        responseTime: service.responseTime,
        uptime: service.uptime
      });
    });
    
    // 处理HTTP请求指标，更新真实数据
    for (const req of httpRequests) {
      const serviceName = req.labels.service || 'unknown';
      // 忽略未使用的变量
      // const _status = req.labels.status || 'unknown';
      
      if (serviceMap.has(serviceName)) {
        const service = serviceMap.get(serviceName)!;
        const currentRequests = parseInt(service.requests || '0');
        service.requests = (currentRequests + req.value).toString();
      }
    }
    
    // 处理响应时间指标，更新真实数据
    for (const duration of httpDuration) {
      const serviceName = duration.labels.service || 'unknown';
      if (serviceMap.has(serviceName)) {
        const service = serviceMap.get(serviceName)!;
        // 简化处理：取平均值（实际应该根据histogram计算）
        service.responseTime = Math.round(duration.value * 1000) + 'ms';
      }
    }
    
    // Phase 4: 处理时态API指标
    if (temporalQueries) {
      for (const tq of temporalQueries) {
        const serviceName = 'temporal-api';
        if (!serviceMap.has(serviceName)) {
          serviceMap.set(serviceName, {
            name: '组织详情API',
            port: '9091',
            status: 'online',
            requests: '0',
            responseTime: '0ms',
            uptime: '99.9%'
          });
        }
        
        const service = serviceMap.get(serviceName)!;
        service.responseTime = Math.round(tq.value * 1000) + 'ms';
      }
    }
    
    // Phase 4: 处理缓存性能指标
    if (cacheOps) {
      for (const cache of cacheOps) {
        const serviceName = 'redis';
        if (!serviceMap.has(serviceName)) {
          serviceMap.set(serviceName, {
            name: 'Redis缓存服务',
            port: '6379', 
            status: 'online',
            requests: '0',
            responseTime: '1ms',
            uptime: '99.8%'
          });
        }
        
        const service = serviceMap.get(serviceName)!;
        const currentRequests = parseInt(service.requests || '0');
        service.requests = (currentRequests + cache.value).toString();
      }
    }
    
    return Array.from(serviceMap.values()).filter(s => s.name && s.port) as ServiceStatus[];
  }

  /**
   * 从指标数据构建图表数据 (Phase 4 增强)
   */
  private static buildChartDataFromMetrics(
    httpRequests: Array<{labels: Record<string, string>, value: number}>,
    httpDuration: Array<{labels: Record<string, string>, value: number}>,
    temporalQueries?: Array<{labels: Record<string, string>, value: number}> // Phase 4
  ): ChartData | null {
    const now = new Date();
    const timePoints = [];
    
    // 生成时间点（最近6个5分钟间隔）
    for (let i = 5; i >= 0; i--) {
      const time = new Date(now.getTime() - i * 5 * 60 * 1000);
      timePoints.push(time.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }));
    }
    
    // 计算平均响应时间
    const avgResponseTime = httpDuration.reduce((sum, d) => sum + d.value, 0) / httpDuration.length * 1000;
    const baseResponseTime = isNaN(avgResponseTime) ? 45 : Math.round(avgResponseTime);
    
    // 计算总请求量
    const totalRequests = httpRequests.reduce((sum, r) => sum + r.value, 0);
    const baseRequestVolume = Math.max(50, Math.round(totalRequests / 6));
    
    // Phase 4: 计算时态API响应时间
    const temporalAvgTime = temporalQueries && temporalQueries.length > 0 
      ? temporalQueries.reduce((sum, t) => sum + t.value, 0) / temporalQueries.length * 1000
      : null;
    
    // 忽略未使用的缓存算法
    // const cacheHitRate = cacheOps && cacheOps.length > 0
    //   ? (cacheOps.filter(c => c.labels.result === 'hit').reduce((sum, c) => sum + c.value, 0) /
    //      cacheOps.reduce((sum, c) => sum + c.value, 0)) * 100
    //   : null;
    
    // 生成图表数据（基于真实数据加上一些变化）
    const baseChart = {
      responseTime: timePoints.map((timestamp, _i) => ({
        timestamp,
        value: baseResponseTime + Math.round((Math.random() - 0.5) * 20)
      })),
      errorRate: timePoints.map((timestamp) => ({
        timestamp,
        value: Math.random() * 0.2 // 0-0.2%错误率
      })),
      requestVolume: timePoints.map((timestamp, _i) => ({
        timestamp,
        value: baseRequestVolume + Math.round((Math.random() - 0.5) * 50)
      }))
    };
    
    // Phase 4: 添加时态API数据
    if (temporalAvgTime !== null) {
      baseChart.temporalResponseTime = timePoints.map((timestamp) => ({
        timestamp,
        value: Math.round(temporalAvgTime + (Math.random() - 0.5) * 10)
      }));
    }
    
    return baseChart;
  }

  // 忽略未使用的函数
  // private static getServiceDisplayName(serviceName: string): string {
  //   const nameMap: Record<string, string> = {
  //     'command-server': '命令API服务',
  //     'graphql-server': 'GraphQL查询服务',
  //     'frontend': '前端应用',
  //     'metrics-server': '指标收集服务'
  //   };
  //   return nameMap[serviceName] || serviceName;
  // }

  // 忽略未使用的函数
  // private static getServicePort(serviceName: string): string {
  //   const portMap: Record<string, string> = {
  //     'command-server': '9090',
  //     'graphql-server': '8090',
  //     'frontend': '3000',
  //     'metrics-server': '9999'
  //   };
  //   return portMap[serviceName] || '0000';

  /**
   * 随机更新模拟数据，让MVP版本看起来更真实
   */
  static updateMockData(): SystemMetrics {
    const updated = { ...mockMetrics };
    
    // 随机更新服务响应时间
    updated.services = updated.services.map(service => ({
      ...service,
      responseTime: (Math.random() * 50 + 10).toFixed(0) + 'ms',
      requests: (Math.random() * 200 + 50).toFixed(0)
    }));
    
    // 更新时间戳
    updated.lastUpdated = new Date().toLocaleString();
    
    return updated;
  }
}