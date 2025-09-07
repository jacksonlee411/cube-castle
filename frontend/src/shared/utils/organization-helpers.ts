/**
 * 组织单元相关的工具函数
 * 专门处理parentCode等字段的null值转换问题
 */

/**
 * 安全地处理parentCode字段
 * API要求：根级组织应发送null，而不是空字符串
 * 
 * @param parentCode - 可能是null、undefined或字符串
 * @returns 用于表单显示的字符串，或用于API的null值
 */
export const normalizeParentCode = {
  /**
   * 用于表单显示：将null/undefined转为空字符串
   * @param value - API返回的parentCode值
   * @returns 用于表单输入的字符串
   */
  forForm: (value: string | null | undefined): string => {
    return value || '';
  },

  /**
   * 用于API发送：将空字符串转为null
   * @param value - 表单输入的parentCode值
   * @returns 符合API规范的值（null或有效代码）
   */
  forAPI: (value: string | null | undefined): string | null => {
    if (!value || value.trim() === '') {
      return null; // 根级组织使用null
    }
    return value.trim();
  }
};

/**
 * 判断是否为根级组织
 * @param parentCode - 父组织代码
 * @returns 是否为根级组织
 */
export const isRootOrganization = (parentCode: string | null | undefined): boolean => {
  return !parentCode || parentCode.trim() === '';
};

/**
 * 获取组织层级显示文本
 * @param parentCode - 父组织代码
 * @returns 层级描述
 */
export const getOrganizationLevelText = (parentCode: string | null | undefined): string => {
  return isRootOrganization(parentCode) ? '根级组织' : '子级组织';
};