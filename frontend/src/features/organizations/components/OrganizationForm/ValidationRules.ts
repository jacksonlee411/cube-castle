import type { ValidationRules } from './FormTypes';

export const validationRules: ValidationRules = {
  name: (value: string) => {
    if (!value.trim()) return '组织名称不能为空';
    if (value.trim().length > 100) return '组织名称不能超过100个字符';
    return null;
  },
  
  code: (value: string) => {
    if (!value) return null; // 编码为空时由系统生成
    if (!/^\d{7}$/.test(value)) return '组织编码必须为7位数字';
    return null;
  },
  
  level: (value: number) => {
    if (value < 1 || value > 10) return '组织层级必须在1-10之间';
    return null;
  }
};

export const validateForm = (formData: Record<string, unknown>): Record<string, string> => {
  const errors: Record<string, string> = {};
  
  const nameError = validationRules.name(formData['name'] as string);
  if (nameError) errors['name'] = nameError;
  
  const codeError = validationRules.code(formData['code'] as string);
  if (codeError) errors['code'] = codeError;
  
  const levelError = validationRules.level(formData['level'] as number);
  if (levelError) errors['level'] = levelError;
  
  return errors;
};