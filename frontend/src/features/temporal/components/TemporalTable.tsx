/**
 * 时态感知数据表格组件
 * 支持时态模式的组织架构数据展示和操作
 */
import React, { useState, useMemo, useCallback } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { Table } from '@workday/canvas-kit-react/table';
import { PrimaryButton, SecondaryButton, ToolbarIconButton as IconButton } from '@workday/canvas-kit-react/button';
import { Badge } from '../../../shared/components/Badge';
import { Tooltip } from '@workday/canvas-kit-react/tooltip';
import { Checkbox } from '@workday/canvas-kit-react/checkbox';
import { 
  colors, 
  space, 
  borderRadius 
} from '@workday/canvas-kit-react/tokens';
import {
  EditIcon,
  DeleteIcon,
  HistoryIcon,
  TimelineIcon,
  InfoIcon,
  CompareIcon,
  FilterIcon
} from '@workday/canvas-kit-react/icon';
import { useTemporalOrganizations } from '../../shared/hooks/useTemporalQuery';
import { temporalSelectors } from '../../shared/stores/temporalStore';
import type { OrganizationUnit, OrganizationQueryParams } from '../../shared/types/organization';
import type { TemporalMode } from '../../shared/types/temporal';

export interface TemporalTableProps {
  /** 查询参数 */
  queryParams?: OrganizationQueryParams;
  /** 是否显示时态指示器 */
  showTemporalIndicators?: boolean;
  /** 是否显示操作列 */
  showActions?: boolean;
  /** 是否显示选择列 */
  showSelection?: boolean;
  /** 是否紧凑模式 */
  compact?: boolean;
  /** 每页显示数量 */
  pageSize?: number;
  /** 行点击回调 */
  onRowClick?: (organization: OrganizationUnit) => void;
  /** 编辑回调 */
  onEdit?: (organization: OrganizationUnit) => void;
  /** 删除回调 */
  onDelete?: (organization: OrganizationUnit) => void;
  /** 查看历史回调 */
  onViewHistory?: (organization: OrganizationUnit) => void;
  /** 查看时间线回调 */
  onViewTimeline?: (organization: OrganizationUnit) => void;
  /** 选择变更回调 */
  onSelectionChange?: (selectedOrganizations: OrganizationUnit[]) => void;
}

/**
 * 时态状态指示器组件
 */
interface TemporalIndicatorProps {
  mode: TemporalMode;
  organization: OrganizationUnit;
  compact: boolean;
}

const TemporalIndicator: React.FC<TemporalIndicatorProps> = ({
  mode,
  organization,
  compact
}) => {
  const getIndicatorStyle = () => {
    switch (mode) {
      case 'current':
        return {
          color: colors.greenFresca600,
          bgColor: colors.greenFresca100,
          label: '当前',
          icon: '🟢'
        };
      case 'historical':
        return {
          color: colors.blueberry600,
          bgColor: colors.blueberry100,
          label: '历史',
          icon: '🔵'
        };
      case 'planning':
        return {
          color: colors.peach600,
          bgColor: colors.peach100,
          label: '规划',
          icon: '🟠'
        };
    }
  };

  const style = getIndicatorStyle();
  
  if (compact) {
    return (
      <Tooltip title={`${style.label}模式`}>
        <Box
          width="8px"
          height="8px"
          borderRadius="50%"
          backgroundColor={style.color}
        />
      </Tooltip>
    );
  }

  return (
    <Badge
      color={style.color}
      variant="outline"
      size="small"
    >
      {style.icon} {style.label}
    </Badge>
  );
};

/**
 * 时态字段显示组件
 */
interface TemporalFieldProps {
  organization: OrganizationUnit;
  field: keyof OrganizationUnit;
  mode: TemporalMode;
}

