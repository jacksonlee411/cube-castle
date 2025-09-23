import type { FormData } from './FormTypes';

/**
 * 统一清洗表单数据，确保传入共享验证 Schema 前格式正确。
 */
export const prepareFormDataForValidation = (data: FormData): Record<string, unknown> => {
  const normalized: Record<string, unknown> = { ...data };

  if (typeof normalized.code === 'string') {
    const trimmed = normalized.code.trim();
    if (trimmed.length === 0) {
      delete normalized.code;
    } else {
      normalized.code = trimmed;
    }
  }

  if (typeof normalized.parentCode === 'string') {
    normalized.parentCode = normalized.parentCode.trim();
  }

  if (typeof normalized.description === 'string') {
    normalized.description = normalized.description.trim();
  }

  if (typeof normalized.changeReason === 'string') {
    normalized.changeReason = normalized.changeReason.trim();
  }

  return normalized;
};
