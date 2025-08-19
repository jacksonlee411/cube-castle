/**
 * 组织详情主从视图组件
 * 左侧：垂直交互式时间轴导航
 * 右侧：动态版本详情卡片
 * 实现强制时间连续性的完整用户体验
 */
import React, { useState, useCallback, useEffect } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text, Heading } from '@workday/canvas-kit-react/text';
import { Card } from '@workday/canvas-kit-react/card';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { Modal, useModalModel } from '@workday/canvas-kit-react/modal';
import TemporalEditForm, { type TemporalEditFormData } from './TemporalEditForm';
import { InlineNewVersionForm } from './InlineNewVersionForm';
import { LoadingDots } from '@workday/canvas-kit-react/loading-dots';
import { 
  colors, 
  borderRadius 
} from '@workday/canvas-kit-react/tokens';
import { plusIcon } from '@workday/canvas-system-icons-web';
import { baseColors } from '../../../shared/utils/colorTokens';
import { 
  LifecycleStatusBadge, 
  LIFECYCLE_STATES 
} from './FiveStateStatusSelector';

// 状态映射函数：将后端状态映射到前端五状态生命周期管理系统
const mapBackendStatusToLifecycleStatus = (backendStatus: string, isCurrent: boolean): 'CURRENT' | 'HISTORICAL' | 'PLANNED' => {
  // 根据后端状态和is_current标志确定生命周期状态
  if (backendStatus === 'PLANNED') {
    return 'PLANNED';
  } else if (backendStatus === 'ACTIVE' && isCurrent) {
    return 'CURRENT';
  } else {
    // INACTIVE 或 非当前的 ACTIVE 都被视为历史记录
    return 'HISTORICAL';
  }
};

// Types - 五状态生命周期管理系统
export interface TemporalVersion {
  record_id: string; // UUID唯一标识符
  code: string;
  name: string;
  unit_type: string;
  status: string; // 组织状态：ACTIVE, INACTIVE, PLANNED
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
  
  // 五状态生命周期管理字段
  lifecycle_status: 'CURRENT' | 'HISTORICAL' | 'PLANNED'; // 生命周期状态
  business_status: 'ACTIVE' | 'SUSPENDED'; // 业务状态
  data_status: 'NORMAL' | 'DELETED'; // 数据状态
  suspended_at?: string | null; // 停用时间
  suspended_by?: string | null; // 停用者
  suspension_reason?: string | null; // 停用原因
  deleted_at?: string | null; // 删除时间
  deleted_by?: string | null; // 删除者
  deletion_reason?: string | null; // 删除原因
}

export interface TemporalMasterDetailViewProps {
  organizationCode: string;
  readonly?: boolean;
}

/**
 * 左侧垂直时间轴导航区
 */
interface TimelineNavigationProps {
  versions: TemporalVersion[];
  selectedVersion: TemporalVersion | null;
  onVersionSelect: (version: TemporalVersion) => void;
  onDeleteVersion?: (version: TemporalVersion) => void;
  isLoading: boolean;
  readonly?: boolean;
}