const TemporalField: React.FC<TemporalFieldProps> = ({
  organization,
  field,
  mode
}) => {
  const value = organization[field];
  const isTemporalField = field === 'effective_from' || field === 'effective_to';
  
  // 格式化显示值
  const formatValue = (val: unknown) => {
    if (val === null || val === undefined) return '-';
    if (typeof val === 'boolean') return val ? '是' : '否';
    if (field === 'created_at' || field === 'updated_at' || isTemporalField) {
      try {
        return new Date(val as string).toLocaleDateString('zh-CN');
      } catch {
        return String(val);
      }
    }
    return String(val);
  };

  // 获取状态样式
  const getStatusStyle = (status: string) => {
    switch (status) {
      case 'ACTIVE':
        return { color: colors.greenFresca600, label: '启用' };
      case 'INACTIVE':
        return { color: colors.licorice400, label: '停用' };
      case 'PLANNED':
        return { color: colors.peach600, label: '规划' };
      default:
        return { color: colors.licorice600, label: status };
    }
  };

  // 特殊字段处理
  if (field === 'status') {
    const statusStyle = getStatusStyle(String(value));
    return (
      <Badge color={statusStyle.color} variant="outline" size="small">
        {statusStyle.label}
      </Badge>
    );
  }

  if (field === 'unit_type') {
    const typeLabels = {
      'COMPANY': '公司',
      'DEPARTMENT': '部门',
      'COST_CENTER': '成本中心',
      'PROJECT_TEAM': '项目组'
    };
    return <Text>{typeLabels[value as keyof typeof typeLabels] || value}</Text>;
  }

  // 时态字段高亮显示
  if (isTemporalField && mode !== 'current' && value) {
    return (
      <Text color={colors.blueberry600} fontWeight="medium">
        {formatValue(value)}
      </Text>
    );
  }

  return <Text>{formatValue(value)}</Text>;
};

/**
 * 时态感知数据表格组件
 */
