/**
 * 统一路由配置 - Cube Castle 企业管理系统
 * 
 * 用途：
 * 1. 集中管理所有API端点配置
 * 2. 统一环境变量使用
 * 3. 提供类型安全的路由常量
 * 4. 支持不同服务的路由分组
 */

// ========== 环境变量配置 ==========

// 主要API服务基础URL
export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

// AI服务基础URL (gRPC Gateway)
export const AI_API_URL = process.env.NEXT_PUBLIC_AI_API_URL || 'http://localhost:8081'

// 默认租户ID (多租户支持)
export const DEFAULT_TENANT_ID = process.env.NEXT_PUBLIC_DEFAULT_TENANT_ID || '00000000-0000-0000-0000-000000000001'

// 默认请求超时时间
export const DEFAULT_TIMEOUT = {
  STANDARD: 10000,    // 10秒 - 标准API请求
  AI_SERVICE: 15000,  // 15秒 - AI服务请求
  UPLOAD: 30000,      // 30秒 - 文件上传
} as const

// ========== API路由常量 ==========

/**
 * CQRS架构路由配置
 * 分离命令(Command)和查询(Query)端点
 */
export const CQRS_ROUTES = {
  // 员工 CQRS 路由
  EMPLOYEE: {
    // 查询端点 (Query Side)
    QUERIES: {
      SEARCH: '/api/v1/queries/employees',
      GET_BY_ID: (id: string) => `/api/v1/queries/employees/${id}`,
      STATS: '/api/v1/queries/employees/stats',
    },
    // 命令端点 (Command Side)
    COMMANDS: {
      HIRE: '/api/v1/commands/hire-employee',
      UPDATE: '/api/v1/commands/update-employee',
      TERMINATE: '/api/v1/commands/terminate-employee',
    },
  },

  // 组织架构 CQRS 路由
  ORGANIZATION: {
    // 查询端点
    QUERIES: {
      LIST: '/api/v1/queries/organizations',
      GET_BY_ID: (id: string) => `/api/v1/queries/organizations/${id}`,
      STATS: '/api/v1/queries/organizations/stats',
      HIERARCHY: '/api/v1/queries/organizations/hierarchy',
    },
    // 命令端点
    COMMANDS: {
      CREATE: '/api/v1/commands/create-organization',
      UPDATE: '/api/v1/commands/update-organization',
      DELETE: '/api/v1/commands/delete-organization',
    },
  },
} as const

/**
 * 传统REST API路由配置 (正在迁移到CQRS)
 * 保留用于向后兼容和渐进式迁移
 */
export const REST_ROUTES = {
  // CoreHR适配器API (与飞书等外部系统集成)
  COREHR: {
    EMPLOYEES: '/api/v1/corehr/employees',
    EMPLOYEE_BY_ID: (id: string) => `/api/v1/corehr/employees/${id}`,
    ORGANIZATIONS: '/api/v1/corehr/organizations',
    ORGANIZATION_BY_ID: (id: string) => `/api/v1/corehr/organizations/${id}`,
    ORGANIZATION_STATS: '/api/v1/corehr/organizations/stats',
  },

  // 系统管理API
  SYSTEM: {
    HEALTH: '/api/v1/system/health',
    INFO: '/api/v1/system/info',
    METRICS: '/api/v1/system/metrics/business',
  },

  // 工作流API
  WORKFLOWS: {
    INSTANCES: '/api/v1/workflows/instances',
    INSTANCE_BY_ID: (id: string) => `/api/v1/workflows/instances/${id}`,
    START: '/api/v1/workflows/start',
    STATS: '/api/v1/workflows/stats',
  },
} as const

/**
 * AI服务路由配置
 */
export const AI_ROUTES = {
  INTELLIGENCE: {
    INTERPRET: '/api/v1/intelligence/interpret',
    CONVERSATION_HISTORY: (sessionId: string) => `/api/v1/intelligence/conversations/${sessionId}`,
  },
} as const

/**
 * 内部API路由 (Next.js API Routes)
 */
export const INTERNAL_ROUTES = {
  // Next.js API Routes
  EMPLOYEES: '/api/employees',
  ORGANIZATIONS: '/api/organizations',
  HEALTH: '/api/health',
} as const

// ========== 路由构建工具函数 ==========

/**
 * 构建完整的API URL
 * @param endpoint - API端点路径
 * @param baseUrl - 基础URL (可选，默认使用API_BASE_URL)
 */
export function buildApiUrl(endpoint: string, baseUrl: string = API_BASE_URL): string {
  // 确保baseUrl没有尾部斜杠
  const cleanBaseUrl = baseUrl.replace(/\/$/, '')
  // 确保endpoint有前导斜杠
  const cleanEndpoint = endpoint.startsWith('/') ? endpoint : `/${endpoint}`
  
  return `${cleanBaseUrl}${cleanEndpoint}`
}

/**
 * 构建带查询参数的URL
 * @param endpoint - API端点路径
 * @param params - 查询参数对象
 * @param baseUrl - 基础URL (可选)
 */
