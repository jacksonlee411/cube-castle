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
import { baseColors } from '../../../shared/utils/colorTokens';
import { StatusBadge, type OrganizationStatus } from '../../../shared/components/StatusBadge';

// 状态映射函数：将后端状态映射到新的四状态系统
const mapBackendStatusToOrganizationStatus = (backendStatus: string): OrganizationStatus => {
  // 映射到新的四状态系统：ACTIVE, SUSPENDED, PLANNED, DELETED
  switch (backendStatus) {
    case 'ACTIVE':
      return 'ACTIVE';
    case 'INACTIVE':
    case 'SUSPENDED':
      return 'SUSPENDED';
    case 'PLANNED':
      return 'PLANNED';
    case 'DELETED':
      return 'DELETED';
    default:
      return 'ACTIVE'; // 默认状态
  }
};

// 状态映射函数：将后端状态映射到生命周期状态
// 移除：未使用的状态映射函数

// Types - 五状态生命周期管理系统
export interface TemporalVersion {
  recordId: string; // UUID唯一标识符
  code: string;
  name: string;
  unitType: string;
  status: string; // 组织状态：ACTIVE, INACTIVE, PLANNED
  effectiveDate: string;
  endDate?: string | null;
  changeReason?: string;
  isCurrent: boolean;
  createdAt: string;
  updatedAt: string;
  description?: string;
  level: number;
  path: string;
  parentCode?: string;
  sortOrder: number;
  
