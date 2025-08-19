import type { OrganizationUnit, OrganizationStatus } from '../../../../shared/types';
import type { TemporalMode } from '../../../../shared/types/temporal';

export interface OrganizationTableProps {
  organizations: OrganizationUnit[];
  onToggleStatus?: (code: string, currentStatus: OrganizationStatus) => void | undefined;
  onTemporalManage?: (code: string) => void | undefined;
  loading?: boolean;
  togglingId?: string | undefined;
  // 时态相关属性
  temporalMode?: TemporalMode;
  isHistorical?: boolean;
  showTemporalInfo?: boolean;
}

export interface OrganizationTableRowProps {
  organization: OrganizationUnit;
  onToggleStatus?: (code: string, currentStatus: OrganizationStatus) => void;
  onTemporalManage?: (code: string) => void;
  isToggling: boolean;
  isAnyToggling: boolean;
  // 时态相关属性
  temporalMode?: TemporalMode;
  isHistorical?: boolean;
  showTemporalInfo?: boolean;
}

export interface TableActionsProps {
  organization: OrganizationUnit;
  onToggleStatus?: (code: string, currentStatus: OrganizationStatus) => void;
  onTemporalManage?: (code: string) => void;
  isToggling: boolean;
  disabled: boolean;
  // 时态相关属性
  isHistorical?: boolean;
}