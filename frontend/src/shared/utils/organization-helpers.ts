/**
 * 组织单元相关的工具函数
 * 专门处理parentCode等字段的null值转换问题
 */

export const ROOT_PARENT_CODE = '0000000'
const LEGACY_ROOT_PARENT_CODES = new Set([ROOT_PARENT_CODE, '0'])

const isRootParentCode = (value: string | null | undefined): boolean => {
  if (value == null) {
    return true
  }
  const trimmed = value.trim()
  if (trimmed === '') {
    return true
  }
  return LEGACY_ROOT_PARENT_CODES.has(trimmed)
}

/**
 * 安全地处理parentCode字段
 * API要求：根级组织使用字符串"0000000"，子组织使用7位数字编码
 *
 * @param parentCode - 可能是null、undefined或字符串
 * @returns 用于表单显示的字符串，或用于API的字符串值
 */
export const normalizeParentCode = {
  /**
   * 用于表单显示：将null/undefined/"0"转为合适的表单值
   * @param value - API返回的parentCode值
   * @returns 用于表单输入的字符串
   */
  forForm: (value: string | null | undefined): string => {
    if (isRootParentCode(value)) {
      return ROOT_PARENT_CODE; // 根组织在表单中显示为规范编码
    }
    return value;
  },

  /**
   * 用于API发送：确保根组织使用"0"
   * @param value - 表单输入的parentCode值
   * @returns 符合API规范的值（"0"或有效的7位代码）
   */
  forAPI: (value: string | null | undefined): string => {
    if (isRootParentCode(value)) {
      return ROOT_PARENT_CODE;
    }
    const trimmed = value.trim();
    if (LEGACY_ROOT_PARENT_CODES.has(trimmed)) {
      return ROOT_PARENT_CODE; // 根级组织使用规范编码
    }
    return trimmed;
  }
};

/**
 * 判断是否为根级组织
 * @param parentCode - 父组织代码
 * @returns 是否为根级组织
 */
export const isRootOrganization = (parentCode: string | null | undefined): boolean => {
  return isRootParentCode(parentCode);
};

/**
 * 获取组织层级显示文本
 * @param parentCode - 父组织代码
 * @returns 层级描述
 */
export const getOrganizationLevelText = (parentCode: string | null | undefined): string => {
  return isRootOrganization(parentCode) ? '根级组织' : '子级组织';
};

/**
 * 将层级字段安全转换为数值。
 *
 * - 支持传入 number / string / null / undefined
 * - 非法值统一回退为 0
 * - 可选地提供备用候选值（例如 hierarchyDepth）
 */
export const coerceOrganizationLevel = (
  levelValue: number | string | null | undefined,
  fallbackValue?: number | string | null | undefined
): number => {
  const candidates = [levelValue, fallbackValue];

  for (const candidate of candidates) {
    if (candidate === null || candidate === undefined) {
      continue;
    }

    const parsed = typeof candidate === 'number' ? candidate : Number(candidate);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }

  return 0;
};

/**
 * 计算用于界面展示的层级。
 *
 * 默认将存储层级视为 0 起始，可通过 `offset` 调整为 1 起始显示。
 */
export const getDisplayLevel = (level: number, offset: number = 0): number => {
  const parsed = Number.isFinite(level) ? level : 0;
  return parsed + offset;
};