export const TemporalTable: React.FC<TemporalTableProps> = ({
  queryParams,
  showTemporalIndicators = true,
  showActions = true,
  showSelection = false,
  compact = false,
  pageSize = 20,
  onRowClick,
  onEdit,
  onDelete,
  onViewHistory,
  onViewTimeline,
  onSelectionChange
}) => {
  const [selectedRows, setSelectedRows] = useState<Set<string>>(new Set());
  const [currentPage, setCurrentPage] = useState(1);

  // 时态状态
  const temporalContext = temporalSelectors.useContext();
  const isHistorical = temporalContext.mode === 'historical';
  const isPlanning = temporalContext.mode === 'planning';

  // 获取组织数据
  const {
    data: organizations = [],
    isLoading,
    isError,
    error,
    temporalContext: queryContext
  } = useTemporalOrganizations({
    ...queryParams,
    page: currentPage,
    pageSize
  });

  // 表格列定义
  const columns = useMemo(() => {
    const baseColumns = [
      {
        key: 'code' as keyof OrganizationUnit,
        label: '组织代码',
        width: '120px',
        sortable: true
      },
      {
        key: 'name' as keyof OrganizationUnit,
        label: '组织名称',
        width: 'auto',
        sortable: true
      },
      {
        key: 'unit_type' as keyof OrganizationUnit,
        label: '类型',
        width: '100px',
        sortable: true
      },
      {
        key: 'status' as keyof OrganizationUnit,
        label: '状态',
        width: '80px',
        sortable: true
      },
      {
        key: 'level' as keyof OrganizationUnit,
        label: '层级',
        width: '60px',
        sortable: true
      }
    ];

    // 时态模式下添加时态相关列
    if (isHistorical || isPlanning) {
      baseColumns.push(
        {
          key: 'effective_from' as keyof OrganizationUnit,
          label: '生效时间',
          width: '120px',
          sortable: true
        },
        {
          key: 'effective_to' as keyof OrganizationUnit,
          label: '失效时间',
          width: '120px',
          sortable: true
        }
      );
    }

    if (!compact) {
      baseColumns.push({
        key: 'updated_at' as keyof OrganizationUnit,
        label: '更新时间',
        width: '120px',
        sortable: true
      });
    }

    return baseColumns;
  }, [isHistorical, isPlanning, compact]);

  // 选择处理
  const handleRowSelect = useCallback((orgCode: string, selected: boolean) => {
    const newSelection = new Set(selectedRows);
    if (selected) {
      newSelection.add(orgCode);
    } else {
      newSelection.delete(orgCode);
    }
    setSelectedRows(newSelection);

    // 回调选中的组织
    if (onSelectionChange) {
      const selectedOrgs = organizations.filter(org => newSelection.has(org.code));
      onSelectionChange(selectedOrgs);
    }
  }, [selectedRows, organizations, onSelectionChange]);

  // 全选/取消全选
  const handleSelectAll = useCallback((selected: boolean) => {
    if (selected) {
      const allCodes = new Set(organizations.map(org => org.code));
      setSelectedRows(allCodes);
      onSelectionChange?.(organizations);
    } else {
      setSelectedRows(new Set());
      onSelectionChange?.([]);
    }
  }, [organizations, onSelectionChange]);

  // 页面变更
  const handlePageChange = useCallback((page: number) => {
    setCurrentPage(page);
    setSelectedRows(new Set()); // 清空选择
  }, []);

  if (isLoading) {
    return (
      <Box padding={space.m}>
        <Text>🔄 加载组织数据...</Text>
      </Box>
    );
  }

  if (isError) {
    return (
      <Box padding={space.m}>
        <Text color={colors.cinnamon600}>
          ❌ 加载数据失败: {error?.message || '未知错误'}
        </Text>
      </Box>
    );
  }

  const isAllSelected = selectedRows.size > 0 && selectedRows.size === organizations.length;
  const isIndeterminate = selectedRows.size > 0 && selectedRows.size < organizations.length;

  return (
    <Box>
      {/* 表格工具栏 */}
      <Flex justifyContent="space-between" alignItems="center" marginBottom={space.m}>
        <Flex alignItems="center" gap={space.s}>
          {/* 时态模式指示器 */}
          {showTemporalIndicators && (
            <TemporalIndicator
              mode={queryContext.mode}
              organization={organizations[0]}
              compact={compact}
            />
          )}

          <Text fontSize="medium" fontWeight="medium">
            组织架构 ({organizations.length})
          </Text>

          {/* 选择统计 */}
          {showSelection && selectedRows.size > 0 && (
            <Badge color={colors.blueberry600} variant="solid">
              已选择 {selectedRows.size} 项
            </Badge>
          )}
        </Flex>

        {/* 批量操作按钮 */}
        {showSelection && selectedRows.size > 0 && (
          <Flex gap={space.s}>
            <SecondaryButton variant="secondary" size="small">
              <CompareIcon /> 批量对比
            </SecondaryButton>
            <SecondaryButton variant="secondary" size="small">
              <FilterIcon /> 导出选中
            </SecondaryButton>
          </Flex>
        )}
      </Flex>

      {/* 数据表格 */}
      <Box
        border={`1px solid ${colors.soap300}`}
        borderRadius={borderRadius.m}
        overflow="hidden"
      >
        <Table>
          <Table.Head>
            <Table.Row>
              {/* 选择列 */}
              {showSelection && (
                <Table.Header width="50px">
                  <Checkbox
                    checked={isAllSelected}
                    indeterminate={isIndeterminate}
                    onChange={(e) => handleSelectAll(e.target.checked)}
                  />
                </Table.Header>
              )}

              {/* 时态指示列 */}
              {showTemporalIndicators && (
                <Table.Header width="40px">
                  <Tooltip title="时态状态">
                    <InfoIcon size="small" />
                  </Tooltip>
                </Table.Header>
              )}

              {/* 数据列 */}
              {columns.map(column => (
                <Table.Header key={column.key} width={column.width}>
                  {column.label}
                </Table.Header>
              ))}

              {/* 操作列 */}
              {showActions && (
                <Table.Header width="120px">操作</Table.Header>
              )}
            </Table.Row>
          </Table.Head>

          <Table.Body>
            {organizations.map((organization) => (
              <Table.Row
                key={organization.code}
                style={{
                  cursor: onRowClick ? 'pointer' : 'default'
                }}
                onClick={() => onRowClick?.(organization)}
              >
                {/* 选择列 */}
                {showSelection && (
                  <Table.Cell>
                    <Checkbox
                      checked={selectedRows.has(organization.code)}
                      onChange={(e) => {
                        e.stopPropagation();
                        handleRowSelect(organization.code, e.target.checked);
                      }}
                    />
                  </Table.Cell>
                )}

                {/* 时态指示列 */}
                {showTemporalIndicators && (
                  <Table.Cell>
                    <TemporalIndicator
                      mode={queryContext.mode}
                      organization={organization}
                      compact={true}
                    />
                  </Table.Cell>
                )}

                {/* 数据列 */}
                {columns.map(column => (
                  <Table.Cell key={column.key}>
                    <TemporalField
                      organization={organization}
                      field={column.key}
                      mode={queryContext.mode}
                    />
                  </Table.Cell>
                ))}

                {/* 操作列 */}
                {showActions && (
                  <Table.Cell>
                    <Flex gap={space.xs}>
                      {/* 编辑按钮 - 历史模式下禁用 */}
                      <Tooltip title={isHistorical ? '历史模式下不可编辑' : '编辑组织'}>
                        <IconButton
                          variant="plain"
                          size="small"
                          disabled={isHistorical}
                          onClick={(e) => {
                            e.stopPropagation();
                            onEdit?.(organization);
                          }}
                        >
                          <EditIcon />
                        </IconButton>
                      </Tooltip>

                      {/* 历史按钮 */}
                      {onViewHistory && (
                        <Tooltip title="查看历史版本">
                          <IconButton
                            variant="plain"
                            size="small"
                            onClick={(e) => {
                              e.stopPropagation();
                              onViewHistory(organization);
                            }}
                          >
                            <HistoryIcon />
                          </IconButton>
                        </Tooltip>
                      )}

                      {/* 时间线按钮 */}
                      {onViewTimeline && (
                        <Tooltip title="查看时间线">
                          <IconButton
                            variant="plain"
                            size="small"
                            onClick={(e) => {
                              e.stopPropagation();
                              onViewTimeline(organization);
                            }}
                          >
                            <TimelineIcon />
                          </IconButton>
                        </Tooltip>
                      )}

                      {/* 删除按钮 - 历史模式下禁用 */}
                      {onDelete && (
                        <Tooltip title={isHistorical ? '历史模式下不可删除' : '删除组织'}>
                          <IconButton
                            variant="plain"
                            size="small"
                            disabled={isHistorical}
                            onClick={(e) => {
                              e.stopPropagation();
                              onDelete(organization);
                            }}
                          >
                            <DeleteIcon />
                          </IconButton>
                        </Tooltip>
                      )}
                    </Flex>
                  </Table.Cell>
                )}
              </Table.Row>
            ))}
          </Table.Body>
        </Table>

        {/* 空状态 */}
        {organizations.length === 0 && (
          <Box padding={space.l} textAlign="center">
            <Text color={colors.licorice500}>
              📭 没有找到符合条件的组织数据
            </Text>
          </Box>
        )}
      </Box>

      {/* 分页信息 */}
      {organizations.length > 0 && (
        <Flex justifyContent="space-between" alignItems="center" marginTop="m">
          <Text typeLevel="subtext.small" color="hint">
            显示第 {(currentPage - 1) * pageSize + 1} - {Math.min(currentPage * pageSize, organizations.length)} 项，
            共 {organizations.length} 项
          </Text>
          
          <Flex gap="s" alignItems="center">
            <SecondaryButton
              size="small"
              disabled={currentPage <= 1}
              onClick={() => handlePageChange(currentPage - 1)}
            >
              上一页
            </SecondaryButton>
            
            <Text typeLevel="subtext.small">
              第 {currentPage} 页
            </Text>
            
            <SecondaryButton
              size="small"
              disabled={currentPage >= Math.ceil(organizations.length / pageSize)}
              onClick={() => handlePageChange(currentPage + 1)}
            >
              下一页
            </SecondaryButton>
          </Flex>
        </Flex>
      )}

      {/* 时态模式提示 */}
      {(isHistorical || isPlanning) && (
        <Box marginTop={space.s}>
          <Text fontSize="small" color={colors.licorice500}>
            ℹ️ {isHistorical ? '当前显示历史' : '当前显示规划'}模式数据，
            {isHistorical && '编辑和删除功能已禁用'}
            {isPlanning && '显示未来规划的组织变更'}
          </Text>
        </Box>
      )}
    </Box>
  );
};

export default TemporalTable;