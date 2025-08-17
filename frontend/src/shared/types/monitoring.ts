// 监控相关类型定义
export interface ServiceStatus {
  name: string;
  status: 'online' | 'offline' | 'warning';
  port: string;
  responseTime: string;
  requests: string;
  uptime?: string;
}

export interface MetricPoint {
  timestamp: string;
  value: number;
}

export interface ChartData {
  responseTime: MetricPoint[];
  errorRate: MetricPoint[];
  requestVolume: MetricPoint[];
  // Phase 4: 新增时态API和缓存性能指标
  temporalResponseTime?: MetricPoint[]; // 时态API响应时间
  cacheHitRate?: MetricPoint[];          // 缓存命中率
  cacheMemoryUsage?: MetricPoint[];      // Redis内存使用
}

export interface SystemMetrics {
  services: ServiceStatus[];
  charts: ChartData;
  lastUpdated: string;
  // Phase 4: 新增性能统计
  performanceStats?: {
    graphqlImprovement: string;    // GraphQL性能提升百分比
    temporalImprovement: string;   // 时态API性能提升百分比
    cacheHitRate: string;          // 当前缓存命中率
    avgResponseTime: string;       // 平均响应时间
  };
}

// 模拟监控数据（用于MVP版本）- Phase 4增强版
export const mockMetrics: SystemMetrics = {
  services: [
    { 
      name: 'GraphQL查询服务', 
      status: 'online', 
      port: '8090', 
      responseTime: '45ms', 
      requests: '127',
      uptime: '99.9%' 
    },
    { 
      name: '命令API服务', 
      status: 'online', 
      port: '9090', 
      responseTime: '32ms', 
      requests: '89',
      uptime: '99.8%' 
    },
    // Phase 4: 新增组织详情API
    {
      name: '组织详情API',
      status: 'online',
      port: '9091',
      responseTime: '12ms', // 94%性能提升后的响应时间
      requests: '64',
      uptime: '99.9%'
    },
    { 
      name: '前端应用', 
      status: 'online', 
      port: '3000', 
      responseTime: '12ms', 
      requests: '245',
      uptime: '100%' 
    },
    { 
      name: 'PostgreSQL数据库', 
      status: 'online', 
      port: '5432', 
      responseTime: '5ms', 
      requests: '234',
      uptime: '99.9%' 
    },
    { 
      name: 'Neo4j图数据库', 
      status: 'online', 
      port: '7474', 
      responseTime: '18ms', 
      requests: '78',
      uptime: '99.5%' 
    },
    // Phase 4: 新增Redis缓存服务
    {
      name: 'Redis缓存服务',
      status: 'online',
      port: '6379',
      responseTime: '1ms', // 缓存响应时间
      requests: '1247', // 缓存操作数
      uptime: '99.8%'
    },
    { 
      name: '指标收集服务', 
      status: 'online', 
      port: '9999', 
      responseTime: '8ms', 
      requests: '15',
      uptime: '100%' 
    },
    { 
      name: '数据同步服务', 
      status: 'online', 
      port: 'CDC', 
      responseTime: '2ms', 
      requests: 'PostgreSQL→Neo4j',
      uptime: '99.7%' 
    }
  ],
  charts: {
    responseTime: [
      { timestamp: '14:00', value: 45 },
      { timestamp: '14:05', value: 52 },
      { timestamp: '14:10', value: 38 },
      { timestamp: '14:15', value: 41 },
      { timestamp: '14:20', value: 47 },
      { timestamp: '14:25', value: 35 }
    ],
    errorRate: [
      { timestamp: '14:00', value: 0.1 },
      { timestamp: '14:05', value: 0.2 },
      { timestamp: '14:10', value: 0.05 },
      { timestamp: '14:15', value: 0.15 },
      { timestamp: '14:20', value: 0.08 },
      { timestamp: '14:25', value: 0.12 }
    ],
    requestVolume: [
      { timestamp: '14:00', value: 120 },
      { timestamp: '14:05', value: 135 },
      { timestamp: '14:10', value: 98 },
      { timestamp: '14:15', value: 156 },
      { timestamp: '14:20', value: 142 },
      { timestamp: '14:25', value: 167 }
    ],
    // Phase 4: 新增时态API性能数据
    temporalResponseTime: [
      { timestamp: '14:00', value: 15 },
      { timestamp: '14:05', value: 12 },
      { timestamp: '14:10', value: 18 },
      { timestamp: '14:15', value: 11 },
      { timestamp: '14:20', value: 14 },
      { timestamp: '14:25', value: 16 }
    ],
    // Phase 4: 新增缓存性能数据
    cacheHitRate: [
      { timestamp: '14:00', value: 91.5 },
      { timestamp: '14:05', value: 92.1 },
      { timestamp: '14:10', value: 91.8 },
      { timestamp: '14:15', value: 92.3 },
      { timestamp: '14:20', value: 91.7 },
      { timestamp: '14:25', value: 92.0 }
    ],
    cacheMemoryUsage: [
      { timestamp: '14:00', value: 1.25 },
      { timestamp: '14:05', value: 1.31 },
      { timestamp: '14:10', value: 1.28 },
      { timestamp: '14:15', value: 1.33 },
      { timestamp: '14:20', value: 1.30 },
      { timestamp: '14:25', value: 1.29 }
    ]
  },
  // Phase 4: 新增性能统计
  performanceStats: {
    graphqlImprovement: '65%',
    temporalImprovement: '94%', 
    cacheHitRate: '91.7%',
    avgResponseTime: '3.7ms'
  },
  lastUpdated: new Date().toLocaleString()
};