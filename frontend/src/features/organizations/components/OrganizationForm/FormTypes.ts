import type { OrganizationUnit } from '../../../../shared/types';
import type { CreateOrganizationInput, UpdateOrganizationInput } from '../../../../shared/hooks/useOrganizationMutations';
import type { TemporalMode } from '../../../../shared/types/temporal';

export interface OrganizationFormProps {
  organization?: OrganizationUnit | undefined;
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: CreateOrganizationInput | UpdateOrganizationInput) => void;
  // 时态相关属性
  temporalMode?: TemporalMode;
  isHistorical?: boolean;
  enableTemporalFeatures?: boolean;
}

export interface FormData {
  [key: string]: unknown;
  code?: string | undefined;
  name: string;
  unitType: string;  // camelCase
  status: string;
  description: string;
  parentCode: string;  // camelCase
  level: number;
  sortOrder: number;  // camelCase
  // 时态字段 (camelCase)
  isTemporal?: boolean;  // camelCase
  effectiveFrom?: string;  // camelCase
  effectiveTo?: string;  // camelCase
  changeReason?: string;  // camelCase
}

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