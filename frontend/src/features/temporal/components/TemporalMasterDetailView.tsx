/**
 * 组织详情主从视图组件
 * 左侧：垂直交互式时间轴导航
 * 右侧：动态版本详情卡片
 * 实现强制时间连续性的完整用户体验
 */
import React, { useState, useCallback, useEffect } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text, Heading } from '@workday/canvas-kit-react/text';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { Card } from '@workday/canvas-kit-react/card';
import { Modal, useModalModel } from '@workday/canvas-kit-react/modal';
import { checkCircleIcon, exclamationCircleIcon, activityStreamIcon } from '@workday/canvas-system-icons-web';
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import TemporalEditForm, { type TemporalEditFormData } from './TemporalEditForm';
import { InlineNewVersionForm } from './InlineNewVersionForm';
import { TimelineComponent, type TimelineVersion } from './TimelineComponent';
import { TabNavigation, type TabType } from './TabNavigation';
import { 
  colors, 
  borderRadius 
} from '@workday/canvas-kit-react/tokens';
import { baseColors } from '../../../shared/utils/colorTokens';
import { unifiedGraphQLClient, unifiedRESTClient } from '../../../shared/api/unified-client';
import { env } from '../../../shared/config/environment';
import { useNavigate } from 'react-router-dom';
// 审计历史组件导入
import { AuditHistorySection } from '../../audit/components/AuditHistorySection';
import { normalizeParentCode } from '../../../shared/utils/organization-helpers';
import { OrganizationBreadcrumb } from '../../../shared/components/OrganizationBreadcrumb';

// 使用来自TimelineComponent的TimelineVersion类型
// export interface TemporalVersion 已移动到 TimelineComponent.tsx

// Organization Version interface
interface OrganizationVersion {
  code: string;
  name: string;
  unitType: string;
  status: string;
  level: number;
  parentCode?: string;
  description?: string;
  sortOrder: number;
  path?: string | null;
  effectiveDate: string;
  endDate?: string | null;
  recordId: string;
  createdAt: string;
  updatedAt: string;
}

export interface TemporalMasterDetailViewProps {
  organizationCode: string | null; // 允许null用于创建模式
  readonly?: boolean;
  onBack?: () => void; // 返回回调
  onCreateSuccess?: (newOrganizationCode: string) => void; // 创建成功回调
  isCreateMode?: boolean; // 是否为创建模式
}

// TimelineNavigationProps 已移动到 TimelineComponent.tsx 作为 TimelineComponentProps

