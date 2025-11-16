/**
 * E2E测试端口配置
 * 统一的测试端口管理，避免硬编码
 */
import { CQRS_ENDPOINTS, SERVICE_PORTS } from '../../src/shared/config/ports';

// 🎯 测试环境端点配置
export const TEST_ENDPOINTS = {
  // 前端应用
  FRONTEND: `http://localhost:${SERVICE_PORTS.FRONTEND_DEV}`,
  
  // 后端服务 (直连，不通过代理)
  REST_COMMAND: CQRS_ENDPOINTS.COMMAND_BASE,
  // 运行时已合流：GraphQL 由单体进程 (:9090) 提供
  GRAPHQL_QUERY: CQRS_ENDPOINTS.QUERY_BASE,
  GRAPHQL_ENDPOINT: CQRS_ENDPOINTS.GRAPHQL_ENDPOINT,
  
  // API端点
  ORGANIZATIONS_API: `${CQRS_ENDPOINTS.COMMAND_API}/organization-units`,
  AUTH_API: `${CQRS_ENDPOINTS.COMMAND_BASE}/auth`,
  
  // 监控端点
  METRICS_COMMAND: `${CQRS_ENDPOINTS.COMMAND_BASE}/metrics`,
  METRICS_QUERY: `${CQRS_ENDPOINTS.QUERY_BASE}/metrics`
} as const;

// 🎯 测试用端口列表 (用于健康检查)
export const TEST_SERVICE_PORTS = [
  // GraphQL 已由单体进程提供，仅检查 9090
  SERVICE_PORTS.REST_COMMAND_SERVICE     // 9090
] as const;

// 🎯 端口可用性检查
export const checkPortAvailability = async (port: number): Promise<boolean> => {
  try {
    const response = await fetch(`http://localhost:${port}/health`, {
      method: 'GET',
      headers: { 'Accept': 'application/json' }
    });
    return response.ok;
  } catch {
    return false;
  }
};

// 🎯 等待服务启动
export const waitForServices = async (timeoutMs: number = 30000): Promise<boolean> => {
  const startTime = Date.now();
  
  while (Date.now() - startTime < timeoutMs) {
    const results = await Promise.all(
      TEST_SERVICE_PORTS.map(port => checkPortAvailability(port))
    );
    
    if (results.every(Boolean)) {
      return true;
    }
    
    await new Promise(resolve => setTimeout(resolve, 1000));
  }
  
  return false;
};
