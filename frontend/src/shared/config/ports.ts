/**
 * 统一端口配置管理
 * 🎯 单一真源：所有端口配置的权威来源
 * 🔒 零容忍：严禁在其他文件中硬编码端口
 */

// 🎯 核心服务端口配置
export const SERVICE_PORTS = {
  // 前端开发服务器
  FRONTEND_DEV: 3000,
  FRONTEND_PREVIEW: 3001,
  
  // 后端核心服务 (CQRS架构)
  REST_COMMAND_SERVICE: 9090,    // 命令操作 (REST API)
  GRAPHQL_QUERY_SERVICE: 8090,   // 查询操作 (GraphQL)
  
  // 基础设施服务
  POSTGRESQL: 5432,
  REDIS: 6379,
  
  // 监控服务 (可选)
  PROMETHEUS: 9091,
  GRAFANA: 3002,
  ALERT_MANAGER: 9093,
  NODE_EXPORTER: 9100
} as const;

// 🎯 环境相关端口映射
export const getServicePort = (service: keyof typeof SERVICE_PORTS, env: string = 'development'): number => {
  // 开发环境使用默认端口
  if (env === 'development') {
    return SERVICE_PORTS[service];
  }
  
  // 生产环境可能需要端口映射 (预留扩展)
  // TODO: 根据部署环境调整端口映射
  return SERVICE_PORTS[service];
};

// 🎯 服务端点构造器
export const buildServiceURL = (service: keyof typeof SERVICE_PORTS, path: string = '', env: string = 'development'): string => {
  const port = getServicePort(service, env);
  const host = env === 'development' ? 'localhost' : process.env.SERVICE_HOST || 'localhost';
  const protocol = env === 'development' ? 'http' : process.env.SERVICE_PROTOCOL || 'http';
  
  return `${protocol}://${host}:${port}${path}`;
};

// 🎯 CQRS端点配置 (企业级架构标准)
export const CQRS_ENDPOINTS = {
  // 命令操作端点 (REST)
  COMMAND_BASE: buildServiceURL('REST_COMMAND_SERVICE'),
  COMMAND_API: buildServiceURL('REST_COMMAND_SERVICE', '/api/v1'),
  AUTH_ENDPOINT: buildServiceURL('REST_COMMAND_SERVICE', '/auth'),
  METRICS_COMMAND: buildServiceURL('REST_COMMAND_SERVICE', '/metrics'),
  
  // 查询操作端点 (GraphQL)
  QUERY_BASE: buildServiceURL('GRAPHQL_QUERY_SERVICE'),
  GRAPHQL_ENDPOINT: buildServiceURL('GRAPHQL_QUERY_SERVICE', '/graphql'),
  GRAPHQL_PLAYGROUND: buildServiceURL('GRAPHQL_QUERY_SERVICE', '/graphiql'),
  METRICS_QUERY: buildServiceURL('GRAPHQL_QUERY_SERVICE', '/metrics')
} as const;

// 🎯 前端开发端点
export const FRONTEND_ENDPOINTS = {
  DEV_SERVER: buildServiceURL('FRONTEND_DEV'),
  PREVIEW_SERVER: buildServiceURL('FRONTEND_PREVIEW')
} as const;

// 🎯 基础设施端点
export const INFRASTRUCTURE_ENDPOINTS = {
  DATABASE: buildServiceURL('POSTGRESQL'),
  CACHE: buildServiceURL('REDIS')
} as const;

// 🎯 监控端点 (可选)
export const MONITORING_ENDPOINTS = {
  PROMETHEUS: buildServiceURL('PROMETHEUS'),
  GRAFANA: buildServiceURL('GRAFANA'),
  ALERTS: buildServiceURL('ALERT_MANAGER'),
  NODE_METRICS: buildServiceURL('NODE_EXPORTER')
} as const;

// 🎯 端口配置验证
export const validatePortConfiguration = (): { isValid: boolean; errors: string[] } => {
  const errors: string[] = [];
  
  // 检查端口冲突
  const portValues = Object.values(SERVICE_PORTS);
  const duplicates = portValues.filter((port, index) => portValues.indexOf(port) !== index);
  
  if (duplicates.length > 0) {
    errors.push(`端口冲突检测到: ${duplicates.join(', ')}`);
  }
  
  // 检查端口范围
  const invalidPorts = portValues.filter(port => port < 1024 || port > 65535);
  if (invalidPorts.length > 0) {
    errors.push(`无效端口范围: ${invalidPorts.join(', ')} (有效范围: 1024-65535)`);
  }
  
  return {
    isValid: errors.length === 0,
    errors
  };
};

// 🎯 开发工具：端口配置报告
export const generatePortConfigReport = (): string => {
  const validation = validatePortConfiguration();
  
  return [
    '🎯 端口配置报告',
    '====================',
    '',
    '🏗️ 核心服务:',
    `  前端开发服务器: ${SERVICE_PORTS.FRONTEND_DEV}`,
    `  REST命令服务: ${SERVICE_PORTS.REST_COMMAND_SERVICE}`,
    `  GraphQL查询服务: ${SERVICE_PORTS.GRAPHQL_QUERY_SERVICE}`,
    '',
    '🛠️ 基础设施:',
    `  PostgreSQL: ${SERVICE_PORTS.POSTGRESQL}`,
    `  Redis: ${SERVICE_PORTS.REDIS}`,
    '',
    '📊 监控服务:',
    `  Prometheus: ${SERVICE_PORTS.PROMETHEUS}`,
    `  Grafana: ${SERVICE_PORTS.GRAFANA}`,
    '',
    '🔍 配置验证:',
    `  状态: ${validation.isValid ? '✅ 通过' : '❌ 失败'}`,
    ...(validation.errors.map(error => `  错误: ${error}`)),
    '',
    '🎯 CQRS端点:',
    `  命令API: ${CQRS_ENDPOINTS.COMMAND_API}`,
    `  GraphQL查询: ${CQRS_ENDPOINTS.GRAPHQL_ENDPOINT}`,
    ''
  ].join('\n');
};

// 🔒 类型安全导出
export type ServicePortKey = keyof typeof SERVICE_PORTS;
export type CQRSEndpointKey = keyof typeof CQRS_ENDPOINTS;

// 📋 开发提醒
if (process.env.NODE_ENV === 'development') {
  console.log('🎯 端口配置已加载 - 使用统一配置，严禁硬编码端口');
  console.log(`📊 核心服务: REST(${SERVICE_PORTS.REST_COMMAND_SERVICE}) + GraphQL(${SERVICE_PORTS.GRAPHQL_QUERY_SERVICE})`);
}