export function buildUrlWithParams(
  endpoint: string, 
  params: Record<string, any> = {}, 
  baseUrl: string = API_BASE_URL
): string {
  const url = buildApiUrl(endpoint, baseUrl)
  
  // 过滤掉undefined和null值
  const filteredParams = Object.entries(params)
    .filter(([_, value]) => value !== undefined && value !== null)
    .reduce((acc, [key, value]) => {
      acc[key] = String(value)
      return acc
    }, {} as Record<string, string>)
  
  const searchParams = new URLSearchParams(filteredParams)
  const queryString = searchParams.toString()
  
  return queryString ? `${url}?${queryString}` : url
}

/**
 * 获取员工CQRS查询URL
 */
export function getEmployeeQueryUrl(operation: 'search' | 'stats', employeeId?: string): string {
  switch (operation) {
    case 'search':
      return buildApiUrl(CQRS_ROUTES.EMPLOYEE.QUERIES.SEARCH)
    case 'stats':
      return buildApiUrl(CQRS_ROUTES.EMPLOYEE.QUERIES.STATS)
    default:
      if (employeeId) {
        return buildApiUrl(CQRS_ROUTES.EMPLOYEE.QUERIES.GET_BY_ID(employeeId))
      }
      throw new Error('Employee ID required for single employee query')
  }
}

/**
 * 获取员工CQRS命令URL
 */
export function getEmployeeCommandUrl(operation: 'hire' | 'update' | 'terminate'): string {
  switch (operation) {
    case 'hire':
      return buildApiUrl(CQRS_ROUTES.EMPLOYEE.COMMANDS.HIRE)
    case 'update':
      return buildApiUrl(CQRS_ROUTES.EMPLOYEE.COMMANDS.UPDATE)
    case 'terminate':
      return buildApiUrl(CQRS_ROUTES.EMPLOYEE.COMMANDS.TERMINATE)
    default:
      throw new Error(`Unknown employee command operation: ${operation}`)
  }
}

/**
 * 获取组织CQRS查询URL
 */
export function getOrganizationQueryUrl(operation: 'list' | 'stats' | 'hierarchy', orgId?: string): string {
  switch (operation) {
    case 'list':
      return buildApiUrl(CQRS_ROUTES.ORGANIZATION.QUERIES.LIST)
    case 'stats':
      return buildApiUrl(CQRS_ROUTES.ORGANIZATION.QUERIES.STATS)
    case 'hierarchy':
      return buildApiUrl(CQRS_ROUTES.ORGANIZATION.QUERIES.HIERARCHY)
    default:
      if (orgId) {
        return buildApiUrl(CQRS_ROUTES.ORGANIZATION.QUERIES.GET_BY_ID(orgId))
      }
      throw new Error('Organization ID required for single organization query')
  }
}

// ========== 环境检测工具 ==========

/**
 * 检测当前运行环境
 */
export function getEnvironment(): 'development' | 'production' | 'test' {
  if (typeof window === 'undefined') {
    // 服务端环境
    return process.env.NODE_ENV as any || 'development'
  }
  
  // 客户端环境检测
  const hostname = window.location.hostname
  
  if (hostname === 'localhost' || hostname === '127.0.0.1') {
    return 'development'
  }
  
  if (hostname.includes('test') || hostname.includes('staging')) {
    return 'test'
  }
  
  return 'production'
}

/**
 * 是否为开发环境
 */
export function isDevelopment(): boolean {
  return getEnvironment() === 'development'
}

/**
 * 是否为生产环境
 */
export function isProduction(): boolean {
  return getEnvironment() === 'production'
}

// ========== 路由验证工具 ==========

/**
 * 验证API端点是否有效
 */
export function validateEndpoint(endpoint: string): boolean {
  // 基本格式验证
  if (!endpoint || typeof endpoint !== 'string') {
    return false
  }
  
  // 必须以/api开头
  if (!endpoint.startsWith('/api')) {
    return false
  }
  
  // 不能包含危险字符
  const dangerousChars = ['..', '<', '>', '"', "'", '&']
  if (dangerousChars.some(char => endpoint.includes(char))) {
    return false
  }
  
  return true
}

/**
 * 获取API版本信息
 */
export function getApiVersionFromEndpoint(endpoint: string): string | null {
  const versionMatch = endpoint.match(/\/api\/v(\d+)\//)
  return versionMatch ? `v${versionMatch[1]}` : null
}

// ========== 导出所有路由配置 ==========

export const ROUTES = {
  CQRS: CQRS_ROUTES,
  REST: REST_ROUTES,
  AI: AI_ROUTES,
  INTERNAL: INTERNAL_ROUTES,
} as const

// 类型导出
export type CQRSRoutes = typeof CQRS_ROUTES
export type RESTRoutes = typeof REST_ROUTES
export type AIRoutes = typeof AI_ROUTES
export type InternalRoutes = typeof INTERNAL_ROUTES

// 默认导出
export default {
  API_BASE_URL,
  AI_API_URL,
  DEFAULT_TENANT_ID,
  DEFAULT_TIMEOUT,
  ROUTES,
  buildApiUrl,
  buildUrlWithParams,
  getEmployeeQueryUrl,
  getEmployeeCommandUrl,
  getOrganizationQueryUrl,
  getEnvironment,
  isDevelopment,
  isProduction,
  validateEndpoint,
  getApiVersionFromEndpoint,
}