/**
 * 统一Hook导出 - Phase 1 Hook实现统一化
 * 
 * 🔥 重要变更：Hook统一化策略
 * - 主要实现：useEnterpriseOrganizations
 * - 兼容包装：useOrganizations, useOrganizationList
 * - 废弃清理：逐步移除feature-specific重复Hook
 */

// 🎯 主要实现：企业级组织Hook (推荐使用)
export * from './useEnterpriseOrganizations';
export { default as useEnterpriseOrganizations } from './useEnterpriseOrganizations';

// 🔄 兼容包装：传统Hook保持向后兼容
export * from './useOrganizations';

// 🔧 工具和支持Hook
export * from './useOrganizationMutations';
export * from './useTemporalAPI';
export * from './useDebounce';

// 🌟 统一导出：统一接口访问点
import useEnterpriseOrganizations from './useEnterpriseOrganizations';

// 创建统一Hook别名，逐步迁移到主要实现
export const useOrganizationList = useEnterpriseOrganizations;

/**
 * 📋 迁移指南:
 * 
 * 推荐使用：
 * - useEnterpriseOrganizations (完整功能)
 * - useOrganizationList (简化接口)
 * 
 * 兼容模式：
 * - useOrganizations (保持向后兼容)
 * 
 * 计划废弃：
 * - features/organizations/hooks/* 中的特定Hook
 */