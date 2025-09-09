/**
 * 统一工具函数导出文件
 * 
 * 将所有工具函数集中导出，避免重复实现和循环导入
 */

// ============================================================================
// 时态工具统一导出 (基于temporal-converter.ts)
// ============================================================================

export {
  TemporalConverter,
  TemporalUtils,
  default as Temporal
} from './temporal-converter';

// ============================================================================
// 废弃导入警告
// ============================================================================

/**
 * 废弃导入警告：
 * 
 * ❌ 不要从以下地方导入时态工具：
 * - features/temporal/utils/temporalValidation.ts
 * - shared/validation/schemas.ts (时态工具部分)
 * - shared/hooks/useTemporalAPI.ts (时态工具部分)
 * 
 * ✅ 统一从此处导入：
 * 
 * import { TemporalConverter, TemporalUtils } from '@/shared/utils';
 * 
 * 优势：
 * - 单一来源，避免重复
 * - 完整功能集合
 * - 类型安全保证
 * - 性能优化
 */

// ============================================================================
// 时态工具快捷导出 (最常用的功能)
// ============================================================================

import { TemporalUtils as Utils } from './temporal-converter';

export const DateUtils = {
  // 当前时间
  now: Utils.now,
  today: Utils.today,
  
  // 格式化
  formatDate: Utils.formatDate,
  formatDateTime: Utils.formatDateTime,
  
  // 转换
  toISO: Utils.toISO,
  toDate: Utils.toDate,
  toDateString: Utils.toDateString,
  
  // 比较
  isSameDate: Utils.isSameDate,
  isInRange: Utils.isInRange,
  
  // 验证
  validate: Utils.validate,
  
  // 标准化
  normalize: Utils.normalize
};

// ============================================================================
// 迁移状态信息
// ============================================================================

export const TEMPORAL_UTILS_INFO = {
  version: '2.0.0-unified',
  consolidatedFiles: [
    'shared/utils/temporal-converter.ts (主文件)',
    'shared/utils/index.ts (统一导出)'
  ],
  deprecatedFiles: [
    'features/temporal/utils/temporalValidation.ts',
    'shared/validation/schemas.ts (时态工具部分)',
    'shared/hooks/useTemporalAPI.ts (时态工具部分)'
  ],
  migrationStatus: 'P1-COMPLETED',
  duplicateCodeEliminated: true,
  performanceOptimized: true
} as const;