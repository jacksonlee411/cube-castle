/**
 * 时间轴组件
 * 基于 Canvas Kit v13 企业级设计系统
 * 依赖 GraphQL 返回的组织版本基础字段进行渲染（当前实际可用状态仅为 ACTIVE/INACTIVE + isCurrent；
 * PLANNED/DELETED 等分支需调用方自行派生或在 API 扩展后使用）
 */
import React from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text, Heading } from '@workday/canvas-kit-react/text';
import { Card } from '@workday/canvas-kit-react/card';
import { LoadingDots } from '@workday/canvas-kit-react/loading-dots';
import { 
  colors, 
  borderRadius 
} from '@workday/canvas-kit-react/tokens';
import { StatusBadge } from '../../../shared/components/StatusBadge';
import type { OrganizationStatus } from '@/shared/types';
import { OrganizationStatusEnum } from '@/shared/types/contract_gen';
import temporalEntitySelectors from '@/shared/testids/temporalEntity';

// 时态版本接口定义 - 兼容历史字段（多数由调用方派生）
export interface TimelineVersion {
  recordId: string; // UUID唯一标识符
  code: string;
  name: string;
  unitType: string;
  status: string; // 组织状态：API 目前仅返回 ACTIVE/INACTIVE，PLANNED 等值需调用方自行派生
  effectiveDate: string;
  endDate?: string | null;
  changeReason?: string;
  isCurrent: boolean;
  createdAt: string;
  updatedAt: string;
  description?: string;
  level: number;
  codePath?: string | null;
  namePath?: string | null;
  parentCode?: string;
  sortOrder: number;

  // 兼容字段：调用方可选传入派生状态；若未传入则组件退化为基于二态(status/isCurrent)渲染
  lifecycleStatus?: 'CURRENT' | 'HISTORICAL' | 'PLANNED';
  businessStatus?: 'ACTIVE' | 'INACTIVE';
  dataStatus?: 'NORMAL' | 'DELETED';
  suspendedAt?: string | null; // 停用时间
  suspendedBy?: string | null; // 停用者
  suspensionReason?: string | null; // 停用原因
  deletedAt?: string | null; // 删除时间
  deletedBy?: string | null; // 删除者
  deletionReason?: string | null; // 删除原因
}

// 时间轴组件属性接口
export interface TimelineComponentProps {
  versions: TimelineVersion[];
  selectedVersion: TimelineVersion | null;
  onVersionSelect: (version: TimelineVersion) => void;
  onDeleteVersion?: (version: TimelineVersion) => void;
  isLoading: boolean;
  readonly?: boolean;
  width?: string; // 允许自定义宽度
  height?: string; // 允许自定义高度
  title?: string; // 可自定义标题
  showActions?: boolean; // 是否显示操作按钮
}

// 状态映射函数：将后端状态映射到组织状态系统
const mapBackendStatusToOrganizationStatus = (backendStatus: string): OrganizationStatus => {
  // 映射到API契约的业务状态
  switch (backendStatus) {
    case OrganizationStatusEnum.Inactive:
      return OrganizationStatusEnum.Inactive;
    case OrganizationStatusEnum.Planned:
      return OrganizationStatusEnum.Planned;
    case OrganizationStatusEnum.Deleted:
      return OrganizationStatusEnum.Deleted;
    case OrganizationStatusEnum.Active:
    default:
      return OrganizationStatusEnum.Active;
  }
};

const isSoftDeleted = (version: TimelineVersion) =>
  version.dataStatus === 'DELETED' || version.status === OrganizationStatusEnum.Deleted;

const resolveBusinessStatus = (version: TimelineVersion): 'ACTIVE' | 'INACTIVE' => {
  if (version.businessStatus) {
    return version.businessStatus;
  }

  if (
    version.status === OrganizationStatusEnum.Inactive ||
    version.status === OrganizationStatusEnum.Deleted
  ) {
    return 'INACTIVE';
  }

  return 'ACTIVE';
};

