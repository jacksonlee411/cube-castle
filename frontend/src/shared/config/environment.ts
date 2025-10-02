/**
 * 环境配置管理器
 * 统一管理所有环境变量，避免硬编码
 * 基于项目配置管理原则
 */

export interface EnvironmentConfig {
  // API端点配置
  apiBaseUrl: string;
  graphqlEndpoint: string;
  
  // 多租户配置
  defaultTenantId: string;
  
  // 认证配置
  authConfig: {
    clientId: string;
    clientSecret: string;
    tokenEndpoint: string;
    mode: 'dev' | 'oidc';
  };
  
  // 开发配置
  isDevelopment: boolean;
  isProduction: boolean;
}

/**
 * 从环境变量或默认值获取配置
 */
function getEnvironmentConfig(): EnvironmentConfig {
  const isDevelopment = import.meta.env.DEV;
  const isProduction = import.meta.env.PROD;
  
  return {
    // API端点配置 - 开发环境使用代理，生产环境使用完整URL
    apiBaseUrl: isDevelopment ? '/api/v1' : (import.meta.env.VITE_API_BASE_URL || '/api/v1'),
    graphqlEndpoint: isDevelopment ? '/graphql' : (import.meta.env.VITE_GRAPHQL_ENDPOINT || '/graphql'),
    
    // 多租户配置 - 支持环境变量覆盖
    defaultTenantId: import.meta.env.VITE_DEFAULT_TENANT_ID || '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
    
    // 认证配置
    authConfig: {
      clientId: import.meta.env.VITE_AUTH_CLIENT_ID || 'dev-client',
      clientSecret: import.meta.env.VITE_AUTH_CLIENT_SECRET || 'dev-secret',
      tokenEndpoint: isDevelopment ? '/auth/dev-token' : (import.meta.env.VITE_AUTH_TOKEN_ENDPOINT || '/auth/dev-token'),
      mode: (import.meta.env.VITE_AUTH_MODE as 'dev' | 'oidc') || (isDevelopment ? 'dev' : 'oidc'),
    },
    
    // 环境标识
    isDevelopment,
    isProduction,
  };
}

// 导出单例配置实例
export const env = getEnvironmentConfig();

// 配置验证器
export const validateEnvironmentConfig = (): void => {
  const requiredVars = [
    { key: 'defaultTenantId', value: env.defaultTenantId },
    { key: 'authConfig.clientId', value: env.authConfig.clientId },
  ];
  
  const missing = requiredVars.filter(({ value }) => !value);
  
  if (missing.length > 0) {
    const missingKeys = missing.map(({ key }) => key).join(', ');
    throw new Error(`Environment configuration missing required variables: ${missingKeys}`);
  }
  
  // 开发环境配置验证
  if (env.isDevelopment) {
    logger.info('[Environment] 开发环境配置已加载:', {
      defaultTenantId: env.defaultTenantId.substring(0, 8) + '...',
      authClientId: env.authConfig.clientId,
      apiBaseUrl: env.apiBaseUrl,
      graphqlEndpoint: env.graphqlEndpoint,
    });
  }
};

// 自动验证配置
if (env.isDevelopment) {
  validateEnvironmentConfig();
}

export default env;
