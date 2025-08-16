/**
 * 时态管理主从视图组件
 * 左侧：垂直交互式时间轴导航
 * 右侧：动态版本详情卡片
 * 实现强制时间连续性的完整用户体验
 */
import React, { useState, useCallback, useEffect } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text, Heading } from '@workday/canvas-kit-react/text';
import { Card } from '@workday/canvas-kit-react/card';
import { PrimaryButton, SecondaryButton, TertiaryButton } from '@workday/canvas-kit-react/button';
import { Badge } from '../../../shared/components/Badge';
import { Tooltip } from '@workday/canvas-kit-react/tooltip';
import { Modal, useModalModel } from '@workday/canvas-kit-react/modal';
import TemporalEditForm, { type TemporalEditFormData } from './TemporalEditForm';
import { InlineNewVersionForm } from './InlineNewVersionForm';
import { SimpleTimelineVisualization } from './SimpleTimelineVisualization';
import { LoadingDots } from '@workday/canvas-kit-react/loading-dots';
import { 
  colors, 
  borderRadius 
} from '@workday/canvas-kit-react/tokens';
import { baseColors } from '../../../shared/utils/colorTokens';
// 暂时使用文本图标替代
// import {
//   addIcon,
//   editIcon,
//   deleteIcon,
//   moreVerticalIcon,
//   calendarIcon,
//   infoIcon,
//   warningIcon
// } from '@workday/canvas-system-icons-web';

// Types
export interface TemporalVersion {
  code: string;
  name: string;
  unit_type: string;
  status: string;
  effective_date: string;
  end_date?: string | null;
  change_reason?: string;
  is_current: boolean;
  created_at: string;
  updated_at: string;
  description?: string;
  level: number;
  path: string;
  parent_code?: string;
  sort_order: number;
}

export interface TemporalMasterDetailViewProps {
  organizationCode: string;
  onBack?: () => void;
  readonly?: boolean;
}

/**
 * 左侧垂直时间轴导航区
 */
interface TimelineNavigationProps {
  versions: TemporalVersion[];
  selectedVersion: TemporalVersion | null;
  onVersionSelect: (version: TemporalVersion) => void;
  onAddVersion?: () => void;
  onDeleteVersion?: (version: TemporalVersion) => void;
  isLoading: boolean;
  readonly?: boolean;
}