  // 五状态生命周期管理字段
  lifecycleStatus: 'CURRENT' | 'HISTORICAL' | 'PLANNED'; // 生命周期状态
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
  organizationCode: string | null; // 允许null用于创建模式
  readonly?: boolean;
  onBack?: () => void; // 返回回调
  onCreateSuccess?: (newOrganizationCode: string) => void; // 创建成功回调
  isCreateMode?: boolean; // 是否为创建模式
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
    switch (version.lifecycleStatus) {
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
    const start = formatDate(version.effectiveDate);
    
    // 优先检查删除状态
    if (version.data_status === 'DELETED' || version.status === 'DELETED') {
      return `${start} ~ 已删除`;
    }
    
    // 根据时态管理规则计算结束日期
    if (version.endDate) {
      // 如果有明确的结束日期，使用它
      return `${start} ~ ${formatDate(version.endDate)}`;
    }
    
    // 找到下一个生效日期更晚的版本（排除已删除的版本）
    const nextVersion = allVersions
      .filter(v => new Date(v.effectiveDate) > new Date(version.effectiveDate))
      .filter(v => v.data_status !== 'DELETED' && v.status !== 'DELETED')
      .sort((a, b) => new Date(a.effectiveDate).getTime() - new Date(b.effectiveDate).getTime())[0];
    
    if (nextVersion) {
      // 如果有下一个版本，当前版本的结束日期是下一个版本生效日期的前一天
      const nextDate = new Date(nextVersion.effectiveDate);
      nextDate.setDate(nextDate.getDate() - 1);
      return `${start} ~ ${formatDate(nextDate.toISOString().split('T')[0])}`;
    }
    
    // 如果没有下一个版本，根据生命周期状态决定显示内容
    if (version.lifecycleStatus === 'PLANNED') {
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
            const isSelected = selectedVersion?.effectiveDate === version.effectiveDate;
            
            return (
              <Box
                key={`${version.code}-${version.effectiveDate}`}
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
                          {formatDate(version.effectiveDate)}
                        </Text>
                        
                        {/* 状态标识 - 使用新的简化状态系统 */}
                        <StatusBadge 
                          status={mapBackendStatusToOrganizationStatus(version.status)} 
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
  readonly = false,
  onBack,
  onCreateSuccess,
  isCreateMode = false
}) => {
  // 状态管理
  const [versions, setVersions] = useState<TemporalVersion[]>([]);
  const [selectedVersion, setSelectedVersion] = useState<TemporalVersion | null>(null);
  const [isLoading, setIsLoading] = useState(!isCreateMode); // 创建模式不需要加载数据
  const [showDeleteConfirm, setShowDeleteConfirm] = useState<TemporalVersion | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);
  
  // 编辑表单状态
  const [showEditForm] = useState(isCreateMode); // 创建模式默认显示编辑表单
  const [editMode] = useState<'create' | 'edit'>(isCreateMode ? 'create' : 'edit');
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  // 视图选项卡状态 - 默认显示编辑历史记录页面
  const [activeTab, setActiveTab] = useState<'new-version' | 'edit-history'>('edit-history');
  
  // 表单模式状态 - 新增功能 (TODO: 当前未读取formMode值)
  const [/* formMode */, setFormMode] = useState<'create' | 'edit'>(isCreateMode ? 'create' : 'edit');
  const [formInitialData, setFormInitialData] = useState<{
    name: string;
    unitType: string;
    status: string;
    description?: string;
    parentCode?: string;
    effectiveDate?: string; // 添加生效日期
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
      
      // 使用organizationVersions查询获取完整的版本历史
      const response = await fetch('http://localhost:8090/graphql', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          query: `
            query GetOrganizationVersions($code: String!) {
              organizationVersions(code: $code) {
                code
                name
                unitType
                status
                level
                effectiveDate
                endDate
                isCurrent
                createdAt
                updatedAt
                recordId
                parentCode
                description
              }
            }
          `,
          variables: {
            code: organizationCode
          }
        })
      });
      
      if (response.ok) {
        const data = await response.json();
        const versions = data.data.organizationVersions || [];
        
        // 映射到组件需要的数据格式
        const mappedVersions = versions.map((version: any) => ({
          recordId: version.recordId,
          code: version.code,
          name: version.name,
          unitType: version.unitType,
          status: version.status,
          level: version.level,
          effectiveDate: version.effectiveDate,
          endDate: version.endDate,
          isCurrent: version.isCurrent,
          createdAt: version.createdAt,
          updatedAt: version.updatedAt,
          parentCode: version.parentCode,
          description: version.description,
          // 添加组件需要的字段
          lifecycleStatus: version.isCurrent ? 'CURRENT' : 'HISTORICAL',
          business_status: version.status === 'ACTIVE' ? 'ACTIVE' : 'SUSPENDED',
          data_status: 'NORMAL',
          path: '', // 临时字段，组件中需要
          sortOrder: 1, // 临时字段，组件中需要
          changeReason: '', // 临时字段，组件中需要
        }));
        
        const sortedVersions = mappedVersions.sort((a: any, b: any) => 
          new Date(b.effectiveDate).getTime() - new Date(a.effectiveDate).getTime()
        );
        setVersions(sortedVersions);
        
        // 默认选中当前版本
        const currentVersion = sortedVersions.find((v: any) => v.isCurrent);
        const defaultVersion = currentVersion || sortedVersions[0];
        
        if (defaultVersion) {
          setSelectedVersion(defaultVersion);
          
          // 预设表单数据（保持与现有表单字段格式兼容）
          setFormMode('edit');
          setFormInitialData({
            name: defaultVersion.name,
            unitType: defaultVersion.unitType,
            status: defaultVersion.status,
            description: defaultVersion.description || '',
            parentCode: defaultVersion.parentCode || '',
            effectiveDate: defaultVersion.effectiveDate
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
        `http://localhost:9090/api/v1/organization-units/${organizationCode}/events`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            eventType: 'DEACTIVATE',
            recordId: version.recordId,  // 使用UUID精确定位记录
            effectiveDate: version.effectiveDate,  // 保留用于日志和验证
            changeReason: '通过组织详情页面作废版本'
          })
        }
      );
      
      if (response.ok) {
        // 刷新数据
        await loadVersions();
        setShowDeleteConfirm(null);
        
        // 如果作废的是选中的版本，重新选择
        if (selectedVersion?.effectiveDate === version.effectiveDate) {
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
        unitType: version.unitType,
        status: version.status,
        description: version.description || '',
        parentCode: version.parentCode || '',
        effectiveDate: version.effectiveDate // 添加生效日期绑定
      });
    }
    
    // 如果当前在编辑历史记录选项卡，更新表单数据显示选中版本的信息
    if (activeTab === 'edit-history') {
      setFormMode('edit');
      setFormInitialData({
        name: version.name,
        unitType: version.unitType,
        status: version.status,
        description: version.description || '',
        parentCode: version.parentCode || '',
        effectiveDate: version.effectiveDate
      });
    }
  }, [activeTab]);

  const handleFormSubmit = useCallback(async (formData: TemporalEditFormData) => {
    setIsSubmitting(true);
    try {
      if (isCreateMode) {
        // 创建新组织
        // 状态映射：lifecycle_status -> API status
        // const mapLifecycleStatusToApiStatus = (lifecycleStatus: string) => { // TODO: 暂时未使用
        //   switch (lifecycleStatus) {
        //     case 'CURRENT': return 'ACTIVE';
        //     case 'PLANNED': return 'PLANNED';
        //     case 'HISTORICAL':
        //     case 'SUSPENDED':
        //     case 'DELETED': 
        //       return 'INACTIVE';
        //     default: 
        //       return 'ACTIVE';
        //   }
        // };
        
        const requestBody = {
          name: formData.name,
          unitType: formData.unitType,
          description: formData.description || '',
          parentCode: formData.parentCode || null,
          effectiveDate: formData.effectiveDate
        };
        
        console.log('提交创建组织请求:', requestBody);
        
        const response = await fetch('http://localhost:9090/api/v1/organization-units', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(requestBody)
        });
        
        console.log('API响应状态:', response.status, response.statusText);
        
        if (response.ok) {
          const result = await response.json();
          console.log('创建成功响应:', result);
          const newOrganizationCode = result.code || result.organization?.code;
          
          if (newOrganizationCode && onCreateSuccess) {
            console.log('跳转到新组织:', newOrganizationCode);
            // 触发创建成功回调，跳转到新创建的组织详情页面
            onCreateSuccess(newOrganizationCode);
            return; // 创建模式下不需要后续的刷新逻辑
          } else {
            console.error('创建成功但未返回组织编码:', result);
            alert('创建成功，但未能获取新组织编码，请手动刷新页面');
          }
        } else {
          const errorData = await response.json().catch(() => ({ message: response.statusText }));
          console.error('创建组织失败:', errorData);
          alert(`创建失败: ${errorData.message || response.statusText}`);
        }
      } else {
        // 更新现有组织的时态版本
        const response = await fetch(
          `http://localhost:9090/api/v1/organization-units/${organizationCode}/events`,
          {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              eventType: 'UPDATE',
              effectiveDate: new Date(formData.effectiveDate + 'T00:00:00Z').toISOString(),
              changeData: {
                name: formData.name,
                unitType: formData.unitType,
                status: formData.lifecycleStatus,
                description: formData.description,
                parentCode: formData.parentCode
              },
              changeReason: '通过组织信息详情页面更新组织信息'
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
      }
    } catch (error) {
      console.error(isCreateMode ? '创建组织失败:' : '创建时态版本失败:', error);
      alert(isCreateMode ? '创建失败，请检查网络连接' : '创建失败，请检查网络连接');
    } finally {
      setIsSubmitting(false);
    }
  }, [organizationCode, loadVersions, isCreateMode, onCreateSuccess]);

  const handleFormClose = useCallback(() => {
    if (!isSubmitting) {
      setActiveTab('edit-history'); // 取消时切换回历史记录选项卡
      setFormMode('create'); // 重置为新增模式
      setFormInitialData(null); // 清除预填充数据
      setSelectedVersion(null);
    }
  }, [isSubmitting]);

  // 历史记录编辑相关函数
  // const handleEditHistory = useCallback((version: TemporalVersion) => { // TODO: 暂时未使用
  //   setFormMode('edit');
  //   setFormInitialData({
  //     name: version.name,
  //     unitType: version.unitType,
  //     status: version.status,
  //     description: version.description || '',
  //     parent_code: version.parent_code || '',
  //     effectiveDate: version.effectiveDate
  //   });
  //   setSelectedVersion(version);
  //   setActiveTab('edit-history'); // 切换到历史记录编辑选项卡
  // }, []);

  const handleHistoryEditClose = useCallback(() => {
    if (!isSubmitting) {
      // 历史记录编辑页面关闭时应该返回组织列表页面
      if (onBack) {
        onBack();
      } else {
        // 回退方案：重置状态，但这不是预期的用户体验
        setActiveTab('edit-history');
        setFormMode('create');
        setFormInitialData(null);
      }
    }
  }, [isSubmitting, onBack]);

  const handleHistoryEditSubmit = useCallback(async (updateData: any) => {
    setIsSubmitting(true);
    try {
      // 使用recordId UUID作为唯一标识符
      const response = await fetch(
        `http://localhost:9090/api/v1/organization-units/${organizationCode}/history/${updateData.recordId}`,
        {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            name: updateData.name,
            unitType: updateData.unitType,
            status: updateData.status,
            description: updateData.description,
            effectiveDate: updateData.effectiveDate,
            parentCode: updateData.parentCode,
            changeReason: '通过组织详情页面修改历史记录'
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

  // 组件挂载时加载数据 - 创建模式跳过加载
  useEffect(() => {
    if (!isCreateMode && organizationCode) {
      loadVersions();
    }
  }, [loadVersions, isCreateMode, organizationCode]);

  // 获取当前版本的组织名称用于页面标题
  const getCurrentOrganizationName = () => {
    const currentVersion = versions.find(v => v.isCurrent);
    return currentVersion?.name || '';
  };

  return (
    <Box padding="l">
      {/* 页面头部 */}
      <Flex justifyContent="space-between" alignItems="center" marginBottom="l">
        <Box>
          <Heading size="large">
            {isCreateMode ? (
              '新建组织 - 编辑组织信息'
            ) : (
              `组织详情 - ${organizationCode}${getCurrentOrganizationName() ? ` ${getCurrentOrganizationName()}` : ''}`
            )}
          </Heading>
          <Text typeLevel="subtext.medium" color="hint">
            {isCreateMode ? '填写组织基本信息，系统将自动分配组织代码' : '强制时间连续性的组织架构管理'}
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
        {!isCreateMode && (
          <TimelineNavigation
            versions={versions}
            selectedVersion={selectedVersion}
            onVersionSelect={handleVersionSelect}
            onDeleteVersion={readonly ? undefined : (version) => setShowDeleteConfirm(version)}
            isLoading={isLoading}
            readonly={readonly}
          />
        )}

        {/* 创建模式下的提示区域 */}
        {isCreateMode && (
          <Box
            width="350px"
            height="calc(100vh - 200px)"
            backgroundColor="#F8F9FA"
            borderRadius={borderRadius.m}
            border="1px solid #E9ECEF"
            padding="m"
            style={{
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'center',
              alignItems: 'center'
            }}
          >
            <Box textAlign="center">
              <Text typeLevel="heading.small" marginBottom="m">
                创建新组织
              </Text>
              <Text typeLevel="body.medium" color="hint" marginBottom="l">
                填写右侧表单信息后，系统将自动分配组织编码并生成首个时态记录
              </Text>
              <Box
                width="60px"
                height="60px"
                borderRadius="50%"
                backgroundColor={colors.blueberry600}
                margin="auto"
                style={{
                  display: 'flex',
                  justifyContent: 'center',
                  alignItems: 'center'
                }}
              >
                <Text color="white" typeLevel="heading.medium">
                  +
                </Text>
              </Box>
            </Box>
          </Box>
        )}

        {/* 右侧：选项卡视图 */}
        <Box flex="1">
          {isCreateMode ? (
            // 创建模式：直接显示创建表单
            <InlineNewVersionForm
              organizationCode={null} // 创建模式下传入null
              onSubmit={handleFormSubmit}
              onCancel={() => {
                if (onBack) {
                  onBack(); // 创建模式下取消应该返回上一页
                }
              }}
              isSubmitting={isSubmitting}
              mode="create"
              initialData={null}
              selectedVersion={null}
              allVersions={null} // 创建模式不需要版本数据
            />
          ) : (
            // 统一的记录管理表单
            <InlineNewVersionForm
              organizationCode={organizationCode}
              onSubmit={handleFormSubmit}
              onCancel={handleHistoryEditClose}
              isSubmitting={isSubmitting}
              mode="edit"
              initialData={formInitialData}
              selectedVersion={selectedVersion}
              allVersions={versions.map(v => ({ // 传递版本数据用于日期范围验证
                recordId: v.recordId,
                effectiveDate: v.effectiveDate,
                endDate: v.endDate,
                isCurrent: v.isCurrent
              }))}
              onEditHistory={handleHistoryEditSubmit}
              onDeactivate={handleDeleteVersion} // 传递作废功能
              onInsertRecord={handleFormSubmit} // 传递插入记录功能
              activeTab={activeTab}
              onTabChange={setActiveTab}
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
                  确定要作废生效日期为 <strong>{new Date(showDeleteConfirm.effectiveDate).toLocaleDateString('zh-CN')}</strong> 的版本吗？
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
      {editMode === 'edit' && organizationCode && (
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