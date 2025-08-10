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
  unit_type: string;
  status: string;
  description: string;
  parent_code: string;
  level: number;
  sort_order: number;
  // 时态字段
  is_temporal?: boolean;
  effective_from?: string;
  effective_to?: string;
  change_reason?: string;
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
  unit_type: (value: string) => string | null;
  // 时态验证规则
  effective_from: (value: string, isTemporal: boolean) => string | null;
  effective_to: (value: string, effectiveFrom: string, isTemporal: boolean) => string | null;
  change_reason: (value: string, isTemporal: boolean) => string | null;
}