const TimelineNavigation: React.FC<TimelineNavigationProps> = ({
  versions,
  selectedVersion,
  onVersionSelect,
  onAddVersion,
  onDeleteVersion,
  isLoading,
  readonly = false
}) => {
  // 获取版本状态指示器
  const getVersionStatusIndicator = (version: TemporalVersion) => {
    const today = new Date();
    const effectiveDate = new Date(version.effective_date);
    const endDate = version.end_date ? new Date(version.end_date) : null;
    
    if (version.is_current) {
      return { color: '#2ECC71', icon: '🟢', label: '生效中' }; // 绿色
    } else if (effectiveDate > today) {
      return { color: '#3498DB', icon: '🔵', label: '计划中' }; // 蓝色
    } else if (endDate && endDate < today) {
      return { color: '#95A5A6', icon: '⚫', label: '已结束' }; // 灰色
    } else {
      return { color: '#E74C3C', icon: '🔴', label: '已作废' }; // 红色
    }
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('zh-CN');
  };

  const formatDateRange = (startDate: string, endDate?: string | null) => {
    const start = formatDate(startDate);
    if (!endDate) return `${start} ~ 至今`;
    return `${start} ~ ${formatDate(endDate)}`;
  };

  return (
    <Box
      width="350px"
      height="calc(100vh - 200px)"
      backgroundColor="#F8F9FA"
      borderRadius={borderRadius.m}
      border="1px solid #E9ECEF"
      padding="m"
      overflowY="auto"
    >
      {/* 操作区域 */}
      <Box marginBottom="m">
        <Flex justifyContent="space-between" alignItems="center" marginBottom="s">
          <Heading size="small">时间轴导航</Heading>
          {!readonly && onAddVersion && (
            <Tooltip title="新增版本">
              <TertiaryButton
                aria-label="新增版本"
                onClick={onAddVersion}
                size="small"
              >
                ➕
              </TertiaryButton>
            </Tooltip>
          )}
        </Flex>
        <Text typeLevel="subtext.small" color="hint">
          点击版本节点查看详情
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
          {/* 时间线连接线 */}
          <Box
            position="absolute"
            left="15px"
            top="20px"
            bottom="20px"
            width="2px"
            backgroundColor="#DEE2E6"
            zIndex={0}
          />

          {/* 版本节点 */}
          {versions.map((version) => {
            const statusInfo = getVersionStatusIndicator(version);
            const isSelected = selectedVersion?.effective_date === version.effective_date;
            
            return (
              <Box
                key={`${version.code}-${version.effective_date}`}
                position="relative"
                marginBottom="m"
                zIndex={1}
              >
                {/* 节点圆点 */}
                <Box
                  position="absolute"
                  left="-4px"
                  top="8px"
                  width="12px"
                  height="12px"
                  borderRadius="50%"
                  backgroundColor={statusInfo.color}
                  border="2px solid white"
                  boxShadow="0 2px 4px rgba(0,0,0,0.1)"
                />

                {/* 节点内容卡片 */}
                <Box marginLeft="32px">
                  <Card
                    padding="s"
                    style={{
                      backgroundColor: isSelected ? '#E3F2FD' : 'white',
                      border: isSelected ? '2px solid #2196F3' : '1px solid #E9ECEF',
                      cursor: 'pointer',
                      transition: 'all 0.2s ease'
                    }}
                    onClick={() => onVersionSelect(version)}
                  >
                    {/* 节点头部 */}
                    <Flex justifyContent="space-between" alignItems="flex-start" marginBottom="xs">
                      <Box flex="1">
                        <Text typeLevel="subtext.medium" fontWeight="bold">
                          {formatDate(version.effective_date)}
                        </Text>
                        <Badge 
                          color={statusInfo.color.replace('#', '') as 'primary' | 'secondary' | 'success' | 'warning' | 'danger'}
                          size="small"
                        >
                          {statusInfo.label}
                        </Badge>
                      </Box>
                      
                      {!readonly && onDeleteVersion && !version.is_current && (
                        <Tooltip title="作废版本">
                          <TertiaryButton
                            aria-label="作废版本"
                            size="small"
                            onClick={(e: React.MouseEvent<HTMLButtonElement>) => {
                              e.stopPropagation();
                              onDeleteVersion(version);
                            }}
                          >
                            🗑️
                          </TertiaryButton>
                        </Tooltip>
                      )}
                    </Flex>

                    {/* 版本名称 */}
                    <Text 
                      typeLevel="body.small" 
                      fontWeight="medium"
                      marginBottom="xs"
                      style={{ 
                        whiteSpace: 'nowrap',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis'
                      }}
                    >
                      {version.name}
                    </Text>

                    {/* 变更摘要 */}
                    {version.change_reason && (
                      <Text 
                        typeLevel="subtext.small" 
                        color="hint"
                        style={{ 
                          whiteSpace: 'nowrap',
                          overflow: 'hidden',
                          textOverflow: 'ellipsis'
                        }}
                      >
                        {version.change_reason}
                      </Text>
                    )}

                    {/* 时间范围 */}
                    <Text typeLevel="subtext.small" color="hint" marginTop="xs">
                      {formatDateRange(version.effective_date, version.end_date)}
                    </Text>
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

/**
 * 右侧动态版本详情卡片区
 */
interface VersionDetailCardProps {
  version: TemporalVersion | null;
  onEdit?: (version: TemporalVersion) => void;
  onDelete?: (version: TemporalVersion) => void;
  isLoading?: boolean;
  readonly?: boolean;
}

const VersionDetailCard: React.FC<VersionDetailCardProps> = ({
  version,
  onEdit,
  onDelete,
  isLoading = false,
  readonly = false
}) => {
  if (!version) {
    return (
      <Flex
        flex={1}
        padding="l"
        alignItems="center"
        justifyContent="center"
        backgroundColor="#F8F9FA"
        borderRadius={borderRadius.m}
        border="1px solid #E9ECEF"
      >
        <Box textAlign="center">
          <img src="data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiPjxjaXJjbGUgY3g9IjI0IiBjeT0iMjQiIHI9IjIwIiBmaWxsPSIjQ0NDIi8+PC9zdmc+" width={48} height={48} alt="Calendar" />
          <Text typeLevel="subtext.large" color="hint" marginTop="m">
            请选择左侧时间轴节点查看版本详情
          </Text>
        </Box>
      </Flex>
    );
  }

  const getUnitTypeName = (unitType: string) => {
    const typeNames = {
      'COMPANY': '公司',
      'DEPARTMENT': '部门', 
      'COST_CENTER': '成本中心',
      'PROJECT_TEAM': '项目团队'
    };
    return typeNames[unitType as keyof typeof typeNames] || unitType;
  };

  const getStatusBadge = (status: string) => {
    const statusConfig = {
      'ACTIVE': { label: '启用', color: 'greenFresca600' },
      'INACTIVE': { label: '停用', color: 'cinnamon600' },
      'PLANNED': { label: '计划中', color: 'blueberry600' }
    };
    
    const config = statusConfig[status as keyof typeof statusConfig] || { 
      label: status, 
      color: 'licorice400' 
    };
    return <Badge color={config.color as 'greenFresca600' | 'cinnamon600' | 'blueberry600' | 'licorice400'}>{config.label}</Badge>;
  };

  // 智能操作按钮逻辑
  const getButtonState = () => {
    const today = new Date();
    const effectiveDate = new Date(version.effective_date);
    const endDate = version.end_date ? new Date(version.end_date) : null;
    
    if (endDate && endDate < today) {
      // 历史记录
      return { 
        edit: 'disabled', 
        delete: 'disabled', 
        tooltip: '历史记录不可修改' 
      };
    } else if (version.is_current) {
      // 当前版本
      return { 
        edit: 'limited', 
        delete: 'confirm-as-invalid', 
        tooltip: '当前版本需谨慎操作' 
      };
    } else if (effectiveDate > today) {
      // 未来版本
      return { 
        edit: 'enabled', 
        delete: 'enabled', 
        tooltip: '可自由编辑计划版本' 
      };
    }
    
    return { edit: 'enabled', delete: 'enabled', tooltip: '' };
  };

  const buttonState = getButtonState();

  return (
    <Box flex="1" padding="m">
      <Card padding="l">
        {/* 动态标题 */}
        <Flex justifyContent="space-between" alignItems="flex-start" marginBottom="l">
          <Box>
            <Heading size="medium" marginBottom="s">
              版本详情 (生效于: {new Date(version.effective_date).toLocaleDateString('zh-CN')})
            </Heading>
            <Flex alignItems="center" gap="s">
              {getStatusBadge(version.status)}
              {version.is_current && (
                <Badge color="greenFresca600">当前版本</Badge>
              )}
            </Flex>
          </Box>

          {/* 智能操作按钮 */}
          {!readonly && (
            <Flex gap="s">
              <Tooltip title={buttonState.edit === 'disabled' ? buttonState.tooltip : '编辑版本'}>
                <PrimaryButton
                  size="small"
                  disabled={buttonState.edit === 'disabled' || isLoading}
                  onClick={() => onEdit?.(version)}
                >
                  ✏️ 编辑
                </PrimaryButton>
              </Tooltip>
              
              <Tooltip title={buttonState.delete === 'disabled' ? buttonState.tooltip : '作废版本'}>
                <SecondaryButton
                  size="small"
                  disabled={buttonState.delete === 'disabled' || isLoading}
                  onClick={() => onDelete?.(version)}
                >
                  🗑️ 作废
                </SecondaryButton>
              </Tooltip>
            </Flex>
          )}
        </Flex>

        {/* 版本详细信息 */}
        <Box
          cs={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(250px, 1fr))",
            gap: "16px" // 使用像素值而不是token
          }}
        >
          {/* 基本信息 */}
          <Box>
            <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s" color={baseColors.blueberry[600]}>
              📋 基本信息
            </Text>
            <Box marginLeft="m">
              <Text typeLevel="body.small" marginBottom="xs">
                <strong>组织名称:</strong> {version.name}
              </Text>
              <Text typeLevel="body.small" marginBottom="xs">
                <strong>组织编码:</strong> {version.code}
              </Text>
              <Text typeLevel="body.small" marginBottom="xs">
                <strong>组织类型:</strong> {getUnitTypeName(version.unit_type)}
              </Text>
              <Text typeLevel="body.small" marginBottom="xs">
                <strong>当前状态:</strong> {version.status}
              </Text>
            </Box>
          </Box>

          {/* 层级信息 */}
          <Box>
            <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s" color={baseColors.peach[600]}>
              🏗️ 层级结构
            </Text>
            <Box marginLeft="m">
              <Text typeLevel="body.small" marginBottom="xs">
                <strong>层级:</strong> 第 {version.level} 级
              </Text>
              <Text typeLevel="body.small" marginBottom="xs">
                <strong>上级组织:</strong> {version.parent_code || '无'}
              </Text>
              <Text typeLevel="body.small" marginBottom="xs">
                <strong>路径:</strong> {version.path}
              </Text>
              <Text typeLevel="body.small" marginBottom="xs">
                <strong>排序:</strong> {version.sort_order}
              </Text>
            </Box>
          </Box>

          {/* 时态信息 */}
          <Box>
            <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s" color={baseColors.greenFresca[600]}>
              ⏰ 生效期间
            </Text>
            <Box marginLeft="m">
              <Text typeLevel="body.small" marginBottom="xs">
                <strong>生效日期:</strong> {new Date(version.effective_date).toLocaleDateString('zh-CN')}
              </Text>
              <Text typeLevel="body.small" marginBottom="xs">
                <strong>失效日期:</strong> {
                  version.end_date 
                    ? new Date(version.end_date).toLocaleDateString('zh-CN')
                    : '无限期有效'
                }
              </Text>
              <Text typeLevel="body.small" marginBottom="xs">
                <strong>变更原因:</strong> {version.change_reason || '无'}
              </Text>
            </Box>
          </Box>

          {/* 系统信息 */}
          <Box>
            <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s" color={baseColors.cantaloupe[600]}>
              🔧 系统信息
            </Text>
            <Box marginLeft="m">
              <Text typeLevel="body.small" marginBottom="xs">
                <strong>创建时间:</strong> {new Date(version.created_at).toLocaleString('zh-CN')}
              </Text>
              <Text typeLevel="body.small" marginBottom="xs">
                <strong>更新时间:</strong> {new Date(version.updated_at).toLocaleString('zh-CN')}
              </Text>
              <Text typeLevel="body.small" marginBottom="xs">
                <strong>是否当前:</strong> {version.is_current ? '是' : '否'}
              </Text>
            </Box>
          </Box>
        </Box>

        {/* 描述信息 */}
        {version.description && (
          <Box marginTop="m" padding="m" backgroundColor={baseColors.soap[300]} borderRadius={borderRadius.s}>
            <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">
              📝 描述信息
            </Text>
            <Text typeLevel="body.medium">
              {version.description}
            </Text>
          </Box>
        )}
      </Card>
    </Box>
  );
};

/**
 * 时态管理主从视图主组件
 */
export const TemporalMasterDetailView: React.FC<TemporalMasterDetailViewProps> = ({
  organizationCode,
  onBack,
  readonly = false
}) => {
  // 状态管理
  const [versions, setVersions] = useState<TemporalVersion[]>([]);
  const [selectedVersion, setSelectedVersion] = useState<TemporalVersion | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState<TemporalVersion | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);
  
  // 编辑表单状态
  const [showEditForm, setShowEditForm] = useState(false);
  const [editMode, setEditMode] = useState<'create' | 'edit'>('create');
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  // 视图选项卡状态
  const [activeTab, setActiveTab] = useState<'details' | 'timeline' | 'new-version'>('details');

  // Modal model for delete confirmation
  const deleteModalModel = useModalModel();

  // 同步Modal状态
  React.useEffect(() => {
    if (showDeleteConfirm && deleteModalModel.state.visibility !== 'visible') {
      deleteModalModel.events.show();
    } else if (!showDeleteConfirm && deleteModalModel.state.visibility === 'visible') {
      deleteModalModel.events.hide();
    }
  }, [showDeleteConfirm, deleteModalModel]);

  // 加载时态版本数据
  const loadVersions = useCallback(async () => {
    try {
      setIsLoading(true);
      const response = await fetch(
        `http://localhost:9091/api/v1/organization-units/${organizationCode}/temporal?include_history=true&include_future=true`,
        {
          method: 'GET',
          headers: { 'Content-Type': 'application/json' }
        }
      );
      
      if (response.ok) {
        const data = await response.json();
        const sortedVersions = data.organizations.sort((a: TemporalVersion, b: TemporalVersion) => 
          new Date(b.effective_date).getTime() - new Date(a.effective_date).getTime()
        );
        setVersions(sortedVersions);
        
        // 默认选中当前版本
        const currentVersion = sortedVersions.find((v: TemporalVersion) => v.is_current);
        if (currentVersion) {
          setSelectedVersion(currentVersion);
        } else if (sortedVersions.length > 0) {
          setSelectedVersion(sortedVersions[0]);
        }
      } else {
        console.error('Failed to load temporal versions:', response.statusText);
      }
    } catch (error) {
      console.error('Error loading temporal versions:', error);
    } finally {
      setIsLoading(false);
    }
  }, [organizationCode]);

  // 作废版本处理
  const handleDeleteVersion = useCallback(async (version: TemporalVersion) => {
    if (!version || isDeleting) return;
    
    try {
      setIsDeleting(true);
      
      // 使用DEACTIVATE事件而不是DELETE请求
      const response = await fetch(
        `http://localhost:9091/api/v1/organization-units/${organizationCode}/events`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            event_type: 'DEACTIVATE',
            effective_date: version.effective_date,
            change_reason: '通过时态管理页面作废版本'
          })
        }
      );
      
      if (response.ok) {
        // 刷新数据
        await loadVersions();
        setShowDeleteConfirm(null);
        
        // 如果作废的是选中的版本，重新选择
        if (selectedVersion?.effective_date === version.effective_date) {
          setSelectedVersion(null);
        }
      } else {
        console.error('Failed to deactivate version:', response.statusText);
        alert('作废失败，请稍后重试');
      }
    } catch (error) {
      console.error('Error deactivating version:', error);
      alert('作废失败，请检查网络连接');
    } finally {
      setIsDeleting(false);
    }
  }, [organizationCode, selectedVersion, isDeleting, loadVersions]);

  // 编辑功能处理
  const handleCreateVersion = useCallback(() => {
    setEditMode('create');
    setSelectedVersion(null);
    setActiveTab('new-version'); // 切换到新增版本选项卡，而不是打开Modal
  }, []);

  const handleEditVersion = useCallback((version: TemporalVersion) => {
    setEditMode('edit');
    setSelectedVersion(version);
    setShowEditForm(true);
  }, []);

  const handleFormSubmit = useCallback(async (formData: TemporalEditFormData) => {
    setIsSubmitting(true);
    try {
      const response = await fetch(
        `http://localhost:9091/api/v1/organization-units/${organizationCode}/events`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            event_type: formData.event_type,
            effective_date: new Date(formData.effective_date + 'T00:00:00Z').toISOString(),
            change_data: {
              name: formData.name,
              unit_type: formData.unit_type,
              status: formData.status,
              description: formData.description
            },
            change_reason: formData.change_reason
          })
        }
      );
      
      if (response.ok) {
        // 刷新数据
        await loadVersions();
        setActiveTab('details'); // 创建成功后切换回详情选项卡
        alert('时态版本创建成功！');
      } else {
        const errorData = await response.json();
        console.error('创建失败:', errorData);
        alert(`创建失败: ${errorData.message}`);
      }
    } catch (error) {
      console.error('创建时态版本失败:', error);
      alert('创建失败，请检查网络连接');
    } finally {
      setIsSubmitting(false);
    }
  }, [organizationCode, loadVersions]);

  const handleFormClose = useCallback(() => {
    if (!isSubmitting) {
      setActiveTab('details'); // 取消时切换回详情选项卡
      setSelectedVersion(null);
    }
  }, [isSubmitting]);

  // 组件挂载时加载数据
  useEffect(() => {
    loadVersions();
  }, [loadVersions]);

  return (
    <Box padding="l">
      {/* 页面头部 */}
      <Flex justifyContent="space-between" alignItems="center" marginBottom="l">
        <Box>
          <Heading size="large">时态管理 - {organizationCode}</Heading>
          <Text typeLevel="subtext.medium" color="hint">
            强制时间连续性的组织架构管理
          </Text>
        </Box>
        
        <Flex gap="s">
          <SecondaryButton onClick={loadVersions} disabled={isLoading}>
            🔄 刷新
          </SecondaryButton>
          {onBack && (
            <TertiaryButton onClick={onBack}>
              ← 返回
            </TertiaryButton>
          )}
        </Flex>
      </Flex>

      {/* 主从视图布局 */}
      <Flex gap="l" height="calc(100vh - 220px)">
        {/* 左侧：垂直交互式时间轴导航 */}
        <TimelineNavigation
          versions={versions}
          selectedVersion={selectedVersion}
          onVersionSelect={setSelectedVersion}
          onAddVersion={readonly ? undefined : handleCreateVersion}
          onDeleteVersion={readonly ? undefined : (version) => setShowDeleteConfirm(version)}
          isLoading={isLoading}
          readonly={readonly}
        />

        {/* 右侧：选项卡视图 */}
        <Box flex="1">
          {/* 选项卡头部 */}
          <Flex marginBottom="m" gap="s">
            <SecondaryButton
              size="small"
              onClick={() => setActiveTab('details')}
              style={{
                backgroundColor: activeTab === 'details' ? baseColors.blueberry[600] : 'transparent',
                color: activeTab === 'details' ? 'white' : baseColors.blueberry[600]
              }}
            >
              📋 版本详情
            </SecondaryButton>
            <SecondaryButton
              size="small"
              onClick={() => setActiveTab('timeline')}
              style={{
                backgroundColor: activeTab === 'timeline' ? baseColors.blueberry[600] : 'transparent',
                color: activeTab === 'timeline' ? 'white' : baseColors.blueberry[600]
              }}
            >
              📊 时间线可视化
            </SecondaryButton>
            <SecondaryButton
              size="small"
              onClick={() => setActiveTab('new-version')}
              style={{
                backgroundColor: activeTab === 'new-version' ? baseColors.greenFresca[600] : 'transparent',
                color: activeTab === 'new-version' ? 'white' : baseColors.greenFresca[600]
              }}
            >
              ➕ 新增版本
            </SecondaryButton>
          </Flex>

          {/* 选项卡内容 */}
          {activeTab === 'details' ? (
            <VersionDetailCard
              version={selectedVersion}
              onEdit={readonly ? undefined : handleEditVersion}
              onDelete={readonly ? undefined : (version) => setShowDeleteConfirm(version)}
              isLoading={isLoading}
              readonly={readonly}
            />
          ) : activeTab === 'timeline' ? (
            <SimpleTimelineVisualization
              organizationCode={organizationCode}
              onRefresh={loadVersions}
            />
          ) : (
            <InlineNewVersionForm
              organizationCode={organizationCode}
              onSubmit={handleFormSubmit}
              onCancel={handleFormClose}
              isSubmitting={isSubmitting}
            />
          )}
        </Box>
      </Flex>

      {/* 作废确认对话框 */}
      {showDeleteConfirm && (
        <Modal model={deleteModalModel}>
          <Modal.Overlay>
            <Modal.Card>
              <Modal.CloseIcon onClick={() => setShowDeleteConfirm(null)} />
              <Modal.Heading>确认作废版本</Modal.Heading>
              <Modal.Body>
                <Box padding="l">
            <Flex alignItems="flex-start" gap="m" marginBottom="l">
              <Box fontSize="24px" color={baseColors.cinnamon[600]}>⚠️</Box>
              <Box>
                <Text typeLevel="body.medium" marginBottom="s">
                  确定要作废生效日期为 <strong>{new Date(showDeleteConfirm.effective_date).toLocaleDateString('zh-CN')}</strong> 的版本吗？
                </Text>
                <Text typeLevel="subtext.small" color="hint" marginBottom="s">
                  版本名称: {showDeleteConfirm.name}
                </Text>
                <Text typeLevel="subtext.small" color={baseColors.cinnamon[600]}>
                  ⚠️ 作废后将自动填补时间空洞，此操作不可撤销
                </Text>
              </Box>
            </Flex>
            
            <Flex gap="s" justifyContent="flex-end">
              <SecondaryButton 
                onClick={() => setShowDeleteConfirm(null)}
                disabled={isDeleting}
              >
                取消
              </SecondaryButton>
              <PrimaryButton 
                onClick={() => handleDeleteVersion(showDeleteConfirm)}
                disabled={isDeleting}
              >
                {isDeleting ? '作废中...' : '确认作废'}
              </PrimaryButton>
            </Flex>
              </Box>
            </Modal.Body>
          </Modal.Card>
          </Modal.Overlay>
        </Modal>
      )}

      {/* 编辑表单 - 保留用于编辑现有版本 */}
      {editMode === 'edit' && (
        <TemporalEditForm
          isOpen={showEditForm}
          onClose={handleFormClose}
          onSubmit={handleFormSubmit}
          organizationCode={organizationCode}
          initialData={selectedVersion}
          mode={editMode}
          isSubmitting={isSubmitting}
        />
      )}
    </Box>
  );
};

export default TemporalMasterDetailView;