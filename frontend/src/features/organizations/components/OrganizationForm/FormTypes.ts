import type { OrganizationUnit } from '../../../../shared/types';
import type { CreateOrganizationInput, UpdateOrganizationInput } from '../../../../shared/hooks/useOrganizationMutations';

export interface OrganizationFormProps {
  organization?: OrganizationUnit | undefined;
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: CreateOrganizationInput | UpdateOrganizationInput) => void;
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
}

export interface FormFieldsProps {
  formData: FormData;
  setFormData: (data: FormData) => void;
  isEditing: boolean;
}

export interface ValidationRules {
  name: (value: string) => string | null;
  code: (value: string) => string | null;
  level: (value: number) => string | null;
  unit_type: (value: string) => string | null;
}