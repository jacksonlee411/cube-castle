import type { OrganizationUnit } from '../../../../shared/types';
import type { TemporalMode } from '../../../../shared/types/temporal';

export interface OrganizationTableProps {
  organizations: OrganizationUnit[];
  onTemporalManage?: (code: string) => void | undefined;
  loading?: boolean;
  // 时态相关属性
  temporalMode?: TemporalMode;
  isHistorical?: boolean;
  showTemporalInfo?: boolean;
}

export interface OrganizationTableRowProps {
  organization: OrganizationUnit;
  onTemporalManage?: (code: string) => void;
  isAnyToggling: boolean;
  // 时态相关属性
  temporalMode?: TemporalMode;
  isHistorical?: boolean;
  showTemporalInfo?: boolean;
}

export interface TableActionsProps {
  organization: OrganizationUnit;
  onTemporalManage?: (code: string) => void;
  disabled: boolean;
  // 时态相关属性
  isHistorical?: boolean;
}