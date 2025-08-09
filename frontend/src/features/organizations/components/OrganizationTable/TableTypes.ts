import type { OrganizationUnit, OrganizationStatus } from '../../../../shared/types';

export interface OrganizationTableProps {
  organizations: OrganizationUnit[];
  onEdit: (org: OrganizationUnit) => void;
  onToggleStatus: (code: string, currentStatus: OrganizationStatus) => void;
  loading?: boolean;
  togglingId?: string | undefined;
}

export interface OrganizationTableRowProps {
  organization: OrganizationUnit;
  onEdit: (org: OrganizationUnit) => void;
  onToggleStatus: (code: string, currentStatus: OrganizationStatus) => void;
  isToggling: boolean;
  isAnyToggling: boolean;
}

export interface TableActionsProps {
  organization: OrganizationUnit;
  onEdit: (org: OrganizationUnit) => void;
  onToggleStatus: (code: string, currentStatus: OrganizationStatus) => void;
  isToggling: boolean;
  disabled: boolean;
}