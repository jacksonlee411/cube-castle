import type { OrganizationUnit } from '../../../../shared/types';

export interface OrganizationTableProps {
  organizations: OrganizationUnit[];
  onEdit: (org: OrganizationUnit) => void;
  onDelete: (code: string) => void;
  loading?: boolean;
  deletingId?: string | undefined;
}

export interface OrganizationTableRowProps {
  organization: OrganizationUnit;
  onEdit: (org: OrganizationUnit) => void;
  onDelete: (code: string) => void;
  isDeleting: boolean;
  isAnyDeleting: boolean;
}

export interface TableActionsProps {
  organization: OrganizationUnit;
  onEdit: (org: OrganizationUnit) => void;
  onDelete: (code: string) => void;
  isDeleting: boolean;
  disabled: boolean;
}