// TimelineNavigation组件已提取为独立的TimelineComponent.tsx

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
  const navigate = useNavigate();
  // 状态管理
  const [versions, setVersions] = useState<TimelineVersion[]>([]);
  const [selectedVersion, setSelectedVersion] = useState<TimelineVersion | null>(null);
  const [isLoading, setIsLoading] = useState(!isCreateMode); // 创建模式不需要加载数据
  const [showDeleteConfirm, setShowDeleteConfirm] = useState<TimelineVersion | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);
  
  // 增强用户体验状态
  const [loadingError, setLoadingError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [retryCount, setRetryCount] = useState(0);
  
  // 编辑表单状态
  const [showEditForm] = useState(isCreateMode); // 创建模式默认显示编辑表单
  const [editMode] = useState<'create' | 'edit'>(isCreateMode ? 'create' : 'edit');
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  // 视图选项卡状态 - 默认显示版本历史页面，支持审计信息
  const [activeTab, setActiveTab] = useState<TabType>('edit-history');
  
  // TODO-TEMPORARY: FormMode state is not used; integrate form mode logic in v4.3 by 2025-09-20.
  const [/* formMode */, setFormMode] = useState<'create' | 'edit'>(isCreateMode ? 'create' : 'edit');
  const [formInitialData, setFormInitialData] = useState<{
    name: string;
    unitType: string;
    status: string;
    description?: string;
    parentCode?: string;
    effectiveDate?: string; // 添加生效日期
  } | null>(null);

  // 展示路径（来自层级查询）
  const [displayPaths, setDisplayPaths] = useState<{ codePath: string; namePath: string } | null>(null);

  // Modal model for delete confirmation
  const deleteModalModel = useModalModel();
  
  // 统一的消息处理函数
  const showSuccess = useCallback((message: string) => {
    setError(null);
    setSuccessMessage(message);
    // 3秒后自动清除成功消息
    setTimeout(() => setSuccessMessage(null), 3000);
  }, []);
  
  const showError = useCallback((message: string) => {
    setSuccessMessage(null);
    setError(message);
    // 5秒后自动清除错误消息
    setTimeout(() => setError(null), 5000);
  }, []);

  // 同步Modal状态
  React.useEffect(() => {
    if (showDeleteConfirm && deleteModalModel.state.visibility !== 'visible') {
      deleteModalModel.events.show();
    } else if (!showDeleteConfirm && deleteModalModel.state.visibility === 'visible') {
      deleteModalModel.events.hide();
    }
  }, [showDeleteConfirm, deleteModalModel]);

  // 加载时态版本数据 - 增强版本，包含错误处理和重试机制
  const loadVersions = useCallback(async (isRetry = false) => {
    try {
      setIsLoading(true);
      setLoadingError(null);
      if (!isRetry) {
        setRetryCount(0);
      }
      
      // 使用新的organizationVersions查询获取全部版本（按生效日排序）
      let data;
      try {
        data = await unifiedGraphQLClient.request<{
          organizationVersions: OrganizationVersion[];
        }>(`
          query OrganizationVersions($code: String!) {
            organizationVersions(code: $code) {
              recordId
              code
              name
              unitType
              status
              level
              path
              effectiveDate
              endDate
              createdAt
              updatedAt
              parentCode
              description
            }
          }
        `, {
          code: organizationCode
        });
      } catch (graphqlError: unknown) {
        // 回退策略：新查询不可用时，回退到现有"单体快照"逻辑
        console.warn('organizationVersions查询失败，回退到单体快照逻辑:', graphqlError);
        try {
          data = await unifiedGraphQLClient.request<{
            organization: OrganizationVersion | null;
          }>(`
            query GetOrganization($code: String!) {
              organization(code: $code) {
                code
                name
                unitType
                status
                level
                path
                effectiveDate
                endDate
                createdAt
                updatedAt
                recordId
                parentCode
                description
                hierarchyDepth
              }
            }
          `, {
            code: organizationCode
          });

          // 将单个组织转换为数组格式用于后续处理
          if (data?.organization) {
            data = {
              organizationVersions: [data.organization]
            };
          } else {
            data = { organizationVersions: [] };
          }

          // 显示回退提示
          setLoadingError('历史列表不可用，展示当前快照');
          setTimeout(() => setLoadingError(null), 3000);
        } catch (fallbackError) {
          interface GraphQLErrorWithResponse {
            response?: {
              status: number;
              statusText?: string;
            };
            message?: string;
          }

          const typedError = fallbackError as GraphQLErrorWithResponse;
          if (typedError?.response?.status) {
            const statusCode = typedError.response.status;
            const statusText = typedError.response.statusText || 'Unknown Error';
            throw new Error(`服务器响应错误 (${statusCode}): ${statusText}`);
          }
          throw new Error(`GraphQL调用失败: ${typedError?.message || '未知错误'}`);
        }
      }
        
      // 保留数据验证 - 防御性编程
      if (!data) {
        throw new Error('GraphQL响应为空');
      }
      
      // 处理版本数组数据
      const organizations = data.organizationVersions;
      if (!Array.isArray(organizations)) {
        throw new Error('版本数据格式错误');
      }

      // 将版本数组直接 map 为 TimelineVersion[]，按 effectiveDate ASC（服务端已排序）
      const mappedVersions: TimelineVersion[] = organizations.map((org: OrganizationVersion) => ({
        recordId: org.recordId,
        code: org.code,
        name: org.name,
        unitType: org.unitType,
        status: org.status,
        level: org.level,
        effectiveDate: org.effectiveDate,
        endDate: org.endDate,
        isCurrent: org.endDate === null, // 当前版本: endDate为null
        createdAt: org.createdAt,
        updatedAt: org.updatedAt,
        parentCode: org.parentCode,
        description: org.description,
        // 添加组件需要的字段
        lifecycleStatus: org.endDate === null ? 'CURRENT' as const : 'HISTORICAL' as const,
        businessStatus: org.status === 'ACTIVE' ? 'ACTIVE' : 'INACTIVE',
        dataStatus: 'NORMAL' as const,
        path: org.path ?? undefined,
        sortOrder: 1, // 临时字段，组件中需要
        changeReason: '' // 临时字段，组件中需要
      }));

      // 按 effectiveDate 降序排序（最新版本在上方）
      const sortedVersions = mappedVersions.sort((a: TimelineVersion, b: TimelineVersion) =>
        new Date(b.effectiveDate).getTime() - new Date(a.effectiveDate).getTime()
      );
      setVersions(sortedVersions);

      // 显示成功消息
      if (isRetry) {
        setSuccessMessage('数据加载成功！');
        setTimeout(() => setSuccessMessage(null), 3000);
      }

      // 选中当前版本 = "生效日 ≤ 今日"的最大者
      const currentVersion = sortedVersions.find((v: TimelineVersion) => v.isCurrent) || sortedVersions.at(-1) || null;

      if (currentVersion) {
        setSelectedVersion(currentVersion);

        // 预设表单数据（保持与现有表单字段格式兼容）
        setFormMode('edit');
        setFormInitialData({
          name: currentVersion.name,
          unitType: currentVersion.unitType,
          status: currentVersion.status,
          description: currentVersion.description || '',
          parentCode: normalizeParentCode.forForm(currentVersion.parentCode),
          effectiveDate: currentVersion.effectiveDate
        });

        // 加载层级路径信息（codePath/namePath）
        try {
          const pathData = await unifiedGraphQLClient.request<{ organizationHierarchy: { codePath: string; namePath: string } | null }>(
            `query GetHierarchyPaths($code: String!, $tenantId: String!) {
               organizationHierarchy(code: $code, tenantId: $tenantId) {
                 codePath
                 namePath
               }
             }`,
            { code: currentVersion.code, tenantId: env.defaultTenantId }
          );
          const hierarchy = pathData?.organizationHierarchy;
          if (hierarchy) {
            setDisplayPaths({ codePath: hierarchy.codePath, namePath: hierarchy.namePath });
          } else {
            setDisplayPaths(null);
          }
        } catch (e) {
          console.warn('加载组织层级路径失败（忽略，不阻塞详情展示）:', e);
          setDisplayPaths(null);
        }
      }
      
    } catch (error) {
      console.error('Error loading temporal versions:', error);
      const errorMessage = error instanceof Error 
        ? error.message 
        : '加载版本数据时发生未知错误';
      setLoadingError(errorMessage);
      setRetryCount(prev => prev + 1);
    } finally {
      setIsLoading(false);
    }
  }, [organizationCode]);

  // 作废版本处理
  const handleDeleteVersion = useCallback(async (version: TimelineVersion) => {
    if (!version || isDeleting) return;
    
    try {
      setIsDeleting(true);
      
      // 使用DEACTIVATE事件而不是DELETE请求 - 修复：使用统一认证客户端
      const resp = await unifiedRESTClient.request(
        `/organization-units/${organizationCode}/events`,
        {
          method: 'POST',
          body: JSON.stringify({
            eventType: 'DEACTIVATE',
            recordId: version.recordId,  // 使用UUID精确定位记录
            effectiveDate: version.effectiveDate,  // 保留用于日志和验证
            changeReason: '通过组织详情页面作废版本'
          })
        }
      ) as { data?: { timeline?: Record<string, unknown>[] } };

      // 删除API调用成功后，优先使用后端返回的新时间线，避免读缓存延迟
          const timeline = resp?.data?.timeline;
          if (Array.isArray(timeline)) {
        const mappedVersions: TimelineVersion[] = timeline.map((v: Record<string, unknown>) => {
          const isCurrent = (v.endDate as string) === null || v.endDate === undefined;
          return {
            recordId: v.recordId as string,
            code: v.code as string,
            name: v.name as string,
            unitType: v.unitType as string,
            status: v.status as string,
            level: v.level as number,
            effectiveDate: v.effectiveDate as string,
            endDate: (v.endDate as string) || null,
            isCurrent: isCurrent,
            createdAt: v.createdAt as string,
            updatedAt: v.updatedAt as string,
            parentCode: (v.parentCode as string) || undefined,
            description: (v.description as string) || undefined,
            lifecycleStatus: isCurrent ? 'CURRENT' : 'HISTORICAL',
            businessStatus: v.status === 'ACTIVE' ? 'ACTIVE' : 'INACTIVE',
            dataStatus: 'NORMAL',
            path: (v.path as string | undefined) ?? undefined,
            sortOrder: 1,
            changeReason: '',
          };
        });
        const sorted = mappedVersions.sort((a, b) => new Date(b.effectiveDate).getTime() - new Date(a.effectiveDate).getTime());
        setVersions(sorted);
        const current = sorted.find(v => v.isCurrent) || sorted[0] || null;
        setSelectedVersion(current);
      } else {
        // 回退：如果后端未返回时间线，执行原有刷新逻辑
        try {
          await loadVersions();
        } catch (refreshError) {
          console.warn('数据刷新失败，但删除操作已成功:', refreshError);
        }
      }

      // 删除API调用成功后，处理UI状态
      setShowDeleteConfirm(null);
      // 如果作废的是选中的版本，重新选择（已在返回时间线路径中处理）
      if (!timeline && selectedVersion?.effectiveDate === version.effectiveDate) {
        setSelectedVersion(null);
      }
    } catch (error) {
      console.error('Error deactivating version:', error);
      // 不再在这里显示错误消息，让错误向上传播到调用者
      // 这样可以避免双重错误消息显示
      throw error; // 重新抛出错误，让上层组件处理
    } finally {
      setIsDeleting(false);
    }
  }, [organizationCode, selectedVersion, isDeleting, loadVersions]);

  // 时间轴版本选择处理 - 增强功能，支持编辑历史记录页面联动
  const handleVersionSelect = useCallback((version: TimelineVersion) => {
    setSelectedVersion(version);
    
    // 如果当前在版本历史选项卡，更新表单数据显示选中版本的信息
    if (activeTab === 'edit-history') {
      setFormMode('edit');
      setFormInitialData({
        name: version.name,
        unitType: version.unitType,
        status: version.status,
        description: version.description || '',
        parentCode: normalizeParentCode.forForm(version.parentCode),
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
        // TODO-TEMPORARY: mapLifecycleStatusToApiStatus not implemented; add status mapping in v4.3 by 2025-09-20.
        // const mapLifecycleStatusToApiStatus = (lifecycleStatus: string) => {
        //   switch (lifecycleStatus) {
        //     case 'CURRENT': return 'ACTIVE';
        //     case 'PLANNED': return 'PLANNED';
        //     case 'HISTORICAL':
        //     case 'INACTIVE':
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
          parentCode: normalizeParentCode.forAPI(formData.parentCode),
          effectiveDate: formData.effectiveDate
        };
        
        console.log('提交创建组织请求:', requestBody);
        
        // 修复：使用统一认证客户端替代直接fetch调用
        const result: Record<string, unknown> = await unifiedRESTClient.request('/organization-units', {
          method: 'POST',
          body: JSON.stringify(requestBody)
        });
        
        console.log('创建成功响应:', result);
        interface CreateResult {
          code?: string;
          organization?: {
            code?: string;
          };
          data?: {
            code?: string;
          };
        }
        
        const typedResult = result as CreateResult;
        const newOrganizationCode = typedResult.data?.code || typedResult.code || typedResult.organization?.code;
        
        if (newOrganizationCode && onCreateSuccess) {
          console.log('跳转到新组织:', newOrganizationCode);
          // 触发创建成功回调，跳转到新创建的组织详情页面
          onCreateSuccess(newOrganizationCode);
          return; // 创建模式下不需要后续的刷新逻辑
        } else {
          console.error('创建成功但未返回组织编码:', result);
          showError('创建成功，但未能获取新组织编码，请手动刷新页面');
        }
      } else {
        // 为现有组织创建新的时态版本 - 使用REST API (命令操作)
        await unifiedRESTClient.request(`/organization-units/${organizationCode}/versions`, {
          method: 'POST',
          body: JSON.stringify({
            name: formData.name,
            unitType: formData.unitType,
            parentCode: normalizeParentCode.forAPI(formData.parentCode),
            description: formData.description || null,
            effectiveDate: formData.effectiveDate, // 使用YYYY-MM-DD格式
            operationReason: formData.changeReason || '通过组织详情页面创建新版本'
          })
        });
        
        // unifiedRESTClient成功时直接返回数据，失败时抛出异常
        // 刷新数据
        await loadVersions();
        setActiveTab('edit-history'); // 创建成功后切换回历史记录选项卡
        showSuccess('时态版本创建成功！');
      }
    } catch (error) {
      console.error(isCreateMode ? '创建组织失败:' : '创建时态版本失败:', error);
      
      // 提取实际的错误信息
      let errorMessage = isCreateMode ? '创建组织失败' : '创建时态版本失败';
      if (error instanceof Error) {
        errorMessage = error.message;
      } else if (error && typeof error === 'string') {
        errorMessage = error;
      }
      
      showError(errorMessage);
    } finally {
      setIsSubmitting(false);
    }
  }, [organizationCode, loadVersions, isCreateMode, onCreateSuccess, showError, showSuccess]);

  const handleFormClose = useCallback(() => {
    if (!isSubmitting) {
      setActiveTab('edit-history'); // 取消时切换回历史记录选项卡
      setFormMode('create'); // 重置为新增模式
      setFormInitialData(null); // 清除预填充数据
      setSelectedVersion(null);
    }
  }, [isSubmitting]);

  // 历史记录编辑相关函数
  // TODO-TEMPORARY: handleEditHistory not implemented; add history editing functionality in v4.3 by 2025-09-20.
  // const handleEditHistory = useCallback((version: TimelineVersion) => {
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

  const handleHistoryEditSubmit = useCallback(async (updateData: Record<string, unknown>) => {
    setIsSubmitting(true);
    try {
      // 使用recordId UUID作为唯一标识符 - 修复：使用统一认证客户端
      await unifiedRESTClient.request(
        `/organization-units/${organizationCode}/history/${updateData.recordId}`,
        {
          method: 'PUT',
          body: JSON.stringify({
            name: updateData.name,
            unitType: updateData.unitType,
            status: updateData.status,
            description: updateData.description,
            effectiveDate: updateData.effectiveDate,
            parentCode: normalizeParentCode.forAPI(updateData.parentCode as string),
            changeReason: '通过组织详情页面修改历史记录'
          })
        }
      );
      
      // unifiedRESTClient成功时直接返回数据，失败时抛出异常
      // 刷新数据
      await loadVersions();
      setActiveTab('edit-history'); // 提交成功后切换回历史记录选项卡
      showSuccess('历史记录修改成功！');
    } catch (error) {
      console.error('修改历史记录失败:', error);
      showError('修改失败，请检查网络连接');
    } finally {
      setIsSubmitting(false);
    }
  }, [organizationCode, loadVersions, showError, showSuccess]);

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
          {/* 路径面包屑（名称优先） */}
          {!isCreateMode && displayPaths && (
            <Box marginTop="s">
              <OrganizationBreadcrumb
                codePath={displayPaths.codePath}
                namePath={displayPaths.namePath}
                separator="/"
                onNavigate={(code) => {
                  if (code) {
                    navigate(`/organizations/${code}/temporal`);
                  }
                }}
              />
            </Box>
          )}
        </Box>
        
        <Flex gap="s">
          <SecondaryButton 
            onClick={() => loadVersions()} 
            disabled={isLoading}
          >
            {isLoading ? '刷新中...' : '刷新'}
          </SecondaryButton>
        </Flex>
      </Flex>

      {/* 状态消息区域 */}
      {(loadingError || error || successMessage) && (
        <Box marginBottom="l">
          {(loadingError || error) && (
            <Box
              padding="m"
              backgroundColor={colors.cinnamon100}
              border={`1px solid ${colors.cinnamon600}`}
              borderRadius={borderRadius.m}
              marginBottom="s"
            >
              <Flex alignItems="center" gap="s">
                <SystemIcon icon={exclamationCircleIcon} color={colors.cinnamon600} size="small" />
                <Box flex="1">
                  <Text color={colors.cinnamon600} typeLevel="body.small" fontWeight="medium">
                    {loadingError ? '加载失败' : '操作失败'}
                  </Text>
                  <Text color={colors.cinnamon600} typeLevel="subtext.small">
                    {loadingError || error}
                  </Text>
                </Box>
                {loadingError && retryCount < 3 && (
                  <SecondaryButton
                    size="small"
                    onClick={() => loadVersions(true)}
                    disabled={isLoading}
                  >
                    重试 ({retryCount}/3)
                  </SecondaryButton>
                )}
              </Flex>
            </Box>
          )}
          
          {successMessage && (
            <Box
              padding="m"
              backgroundColor={colors.greenApple100}
              border={`1px solid ${colors.greenApple600}`}
              borderRadius={borderRadius.m}
              marginBottom="s"
            >
              <Flex alignItems="center" gap="s">
                <SystemIcon icon={checkCircleIcon} color={colors.greenApple600} size="small" />
                <Text color={colors.greenApple600} typeLevel="body.small" fontWeight="medium">
                  {successMessage}
                </Text>
              </Flex>
            </Box>
          )}
        </Box>
      )}

      {/* 主从视图布局 */}
      <Flex gap="l" height="calc(100vh - 220px)">
        {/* 左侧：垂直交互式时间轴导航 */}
        {!isCreateMode && (
          <TimelineComponent
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
              hierarchyPaths={displayPaths}
            />
          ) : (
            // 正常模式：带选项卡的多功能视图
            <>
              {/* 选项卡导航 */}
              <TabNavigation
                activeTab={activeTab}
                onTabChange={setActiveTab}
                disabled={isSubmitting || isLoading}
                tabs={[
                  { key: 'edit-history', label: '版本历史' },
                  { key: 'audit-history', label: '审计历史' }
                ]}
              />

              {/* 选项卡内容 */}
              {activeTab === 'edit-history' && (
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
                  onDeactivate={async (version: Record<string, unknown>) => {
                    try {
                      // 类型安全转换
                      const typedVersion = version as unknown as TimelineVersion;
                      await handleDeleteVersion(typedVersion);
                      // 成功时的处理由handleDeleteVersion内部完成
                    } catch (error) {
                      // 显示错误消息
                      const errorMessage = error instanceof Error ? error.message : '作废失败，请重试';
                      showError(errorMessage);
                    }
                  }} // 传递作废功能
                  onInsertRecord={handleFormSubmit} // 传递插入记录功能
                  activeTab="edit-history"
                  onTabChange={setActiveTab}
                  hierarchyPaths={displayPaths}
                />
              )}


              {/* 审计历史标签页 */}
              {activeTab === 'audit-history' && selectedVersion?.recordId && (
                <>
                  {/* 调试信息 */}
                  <Box marginBottom="s" padding="s" backgroundColor="#f5f5f5" borderRadius="4px">
                    <Text typeLevel="subtext.small" color="hint">
                      🔍 调试信息: recordId = {selectedVersion.recordId}
                    </Text>
                  </Box>
                  <AuditHistorySection
                    recordId={selectedVersion.recordId}
                    params={{
                      limit: 50,
                      mode: 'current'
                    }}
                  />
                </>
              )}
              
              {activeTab === 'audit-history' && !selectedVersion?.recordId && (
                <Card padding="m">
                  <Flex alignItems="center" gap="xs" marginBottom="m">
                    <SystemIcon icon={activityStreamIcon} size={16} />
                    <Text as="h3" typeLevel="subtext.large" fontWeight="bold">
                      审计历史
                    </Text>
                  </Flex>
                  <Text typeLevel="body.medium" color="hint">
                    请选择一个版本查看对应的审计历史记录
                  </Text>
                </Card>
              )}
            </>
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
                onClick={async () => {
                  try {
                    await handleDeleteVersion(showDeleteConfirm);
                    // 成功时的处理由handleDeleteVersion内部完成
                  } catch (error) {
                    // 显示错误消息
                    const errorMessage = error instanceof Error ? error.message : '作废失败，请重试';
                    showError(errorMessage);
                  }
                }}
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
