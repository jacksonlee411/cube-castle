import type { OrganizationComponentProps, OrganizationRequest } from '../../../../shared/types';
import type { TemporalMode } from '../../../../shared/types/temporal';

export interface OrganizationFormProps extends Pick<OrganizationComponentProps, 'organization' | 'mode' | 'onSubmit' | 'onCancel' | 'initialData' | 'temporalMode'> {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: OrganizationRequest) => void;
  // 时态相关属性
  isHistorical?: boolean;
  enableTemporalFeatures?: boolean;
}

// 表单数据接口使用统一的OrganizationRequest
export type FormData = OrganizationRequest;

export interface FormFieldsProps {
  formData: FormData;
  setFormData: (data: FormData) => void;
  isEditing: boolean;
  // 时态相关属性
  temporalMode?: TemporalMode;
  enableTemporalFeatures?: boolean;
}

export interface ValidationRules {
  name: (value: string) => string | null;
  code: (value: string) => string | null;
  level: (value: number) => string | null;
  unitType: (value: string) => string | null;  // camelCase
  // 时态验证规则 (camelCase)
  effectiveFrom: (value: string, isTemporal: boolean) => string | null;  // camelCase
  effectiveTo: (value: string, effectiveFrom: string, isTemporal: boolean) => string | null;  // camelCase
  changeReason: (value: string, isTemporal: boolean) => string | null;  // camelCase
}