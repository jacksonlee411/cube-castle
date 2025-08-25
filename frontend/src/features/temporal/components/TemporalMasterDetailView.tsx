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
import { Modal, useModalModel } from '@workday/canvas-kit-react/modal';
import { checkCircleIcon, exclamationCircleIcon } from '@workday/canvas-system-icons-web';
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import TemporalEditForm, { type TemporalEditFormData } from './TemporalEditForm';
import { InlineNewVersionForm } from './InlineNewVersionForm';
import { TimelineComponent, type TimelineVersion } from './TimelineComponent';
import { 
  colors, 
  borderRadius 
} from '@workday/canvas-kit-react/tokens';
import { baseColors } from '../../../shared/utils/colorTokens';
import { unifiedGraphQLClient, unifiedRESTClient } from '../../../shared/api/unified-client';

// 使用来自TimelineComponent的TimelineVersion类型
// export interface TemporalVersion 已移动到 TimelineComponent.tsx

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
      
      // 使用organizationVersions查询获取完整的版本历史 - 修复认证问题，保留健壮错误处理
      let data;
      try {
        data = await unifiedGraphQLClient.request<{
          organizationVersions: any[];
        }>(`
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
        `, {
          code: organizationCode
        });
      } catch (graphqlError: any) {
        // 保留GraphQL层面错误处理 - 符合健壮方案原则
        if (graphqlError?.response?.status) {
          const statusCode = graphqlError.response.status;
          const statusText = graphqlError.response.statusText || 'Unknown Error';
          throw new Error(`服务器响应错误 (${statusCode}): ${statusText}`);
        }
        throw new Error(`GraphQL调用失败: ${graphqlError.message || '未知错误'}`);
      }
        
      // 保留数据验证 - 防御性编程
      if (!data) {
        throw new Error('GraphQL响应为空');
      }
      
      const versions = data.organizationVersions || [];
        
        // 映射到组件需要的数据格式
        const mappedVersions: TimelineVersion[] = versions.map((version: any) => ({
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
          lifecycleStatus: version.isCurrent ? 'CURRENT' as const : 'HISTORICAL' as const,
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
        
        // 显示成功消息
        if (isRetry) {
          setSuccessMessage('数据加载成功！');
          setTimeout(() => setSuccessMessage(null), 3000);
        }
        
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
      await unifiedRESTClient.request(
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
      );
      
      // unifiedRESTClient成功时直接返回数据，失败时抛出异常
      // 刷新数据
      await loadVersions();
      setShowDeleteConfirm(null);
      
      // 如果作废的是选中的版本，重新选择
      if (selectedVersion?.effectiveDate === version.effectiveDate) {
        setSelectedVersion(null);
      }
    } catch (error) {
      console.error('Error deactivating version:', error);
      showError('作废失败，请检查网络连接');
    } finally {
      setIsDeleting(false);
    }
  }, [organizationCode, selectedVersion, isDeleting, loadVersions]);

  // 时间轴版本选择处理 - 增强功能，支持编辑历史记录页面联动
  const handleVersionSelect = useCallback((version: TimelineVersion) => {
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
        
        // 修复：使用统一认证客户端替代直接fetch调用
        const result: any = await unifiedRESTClient.request('/organization-units', {
          method: 'POST',
          body: JSON.stringify(requestBody)
        });
        
        console.log('创建成功响应:', result);
        const newOrganizationCode = result.code || result.organization?.code;
        
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
        // 更新现有组织的时态版本 - 修复：使用统一认证客户端
        await unifiedRESTClient.request(
          `/organization-units/${organizationCode}/events`,
          {
            method: 'POST',
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
        
        // unifiedRESTClient成功时直接返回数据，失败时抛出异常
        // 刷新数据
        await loadVersions();
        setActiveTab('edit-history'); // 创建成功后切换回历史记录选项卡
        showSuccess('时态版本创建成功！');
      }
    } catch (error) {
      console.error(isCreateMode ? '创建组织失败:' : '创建时态版本失败:', error);
      showError(isCreateMode ? '创建失败，请检查网络连接' : '创建失败，请检查网络连接');
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
  // const handleEditHistory = useCallback((version: TimelineVersion) => { // TODO: 暂时未使用
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
            parentCode: updateData.parentCode,
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
              // versions相关props已移除 - 违反原则13
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