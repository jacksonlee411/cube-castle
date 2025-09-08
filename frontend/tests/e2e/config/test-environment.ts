/**
 * E2E测试环境配置
 * 🎯 解决端口硬编码问题：动态端口发现与环境变量支持
 * 📋 遵循06号文档P1任务要求
 */

import { SERVICE_PORTS, buildServiceURL } from '../../../src/shared/config/ports';

// 🎯 E2E测试基础URL配置
export const E2E_CONFIG = {
  // 主要测试目标
  FRONTEND_BASE_URL: process.env.E2E_BASE_URL || buildServiceURL('FRONTEND_DEV'),
  
  // 后端服务端点
  COMMAND_API_URL: process.env.E2E_COMMAND_API_URL || buildServiceURL('REST_COMMAND_SERVICE', '/api/v1'),
  GRAPHQL_API_URL: process.env.E2E_GRAPHQL_API_URL || buildServiceURL('GRAPHQL_QUERY_SERVICE', '/graphql'),
  
  // 超时配置
  PAGE_TIMEOUT: parseInt(process.env.E2E_PAGE_TIMEOUT || '30000'),
  NAVIGATION_TIMEOUT: parseInt(process.env.E2E_NAVIGATION_TIMEOUT || '15000'),
  
  // 服务等待配置
  SERVICE_STARTUP_WAIT: parseInt(process.env.E2E_SERVICE_WAIT || '5000'),
  
  // 调试模式
  DEBUG_MODE: process.env.E2E_DEBUG === 'true',
} as const;

// 🎯 端口可用性检测
export const checkPortAvailability = async (port: number, host: string = 'localhost'): Promise<boolean> => {
  try {
    const response = await fetch(`http://${host}:${port}/health`, {
      method: 'GET',
      timeout: 3000,
    });
    return response.ok;
  } catch (error) {
    if (E2E_CONFIG.DEBUG_MODE) {
      console.log(`Port ${port} not available: ${error}`);
    }
    return false;
  }
};

// 🎯 动态端口发现
export const discoverActivePort = async (basePorts: number[] = [3000, 3001, 3002]): Promise<string | null> => {
  for (const port of basePorts) {
    try {
      const response = await fetch(`http://localhost:${port}`, {
        method: 'GET',
        timeout: 2000,
      });
      if (response.ok) {
        console.log(`✅ 发现活跃前端服务：http://localhost:${port}`);
        return `http://localhost:${port}`;
      }
    } catch (error) {
      // 继续尝试下一个端口
    }
  }
  
  console.warn('⚠️ 未发现活跃的前端服务，使用默认配置');
  return null;
};

// 🎯 测试环境验证
export const validateTestEnvironment = async (): Promise<{
  isValid: boolean;
  errors: string[];
  frontendUrl: string;
}> => {
  const errors: string[] = [];
  let frontendUrl = E2E_CONFIG.FRONTEND_BASE_URL;
  
  // 动态端口发现
  if (!process.env.E2E_BASE_URL) {
    const discoveredUrl = await discoverActivePort();
    if (discoveredUrl) {
      frontendUrl = discoveredUrl;
    }
  }
  
  // 检查前端服务可用性
  try {
    const frontendAvailable = await checkPortAvailability(
      parseInt(frontendUrl.split(':').pop()!), 
      'localhost'
    );
    if (!frontendAvailable) {
      errors.push(`前端服务不可用: ${frontendUrl}`);
    }
  } catch (error) {
    errors.push(`前端服务检查失败: ${frontendUrl}`);
  }
  
  return {
    isValid: errors.length === 0,
    errors,
    frontendUrl
  };
};

// 🎯 测试配置报告
export const generateTestConfigReport = (): string => {
  return [
    '🎯 E2E测试环境配置报告',
    '========================',
    '',
    '🏗️ 服务端点配置:',
    `  前端基址: ${E2E_CONFIG.FRONTEND_BASE_URL}`,
    `  命令API: ${E2E_CONFIG.COMMAND_API_URL}`,
    `  GraphQL API: ${E2E_CONFIG.GRAPHQL_API_URL}`,
    '',
    '⏱️ 超时配置:',
    `  页面超时: ${E2E_CONFIG.PAGE_TIMEOUT}ms`,
    `  导航超时: ${E2E_CONFIG.NAVIGATION_TIMEOUT}ms`,
    `  服务启动等待: ${E2E_CONFIG.SERVICE_STARTUP_WAIT}ms`,
    '',
    '🔍 环境变量:',
    `  E2E_BASE_URL: ${process.env.E2E_BASE_URL || '未设置'}`,
    `  E2E_DEBUG: ${process.env.E2E_DEBUG || '未设置'}`,
    `  E2E_PAGE_TIMEOUT: ${process.env.E2E_PAGE_TIMEOUT || '未设置'}`,
    '',
    '📋 端口配置来源:',
    `  前端开发端口: ${SERVICE_PORTS.FRONTEND_DEV}`,
    `  命令服务端口: ${SERVICE_PORTS.REST_COMMAND_SERVICE}`,
    `  查询服务端口: ${SERVICE_PORTS.GRAPHQL_QUERY_SERVICE}`,
    ''
  ].join('\n');
};

// 🔒 类型安全导出
export type E2EConfigKey = keyof typeof E2E_CONFIG;

// 📋 开发提醒
if (process.env.NODE_ENV === 'development' && E2E_CONFIG.DEBUG_MODE) {
  console.log('🎯 E2E测试环境配置已加载');
  console.log(generateTestConfigReport());
}