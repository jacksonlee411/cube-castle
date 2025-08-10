import type { OrganizationUnit, OrganizationStatus } from '../../../../shared/types';
import type { TemporalMode } from '../../../../shared/types/temporal';

export interface OrganizationTableProps {
  organizations: OrganizationUnit[];
  onEdit?: (org: OrganizationUnit) => void;
  onToggleStatus?: (code: string, currentStatus: OrganizationStatus) => void;
  loading?: boolean;
  togglingId?: string | undefined;
  // 时态相关属性
  temporalMode?: TemporalMode;
  isHistorical?: boolean;
  showTemporalInfo?: boolean;
}

export interface OrganizationTableRowProps {
  organization: OrganizationUnit;
  onEdit?: (org: OrganizationUnit) => void;
  onToggleStatus?: (code: string, currentStatus: OrganizationStatus) => void;
  isToggling: boolean;
  isAnyToggling: boolean;
  // 时态相关属性
  temporalMode?: TemporalMode;
  isHistorical?: boolean;
  showTemporalInfo?: boolean;
}

export interface TableActionsProps {
  organization: OrganizationUnit;
  onEdit?: (org: OrganizationUnit) => void;
  onToggleStatus?: (code: string, currentStatus: OrganizationStatus) => void;
  isToggling: boolean;
  disabled: boolean;
  // 时态相关属性
  isHistorical?: boolean;
}