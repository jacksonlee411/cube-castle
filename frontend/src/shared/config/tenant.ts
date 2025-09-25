/**
 * 前端租户配置统一管理
 * 消除硬编码租户ID，支持多租户架构
 * 对应后端 shared/config/tenant.go
 */

// 默认租户配置常量
export const DEFAULT_TENANT_CONFIG = {
  // 默认租户ID - 与后端保持一致
  id: '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
  name: '高谷集团',
  code: 'GAOYAGU',
} as const;

/**
 * 租户配置管理类
 */
export class TenantManager {
  private static instance: TenantManager;
  private currentTenantId: string;

  private constructor() {
    // 从环境变量或JWT token中获取租户ID，默认使用默认租户
    this.currentTenantId = this.loadTenantId();
  }

  public static getInstance(): TenantManager {
    if (!TenantManager.instance) {
      TenantManager.instance = new TenantManager();
    }
    return TenantManager.instance;
  }

  /**
   * 获取当前租户ID
   */
  public getCurrentTenantId(): string {
    return this.currentTenantId;
  }

  /**
   * 设置当前租户ID
   * @param tenantId 租户ID
   */
  public setCurrentTenantId(tenantId: string): void {
    if (!this.isValidUUID(tenantId)) {
      throw new Error(`Invalid tenant ID format: ${tenantId}`);
    }
    this.currentTenantId = tenantId;
    // 存储到localStorage以便持久化
    localStorage.setItem('cube-castle-tenant-id', tenantId);
  }

  /**
   * 检查是否为默认租户
   */
  public isDefaultTenant(tenantId?: string): boolean {
    const id = tenantId || this.currentTenantId;
    return id === DEFAULT_TENANT_CONFIG.id;
  }

  /**
   * 获取租户配置信息
   */
  public getTenantConfig(tenantId?: string) {
    const id = tenantId || this.currentTenantId;
    
    if (this.isDefaultTenant(id)) {
      return {
        id: DEFAULT_TENANT_CONFIG.id,
        name: DEFAULT_TENANT_CONFIG.name,
        code: DEFAULT_TENANT_CONFIG.code,
      };
    }
    
    // 对于非默认租户，可以通过API获取配置
    return {
      id,
      name: 'Unknown Tenant',
      code: 'UNKNOWN',
    };
  }

  /**
   * 从多个来源加载租户ID
   * 优先级: JWT token > 环境变量 > localStorage > 默认值
   */
  private loadTenantId(): string {
    const tokenTenantId = this.getTenantIdFromToken();
    if (tokenTenantId) {
      return tokenTenantId;
    }

    // 从环境变量读取
    const envTenantId = import.meta.env.VITE_DEFAULT_TENANT_ID;
    if (envTenantId && this.isValidUUID(envTenantId)) {
      return envTenantId;
    }

    // 从localStorage读取
    const storedTenantId = localStorage.getItem('cube-castle-tenant-id');
    if (storedTenantId && this.isValidUUID(storedTenantId)) {
      return storedTenantId;
    }

    // 返回默认租户ID
    return DEFAULT_TENANT_CONFIG.id;
  }

  /**
   * 验证UUID格式
   */
  private isValidUUID(uuid: string): boolean {
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
    return uuidRegex.test(uuid);
  }

  /**
   * 从JWT token（或认证管理器）中获取租户ID。
   * 当前实现返回 null，保留扩展点用于未来集成统一认证。
   */
  private getTenantIdFromToken(): string | null {
    return null;
  }
}

/**
 * 单例租户管理器
 */
export const tenantManager = TenantManager.getInstance();

/**
 * 便捷函数：获取当前租户ID
 */
export const getCurrentTenantId = (): string => {
  return tenantManager.getCurrentTenantId();
};

/**
 * 便捷函数：检查是否为默认租户
 */
export const isDefaultTenant = (tenantId?: string): boolean => {
  return tenantManager.isDefaultTenant(tenantId);
};

/**
 * 便捷函数：获取租户配置
 */
export const getTenantConfig = (tenantId?: string) => {
  return tenantManager.getTenantConfig(tenantId);
};
