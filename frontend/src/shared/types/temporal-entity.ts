/**
 * Plan 245 – Temporal Entity 统一类型
 * 单一事实来源：前端类型由此导出，组织/职位仅作为别名或特化。
 *
 * 注意：
 * - 不改变现有 OrganizationUnit / PositionRecord 的定义与文件结构，避免一次性破坏性变更
 * - 提供最小公共形状与别名，以便后续 codemod 渐进收敛
 * - 所有新增引用建议从本文件导入，旧引用逐步迁移（见 Plan 245 清单）
 *
 * // TODO-TEMPORARY(Plan 245, Day 8-12):
 * - 收敛消费端到本统一类型，并清理 PositionDetailQuery/OrganizationUnit 等旧命名的直接使用
 * - 与 docs/api/schema.graphql、docs/api/openapi.yaml 命名统一（TemporalEntity*）
 */

import type { OrganizationUnit } from './organization';
import type { OrganizationStatus } from './contract_gen';
import type { PositionRecord, PositionStatus } from './positions';
import type { JsonValue } from './json';

// 统一的时态实体类型标识
export type TemporalEntityType = 'organization' | 'position';

// 统一状态命名（组织/职位特化的并集，兼容扩展）
export type TemporalEntityStatus = OrganizationStatus | PositionStatus | string;

// 统一的时态实体“记录”基础形状
export interface TemporalEntityRecord {
  entityType: TemporalEntityType;
  code: string;
  recordId?: string | null;
  // 统一显示名称字段（组织使用 name，职位使用 title）
  displayName?: string | null;
  // 组织/职位特定扩展，保持可选以兼容最小公共子集
  organizationCode?: string | null;
  organizationName?: string | null;
  status?: TemporalEntityStatus;
  effectiveDate?: string | null;
  endDate?: string | null;
  // 允许附加扩展字段，避免重复定义
  profile?: Record<string, JsonValue>;
}

// 统一的时间线条目（与 Plan 244 Timeline 命名一致）
export interface TemporalEntityTimelineEntry {
  recordId: string;
  status: TemporalEntityStatus;
  effectiveDate: string;
  endDate?: string | null;
  isCurrent: boolean;
  changeReason?: string | null;
  // 允许实体特化字段
  category?: string; // 如 POSITION_VERSION/POSITION_ASSIGNMENT
  details?: Record<string, JsonValue>;
}

// 渐进式别名：组织/职位映射到统一实体概念（不改变原类型定义）
export type OrganizationEntityRecord = OrganizationUnit;
export type PositionEntityRecord = PositionRecord;

// 工具类型：结果统一封装
export interface TemporalEntityDetail<TRecord extends TemporalEntityRecord = TemporalEntityRecord> {
  record: TRecord | null;
  versions?: TRecord[]; // 若消费端需要“版本列表”，保留可选
  timeline?: TemporalEntityTimelineEntry[];
}
