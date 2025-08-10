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
  },

  // 时态字段验证规则
  effective_from: (value: string, isTemporal: boolean) => {
    if (!isTemporal) return null;
    if (!value) return '请选择生效时间';
    
    const effectiveDate = new Date(value);
    if (isNaN(effectiveDate.getTime())) return '无效的生效时间格式';
    
    const now = new Date();
    // 允许历史日期，但不允许超过10年前
    const minDate = new Date(now.getFullYear() - 10, 0, 1);
    if (effectiveDate < minDate) return '生效时间不能超过10年前';
    
    // 不允许超过5年后
    const maxDate = new Date(now.getFullYear() + 5, 11, 31);
    if (effectiveDate > maxDate) return '生效时间不能超过5年后';
    
    return null;
  },

  effective_to: (value: string, effectiveFrom: string, isTemporal: boolean) => {
    if (!isTemporal || !value) return null; // 失效时间可以为空
    
    const effectiveToDate = new Date(value);
    if (isNaN(effectiveToDate.getTime())) return '无效的失效时间格式';
    
    if (effectiveFrom) {
      const effectiveFromDate = new Date(effectiveFrom);
      if (effectiveToDate <= effectiveFromDate) {
        return '失效时间必须晚于生效时间';
      }
    }
    
    return null;
  },

  change_reason: (value: string, isTemporal: boolean) => {
    if (!isTemporal) return null;
    if (!value || !value.trim()) return '请输入变更原因';
    if (value.trim().length > 200) return '变更原因不能超过200个字符';
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
  
  // 时态字段验证
  const isTemporal = formData['is_temporal'] as boolean;
  if (isTemporal) {
    const effectiveFromError = validationRules.effective_from(formData['effective_from'] as string, isTemporal);
    if (effectiveFromError) errors['effective_from'] = effectiveFromError;
    
    const effectiveToError = validationRules.effective_to(
      formData['effective_to'] as string, 
      formData['effective_from'] as string, 
      isTemporal
    );
    if (effectiveToError) errors['effective_to'] = effectiveToError;
    
    const changeReasonError = validationRules.change_reason(formData['change_reason'] as string, isTemporal);
    if (changeReasonError) errors['change_reason'] = changeReasonError;
  }
  
  return errors;
};