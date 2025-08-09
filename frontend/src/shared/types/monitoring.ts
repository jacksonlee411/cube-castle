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
}

export interface SystemMetrics {
  services: ServiceStatus[];
  charts: ChartData;
  lastUpdated: string;
}

// 模拟监控数据（用于MVP版本）
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
    { 
      name: '指标收集服务', 
      status: 'online', 
      port: '9999', 
      responseTime: '8ms', 
      requests: '15',
      uptime: '100%' 
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
    ]
  },
  lastUpdated: new Date().toLocaleString()
};