import type { OrganizationUnit, OrganizationComponentProps } from '../../../../shared/types';
import type { TemporalMode } from '../../../../shared/types/temporal';

// 表格组件Props使用统一的组件Props接口
export interface OrganizationTableProps extends Pick<OrganizationComponentProps, 'organizations' | 'loading' | 'onSelect' | 'onEdit' | 'onDelete' | 'temporalMode' | 'className'> {
  onTemporalManage?: (code: string) => void | undefined;
  // 时态相关属性
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