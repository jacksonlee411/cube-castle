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
  },
  
  unit_type: (value: string) => {
    if (!value || !value.trim()) return '请选择组织类型';
    const validTypes = ['DEPARTMENT', 'COST_CENTER', 'COMPANY', 'PROJECT_TEAM'];
    if (!validTypes.includes(value)) return '无效的组织类型';
    return null;
  }
};

export const validateForm = (formData: Record<string, unknown>, isEditing = false): Record<string, string> => {
  const errors: Record<string, string> = {};
  
  const nameError = validationRules.name(formData['name'] as string);
  if (nameError) errors['name'] = nameError;
  
  // 编辑模式下不验证code，因为code字段被禁用
  if (!isEditing) {
    const codeError = validationRules.code(formData['code'] as string);
    if (codeError) errors['code'] = codeError;
  }
  
  const levelError = validationRules.level(formData['level'] as number);
  if (levelError) errors['level'] = levelError;
  
  // 编辑模式下也需要验证unit_type
  const unitTypeError = validationRules.unit_type(formData['unit_type'] as string);
  if (unitTypeError) errors['unit_type'] = unitTypeError;
  
  return errors;
};