const resolveLifecycleStatus = (
  version: TimelineVersion,
): 'CURRENT' | 'HISTORICAL' | 'PLANNED' => {
  if (version.lifecycleStatus) {
    if (version.lifecycleStatus === 'PLANNED') {
      return 'PLANNED';
    }
    return version.lifecycleStatus === 'CURRENT' ? 'CURRENT' : 'HISTORICAL';
  }

  if (version.status === OrganizationStatusEnum.Planned) {
    return 'PLANNED';
  }

  return version.isCurrent ? 'CURRENT' : 'HISTORICAL';
};

/**
 * 健壮版时间轴组件
 * 使用Canvas Kit v13组件，遵循企业级设计标准
 */
export const TimelineComponent: React.FC<TimelineComponentProps> = ({
  versions,
  selectedVersion,
  onVersionSelect,
  onDeleteVersion: _onDeleteVersion,
  isLoading,
  readonly: _readonly = false,
  width = "350px",
  height = "calc(100vh - 200px)",
  title = "时间轴导航",
  showActions: _showActions = true
}) => {
  
  // 获取版本状态指示器 - 基于现有字段派生 UI 提示
  const getVersionStatusIndicator = (version: TimelineVersion) => {
    if (isSoftDeleted(version)) {
      return {
        color: colors.cinnamon600,
        dotColor: colors.cinnamon600,
        label: '已删除',
        isDeactivated: true,
        badge: 'DELETED' as const,
      };
    }

    const businessStatus = resolveBusinessStatus(version);
    if (businessStatus === 'INACTIVE') {
      return {
        color: colors.cantaloupe600,
        dotColor: colors.cantaloupe600,
        label: '已结束',
        isDeactivated: false,
        badge: 'INACTIVE' as const,
      };
    }

    const lifecycleStatus = resolveLifecycleStatus(version);
    if (lifecycleStatus === 'PLANNED') {
      return {
        color: colors.blueberry600,
        dotColor: 'white',
        label: '计划中',
        isDeactivated: false,
        badge: 'PLANNED' as const,
      };
    }

    if (lifecycleStatus === 'CURRENT') {
      return {
        color: colors.greenApple500,
        dotColor: colors.greenApple500,
        label: '生效中',
        isDeactivated: false,
        badge: 'CURRENT' as const,
      };
    }

    return {
      color: colors.licorice400,
      dotColor: colors.licorice400,
      label: '历史记录',
      isDeactivated: false,
      badge: 'HISTORICAL' as const,
    };
  };

  // 格式化日期
  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('zh-CN');
  };

  // 计算日期范围显示
  const formatDateRange = (version: TimelineVersion, allVersions: TimelineVersion[]) => {
    const start = formatDate(version.effectiveDate);

    if (isSoftDeleted(version)) {
      return `${start} ~ 已删除`;
    }
    
    // 根据时态管理规则计算结束日期
    if (version.endDate) {
      // 如果有明确的结束日期，使用它
      return `${start} ~ ${formatDate(version.endDate)}`;
    }
    
    // 找到下一个生效日期更晚的版本（排除已删除的版本）
    const nextVersion = allVersions
      .filter((v) => new Date(v.effectiveDate) > new Date(version.effectiveDate))
      .filter((v) => !isSoftDeleted(v))
      .sort((a, b) => new Date(a.effectiveDate).getTime() - new Date(b.effectiveDate).getTime())[0];

    if (nextVersion) {
      // 如果有下一个版本，当前版本的结束日期是下一个版本生效日期的前一天
      const nextDate = new Date(nextVersion.effectiveDate);
      nextDate.setDate(nextDate.getDate() - 1);
      return `${start} ~ ${formatDate(nextDate.toISOString().split('T')[0])}`;
    }
    
    const lifecycleStatus = resolveLifecycleStatus(version);
    if (lifecycleStatus === 'PLANNED') {
      return `${start} ~ 计划中`;
    }

    return `${start} ~ 至今`;
  };

  // 增强版时间范围显示 - 提供更直观的时间信息
  const getEnhancedDateRange = (version: TimelineVersion, allVersions: TimelineVersion[]) => {
    const baseRange = formatDateRange(version, allVersions);
    const lifecycleStatus = resolveLifecycleStatus(version);
    
    // 计算持续时间
    const calculateDuration = (startDate: string, endDate?: string | null) => {
      const start = new Date(startDate);
      const end = endDate ? new Date(endDate) : new Date();
      const diffTime = Math.abs(end.getTime() - start.getTime());
      const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
      
      if (diffDays < 30) {
        return `${diffDays}天`;
      } else if (diffDays < 365) {
        const months = Math.floor(diffDays / 30);
        return `${months}个月`;
      } else {
        const years = Math.floor(diffDays / 365);
        const remainingMonths = Math.floor((diffDays % 365) / 30);
        return remainingMonths > 0 ? `${years}年${remainingMonths}个月` : `${years}年`;
      }
    };

    // 获取状态图标
    const getStatusIcon = (status: 'CURRENT' | 'HISTORICAL' | 'PLANNED') => {
      switch (status) {
        case 'CURRENT':
          return '🟢';
        case 'PLANNED':
          return '🔵';
        case 'HISTORICAL':
        default:
          return '⚪';
      }
    };

    const duration = version.endDate 
      ? calculateDuration(version.effectiveDate, version.endDate)
      : lifecycleStatus === 'CURRENT'
        ? calculateDuration(version.effectiveDate)
        : '未确定';

    return {
      primary: baseRange,
      duration: lifecycleStatus !== 'PLANNED' ? duration : '计划中',
      icon: getStatusIcon(lifecycleStatus),
      isActive: lifecycleStatus === 'CURRENT'
    };
  };

  return (
    <Box
      width={width}
      height={height}
      backgroundColor="#F8F9FA"
      borderRadius={borderRadius.m}
      border="1px solid #E9ECEF"
      padding="m"
      overflowY="auto"
      data-testid={temporalEntitySelectors.page.timeline}
    >
      {/* 操作区域 */}
      <Box marginBottom="m">
        <Flex justifyContent="space-between" alignItems="center" marginBottom="s">
          <Heading size="small">{title}</Heading>
        </Flex>
        <Text typeLevel="subtext.small" color="hint">
          点击版本节点查看详情；如需修改请前往“版本历史”页签
        </Text>
      </Box>

      {/* 时间轴节点列表 */}
      {isLoading ? (
        <Box textAlign="center" padding="l">
          <LoadingDots />
          <Text marginTop="s" typeLevel="subtext.small">加载中...</Text>
        </Box>
      ) : (
        <Box position="relative">
          {/* 时间线连接线 - 增强版本 */}
          <Box
            position="absolute"
            left="15px"
            top="20px"
            bottom="20px"
            width="3px"
            backgroundColor="#B8C4D0"
            borderRadius="2px"
            zIndex={0}
            style={{
              background: 'linear-gradient(to bottom, #B8C4D0 0%, #D1D9E0 50%, #B8C4D0 100%)'
            }}
          />

          {/* 版本节点 */}
          {versions.map((version) => {
            const statusInfo = getVersionStatusIndicator(version);
            const isSelected = selectedVersion?.effectiveDate === version.effectiveDate;
            
            return (
              <Box
                key={`${version.code}-${version.effectiveDate}`}
                position="relative"
                marginBottom="m"
                zIndex={1}
              >
                {/* 节点圆点 - 增强版本 */}
                <Box
                  position="absolute"
                  left="-5px"
                  top="8px"
                  width="14px"
                  height="14px"
                  borderRadius="50%"
                  backgroundColor={statusInfo.dotColor}
                  border="3px solid white"
                  boxShadow="0 3px 6px rgba(0,0,0,0.15)"
                  style={{
                    transform: isSelected ? 'scale(1.1)' : 'scale(1)',
                    transition: 'all 0.2s ease'
                  }}
                />

                {/* 节点内容卡片 */}
                <Box marginLeft="32px">
                  <Card
                    padding="s"
                    data-testid={temporalEntitySelectors.page.timelineNode}
                    data-lifecycle={statusInfo.badge}
                    data-current={isSelected ? 'true' : 'false'}
                    data-status={version.status}
                    data-business-status={resolveBusinessStatus(version)}
                    style={{
                      backgroundColor: isSelected ? '#E3F2FD' : 'white',
                      border: isSelected ? '2px solid #2196F3' : '1px solid #E9ECEF',
                      cursor: 'pointer',
                      transition: 'all 0.2s ease',
                      boxShadow: isSelected 
                        ? '0 4px 12px rgba(33, 150, 243, 0.2)' 
                        : '0 1px 3px rgba(0,0,0,0.1)',
                      transform: isSelected ? 'translateY(-1px)' : 'translateY(0)',
                      opacity: statusInfo.isDeactivated ? 0.7 : 1
                    }}
                    onClick={() => onVersionSelect(version)}
                    onMouseEnter={(e) => {
                      if (!isSelected && !statusInfo.isDeactivated) {
                        e.currentTarget.style.boxShadow = '0 2px 8px rgba(0,0,0,0.15)';
                        e.currentTarget.style.transform = 'translateY(-0.5px)';
                      }
                    }}
                    onMouseLeave={(e) => {
                      if (!isSelected) {
                        e.currentTarget.style.boxShadow = '0 1px 3px rgba(0,0,0,0.1)';
                        e.currentTarget.style.transform = 'translateY(0)';
                      }
                    }}
                  >
                    {/* 节点头部 - 日期与状态同行 */}
                    <Box marginBottom="xs">
                      <Flex alignItems="center" justifyContent="space-between">
                        {/* 生效日期 */}
                        <Text 
                          typeLevel="body.medium" 
                          fontWeight="bold"
                          style={{
                            textDecoration: statusInfo.isDeactivated ? 'line-through' : 'none'
                          }}
                        >
                          {formatDate(version.effectiveDate)}
                        </Text>
                        
                        {/* 状态标识 - 使用统一的状态系统 */}
                        <Box
                          data-testid={temporalEntitySelectors.page.lifecycleBadge}
                          data-lifecycle={statusInfo.badge}
                        >
                          <StatusBadge 
                            status={mapBackendStatusToOrganizationStatus(version.status)} 
                            size="small"
                          />
                        </Box>
                      </Flex>
                    </Box>
                    
                    {/* 组织名称 */}
                    <Box marginBottom="xs">
                      <Text 
                        typeLevel="body.small" 
                        fontWeight="medium"
                        style={{
                          textDecoration: statusInfo.isDeactivated ? 'line-through' : 'none'
                        }}
                      >
                        {version.name}
                      </Text>
                    </Box>

                    {/* 时间范围 - 增强版本 */}
                    <Box>
                      {(() => {
                        const enhancedRange = getEnhancedDateRange(version, versions);
                        return (
                          <>
                            <Flex alignItems="center" marginBottom="xxs">
                              <Text typeLevel="subtext.small" color="hint">
                                有效期间：
                              </Text>
                              <Text 
                                typeLevel="subtext.small" 
                                marginLeft="xs"
                                style={{ 
                                  fontSize: '14px',
                                  fontWeight: enhancedRange.isActive ? '600' : 'normal'
                                }}
                              >
                                {enhancedRange.icon} {enhancedRange.primary}
                              </Text>
                            </Flex>
                            {enhancedRange.duration !== '未确定' && enhancedRange.duration !== '计划中' && (
                              <Flex alignItems="center">
                                <Text typeLevel="subtext.small" color="hint">
                                  持续时间：
                                </Text>
                                <Text 
                                  typeLevel="subtext.small" 
                                  color={enhancedRange.isActive ? colors.greenApple600 : "hint"}
                                  marginLeft="xs"
                                  fontWeight="medium"
                                >
                                  {enhancedRange.duration}
                                </Text>
                              </Flex>
                            )}
                          </>
                        );
                      })()}
                    </Box>
                  </Card>
                </Box>
              </Box>
            );
          })}

          {versions.length === 0 && (
            <Box textAlign="center" padding="l">
              <Text color="hint">暂无版本记录</Text>
            </Box>
          )}
        </Box>
      )}
    </Box>
  );
};

export default TimelineComponent;
