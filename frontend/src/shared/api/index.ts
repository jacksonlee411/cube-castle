/**
 * 统一API客户端导出 - Phase 1 API客户端统一化
 * 
 * 🔥 重要变更：API客户端统一策略 
 * - 主要实现：unified-client (推荐使用)
 * - 兼容导出：保持向后兼容性
 * - 废弃清理：逐步移除重复客户端实现
 */

// 🎯 主要实现：统一API客户端 (强烈推荐)
export * from './unified-client';

// 🔄 兼容导出：保持向后兼容，但将逐步废弃
export * from './organizations';
export * from './organizations-enterprise';
export * from './client';

// 🔧 适配器和工具
export * from './graphql-enterprise-adapter';
export * from './auth';
export * from './error-handling';

// 🌟 类型导出
export type { OrganizationQueryParams } from '../types/organization';

/**
 * 📋 迁移指南:
 * 
 * 推荐使用：
 * - UnifiedGraphQLClient (查询操作)
 * - UnifiedRESTClient (命令操作)
 * 
 * 兼容模式：
 * - organizationAPI (将被废弃)
 * - enterpriseOrganizationAPI (将被废弃)
 * - ApiClient (将被废弃)
 * 
 * CQRS原则：
 * - 查询 → GraphQL客户端
 * - 命令 → REST客户端
 */