const TimelineNavigation: React.FC<TimelineNavigationProps> = ({
  versions,
  selectedVersion,
  onVersionSelect,
  onDeleteVersion,
  isLoading,
  readonly = false
}) => {
  // 获取版本状态指示器 - 基于五状态生命周期管理系统
  const getVersionStatusIndicator = (version: TemporalVersion) => {
    // 1. 软删除状态（优先级最高）
    if (version.data_status === 'DELETED') {
      return { 
        color: colors.cinnamon600, 
        dotColor: colors.cinnamon600, 
        label: '已删除',
        isDeactivated: true,
        badge: 'DELETED' as const
      };
    }
    
    // 2. 业务停用状态
    if (version.business_status === 'SUSPENDED') {
      return { 
        color: colors.cantaloupe600, 
        dotColor: colors.cantaloupe600, 
        label: '已停用',
        isDeactivated: false,
        badge: 'SUSPENDED' as const
      };
    }
    
    // 3. 生命周期状态
    switch (version.lifecycle_status) {
      case 'CURRENT':
        return { 
          color: colors.greenApple500, 
          dotColor: colors.greenApple500, 
          label: '生效中',
          isDeactivated: false,
          badge: 'CURRENT' as const
        };
      case 'PLANNED':
        return { 
          color: colors.blueberry600, 
          dotColor: 'white', 
          label: '计划中',
          isDeactivated: false,
          badge: 'PLANNED' as const
        };
      case 'HISTORICAL':
        return { 
          color: colors.licorice400, 
          dotColor: colors.licorice400, 
          label: '历史记录',
          isDeactivated: false,
          badge: 'HISTORICAL' as const
        };
      default:
        return { 
          color: colors.licorice400, 
          dotColor: colors.licorice400, 
          label: '未知状态',
          isDeactivated: false,
          badge: 'HISTORICAL' as const
        };
    }
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('zh-CN');
  };

  const formatDateRange = (version: TemporalVersion, allVersions: TemporalVersion[]) => {
    const start = formatDate(version.effective_date);
    
    // 根据时态管理规则计算结束日期
    if (version.end_date) {
      // 如果有明确的结束日期，使用它
      return `${start} ~ ${formatDate(version.end_date)}`;
    }
    
    // 找到下一个生效日期更晚的版本
    const nextVersion = allVersions
      .filter(v => new Date(v.effective_date) > new Date(version.effective_date))
      .sort((a, b) => new Date(a.effective_date).getTime() - new Date(b.effective_date).getTime())[0];
    
    if (nextVersion) {
      // 如果有下一个版本，当前版本的结束日期是下一个版本生效日期的前一天
      const nextDate = new Date(nextVersion.effective_date);
      nextDate.setDate(nextDate.getDate() - 1);
      return `${start} ~ ${formatDate(nextDate.toISOString().split('T')[0])}`;
    }
    
    // 如果没有下一个版本，根据生命周期状态决定显示内容
    if (version.lifecycle_status === 'PLANNED') {
      // 计划中的记录显示"未来"
      return `${start} ~ 未来`;
    } else {
      // 当前记录或历史记录显示"至今"
      return `${start} ~ 至今`;
    }
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
                  backgroundColor={statusInfo.dotColor}
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
                          {formatDate(version.effective_date)}
                        </Text>
                        
                        {/* 五状态生命周期标识 */}
                        <LifecycleStatusBadge 
                          status={statusInfo.badge} 
                          size="small"
                        />
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



                    {/* 时间范围 */}
                    <Box>
                      <Text typeLevel="subtext.small" color="hint">
                        有效期间：
                      </Text>
                      <Text typeLevel="subtext.small" color="hint" marginLeft="xs">
                        {formatDateRange(version, versions)}
                      </Text>
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

/**
 * 组织详情主从视图主组件
 */
export const TemporalMasterDetailView: React.FC<TemporalMasterDetailViewProps> = ({
  organizationCode,
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
  
  // 视图选项卡状态 - 默认显示编辑历史记录页面
  const [activeTab, setActiveTab] = useState<'new-version' | 'edit-history'>('edit-history');
  
  // 表单模式状态 - 新增功能
  const [formMode, setFormMode] = useState<'create' | 'edit'>('create');
  const [formInitialData, setFormInitialData] = useState<{
    name: string;
    unit_type: string;
    status: string;
    description?: string;
    parent_code?: string;
    effective_date?: string; // 添加生效日期
  } | null>(null);

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

  // 加载时态版本数据 - 使用GraphQL查询符合CQRS架构
  const loadVersions = useCallback(async () => {
    try {
      setIsLoading(true);
      const response = await fetch('/graphql', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          query: `
            query GetOrganizationHistory($code: String!, $fromDate: String!, $toDate: String!) {
              organizationHistory(code: $code, fromDate: $fromDate, toDate: $toDate) {
                code
                record_id
                name
                unit_type
                status
                level
                path
                description
                effective_date
                end_date
                change_reason
                is_current
                created_at
                updated_at
              }
            }
          `,
          variables: {
            code: organizationCode,
            fromDate: '1900-01-01',  // 扩展到1900年以包含所有历史数据
            toDate: '2030-12-31'
          }
        })
      });
      
      if (response.ok) {
        const data = await response.json();
        
        // 映射GraphQL数据到前端五状态生命周期管理系统
        const mappedVersions = data.data.organizationHistory.map((version: any) => ({
          ...version,
          // 添加五状态生命周期字段映射
          lifecycle_status: mapBackendStatusToLifecycleStatus(version.status, version.is_current),
          business_status: version.status === 'SUSPENDED' ? 'SUSPENDED' : 'ACTIVE',
          data_status: 'NORMAL' // 默认为正常状态，除非后端明确标记为删除
        }));
        
        const sortedVersions = mappedVersions.sort((a: TemporalVersion, b: TemporalVersion) => 
          new Date(b.effective_date).getTime() - new Date(a.effective_date).getTime()
        );
        setVersions(sortedVersions);
        
        // 默认选中当前版本
        const currentVersion = sortedVersions.find((v: TemporalVersion) => v.is_current);
        const defaultVersion = currentVersion || sortedVersions[0];
        
        if (defaultVersion) {
          setSelectedVersion(defaultVersion);
          
          // 由于默认显示编辑历史记录页面，需要预设表单数据
          setFormMode('edit');
          setFormInitialData({
            name: defaultVersion.name,
            unit_type: defaultVersion.unit_type,
            status: defaultVersion.status,
            description: defaultVersion.description || '',
            parent_code: defaultVersion.parent_code || '',
            effective_date: defaultVersion.effective_date
          });
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
            change_reason: '通过组织详情页面作废版本'
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

  // 时间轴版本选择处理 - 增强功能，支持编辑历史记录页面联动
  const handleVersionSelect = useCallback((version: TemporalVersion) => {
    setSelectedVersion(version);
    
    // 如果当前在新增版本选项卡，自动预填充选中版本的数据
    if (activeTab === 'new-version') {
      setFormMode('edit');
      setFormInitialData({
        name: version.name,
        unit_type: version.unit_type,
        status: version.status,
        description: version.description || '',
        parent_code: version.parent_code || '',
        effective_date: version.effective_date // 添加生效日期绑定
      });
    }
    
    // 如果当前在编辑历史记录选项卡，更新表单数据显示选中版本的信息
    if (activeTab === 'edit-history') {
      setFormMode('edit');
      setFormInitialData({
        name: version.name,
        unit_type: version.unit_type,
        status: version.status,
        description: version.description || '',
        parent_code: version.parent_code || '',
        effective_date: version.effective_date
      });
    }
  }, [activeTab]);

  const handleFormSubmit = useCallback(async (formData: TemporalEditFormData) => {
    setIsSubmitting(true);
    try {
      const response = await fetch(
        `http://localhost:9091/api/v1/organization-units/${organizationCode}/events`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            event_type: 'UPDATE',
            effective_date: new Date(formData.effective_date + 'T00:00:00Z').toISOString(),
            change_data: {
              name: formData.name,
              unit_type: formData.unit_type,
              status: formData.status,
              description: formData.description,
              parent_code: formData.parent_code
            },
            change_reason: '通过组织信息详情页面更新组织信息'
          })
        }
      );
      
      if (response.ok) {
        // 刷新数据
        await loadVersions();
        setActiveTab('edit-history'); // 创建成功后切换回历史记录选项卡
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
      setActiveTab('edit-history'); // 取消时切换回历史记录选项卡
      setFormMode('create'); // 重置为新增模式
      setFormInitialData(null); // 清除预填充数据
      setSelectedVersion(null);
    }
  }, [isSubmitting]);

  // 历史记录编辑相关函数
  const handleEditHistory = useCallback((version: TemporalVersion) => {
    setFormMode('edit');
    setFormInitialData({
      name: version.name,
      unit_type: version.unit_type,
      status: version.status,
      description: version.description || '',
      parent_code: version.parent_code || '',
      effective_date: version.effective_date
    });
    setSelectedVersion(version);
    setActiveTab('edit-history'); // 切换到历史记录编辑选项卡
  }, []);

  const handleHistoryEditClose = useCallback(() => {
    if (!isSubmitting) {
      setActiveTab('edit-history'); // 取消时切换回历史记录选项卡
      setFormMode('create');
      setFormInitialData(null);
      // 保持selectedVersion，以便返回详情页面
    }
  }, [isSubmitting]);

  const handleHistoryEditSubmit = useCallback(async (updateData: any) => {
    setIsSubmitting(true);
    try {
      // 使用record_id UUID作为唯一标识符
      const response = await fetch(
        `http://localhost:9091/api/v1/organization-units/history/${updateData.record_id}`,
        {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            name: updateData.name,
            unit_type: updateData.unit_type,
            status: updateData.status,
            description: updateData.description,
            effective_date: updateData.effective_date,
            parent_code: updateData.parent_code,
            change_reason: '通过组织详情页面修改历史记录'
          })
        }
      );
      
      if (response.ok) {
        // 刷新数据
        await loadVersions();
        setActiveTab('edit-history'); // 提交成功后切换回历史记录选项卡
        alert('历史记录修改成功！');
      } else {
        const errorData = await response.json();
        console.error('修改失败:', errorData);
        alert(`修改失败: ${errorData.message || response.statusText}`);
      }
    } catch (error) {
      console.error('修改历史记录失败:', error);
      alert('修改失败，请检查网络连接');
    } finally {
      setIsSubmitting(false);
    }
  }, [organizationCode, loadVersions]);

  // 组件挂载时加载数据
  useEffect(() => {
    loadVersions();
  }, [loadVersions]);

  // 获取当前版本的组织名称用于页面标题
  const getCurrentOrganizationName = () => {
    const currentVersion = versions.find(v => v.is_current);
    return currentVersion?.name || '';
  };

  return (
    <Box padding="l">
      {/* 页面头部 */}
      <Flex justifyContent="space-between" alignItems="center" marginBottom="l">
        <Box>
          <Heading size="large">
            组织详情 - {organizationCode}
            {getCurrentOrganizationName() && ` ${getCurrentOrganizationName()}`}
          </Heading>
          <Text typeLevel="subtext.medium" color="hint">
            强制时间连续性的组织架构管理
          </Text>
        </Box>
        
        <Flex gap="s">
          <SecondaryButton onClick={loadVersions} disabled={isLoading}>
            刷新
          </SecondaryButton>
        </Flex>
      </Flex>

      {/* 主从视图布局 */}
      <Flex gap="l" height="calc(100vh - 220px)">
        {/* 左侧：垂直交互式时间轴导航 */}
        <TimelineNavigation
          versions={versions}
          selectedVersion={selectedVersion}
          onVersionSelect={handleVersionSelect}
          onDeleteVersion={readonly ? undefined : (version) => setShowDeleteConfirm(version)}
          isLoading={isLoading}
          readonly={readonly}
        />

        {/* 右侧：选项卡视图 */}
        <Box flex="1">
          {/* 选项卡头部 */}
          <Flex marginBottom="m" gap="s">
            <PrimaryButton
              size="small"
              onClick={() => setActiveTab('edit-history')}
            >
              编辑历史记录
            </PrimaryButton>
            <SecondaryButton
              size="small"
              onClick={() => setActiveTab('new-version')}
              icon={plusIcon}
            >
              编辑组织信息
            </SecondaryButton>
          </Flex>

          {/* 选项卡内容 */}
          {activeTab === 'edit-history' ? (
            // 历史记录编辑模式
            <InlineNewVersionForm
              organizationCode={organizationCode}
              onSubmit={handleFormSubmit} // 先使用现有的函数，稍后更新
              onCancel={handleHistoryEditClose}
              isSubmitting={isSubmitting}
              mode="edit-history"
              initialData={formInitialData}
              selectedVersion={selectedVersion}
              onEditHistory={handleHistoryEditSubmit}
              onDeactivate={handleDeleteVersion} // 传递作废功能
            />
          ) : (
            // 编辑组织信息模式
            <InlineNewVersionForm
              organizationCode={organizationCode}
              onSubmit={handleFormSubmit}
              onCancel={handleFormClose}
              isSubmitting={isSubmitting}
              mode={formMode}
              initialData={formInitialData}
              selectedVersion={selectedVersion}
              onEditHistory={handleEditHistory}
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
              <Box fontSize="24px" color={baseColors.cinnamon[600]}>警告</Box>
              <Box>
                <Text typeLevel="body.medium" marginBottom="s">
                  确定要作废生效日期为 <strong>{new Date(showDeleteConfirm.effective_date).toLocaleDateString('zh-CN')}</strong> 的版本吗？
                </Text>
                <Text typeLevel="subtext.small" color="hint" marginBottom="s">
                  版本名称: {showDeleteConfirm.name}
                </Text>
                <Text typeLevel="subtext.small" color={baseColors.cinnamon[600]}>
                  警告 作废后将自动填补时间空洞，此操作不可撤销
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