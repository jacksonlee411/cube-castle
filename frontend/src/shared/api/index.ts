/**
 * 统一API客户端导出 - Phase 1 彻底迁移完成
 * 
 * 🎉 重复代码彻底消除：
 * - ✅ 统一客户端：unified-client (唯一实现)
 * - ❌ 废弃客户端：已彻底删除
 * - 🏗️ CQRS架构：严格查询-命令分离
 */

// 🎯 唯一API客户端实现
export * from './unified-client';

// 🔧 支持工具和适配器
export * from './graphql-enterprise-adapter';
export * from './auth';

// 🛡️ 统一错误处理体系 (P1级重复代码消除完成)
export * from './error-handling';
export * from './type-guards';

// 🌟 类型导出
export type { OrganizationQueryParams } from '../types/organization';

/**
 * 🚀 统一API使用指南:
 * 
 * 查询操作 (GraphQL):
 * import { unifiedGraphQLClient } from '@/shared/api';
 * const data = await unifiedGraphQLClient.request(QUERY, variables);
 * 
 * 命令操作 (REST):
 * import { unifiedRESTClient } from '@/shared/api';
 * const result = await unifiedRESTClient.request('/endpoint', options);
 * 
 * CQRS原则：
 * - 所有查询 → unifiedGraphQLClient (端口8090)
 * - 所有命令 → unifiedRESTClient (端口9090)
 */